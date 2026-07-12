# CLAUDE.md - Technical Context

## 🚀 Commands

### Build & Run

- **Build CLI**: `make build` or `go build -o yfin ./cmd/yfin`
- **Install CLI**: `go install ./cmd/yfin`
- **Run CLI**: `./yfin --help` or `go run ./cmd/yfin`

### Test & Lint

- **Unit Tests**: `make test` or `go test ./...`
- **Integration Tests**: `go test -tags=integration ./...`
- **Coverage**: `make test-coverage`
- **Lint**: `make lint` or `golangci-lint run`
- **Format**: `make fmt` or `go fmt ./...`

---

## 🗂️ Project Structure

Module path: `github.com/bizshuk/yfin` (曾用名 `yfinance-go`).

- `facade/`: Publicly exported, reflection-free plain Go structs (e.g. `facade.Bar`, `facade.Quote`) for external packages (e.g., `stock`, `data`) to avoid importing `svc/*` internals. Two constructors: `facade.NewClient()` / `facade.NewClientWithConfig(*httpx.Config)`. The `*Client` is the **single handler** for both external SDK consumers AND the in-repo `yfin` CLI — all data fetching (Fetch*/ YahooDispatch / Scrape* / TwseDispatch / Parse*) routes through it. The Norm-returning variants (`FetchDailyBarsNorm` / `FetchQuoteNorm` / `FetchFundamentalsNorm` / `FetchMarketDataNorm`) hand back `*model.Normalized\*` types for callers that need ScaledDecimal precision (cmd's emit→proto pipeline). `facade.*` DTOs are now **type aliases** for `model.*` — they exist only for back-compat shim.
- `svc/`: SDK-first business services (HTTP fetch + decode + validation only):
    - `svc/yahoo/`: Yahoo Finance raw HTTP API client (Crumb auth, chart API, fundamentals). Owns its response envelope types (BarsResponse, QuoteResponse, FundamentalsResponse, etc.); these stay in svc/yahoo because model/normalize.go consumes them and Go would create an import cycle if they moved to model/.
    - `svc/scrape/`: Web scraping engine (robots.txt compliant, fills API gaps). Owns HTTP/parse logic + scrape-config types (`Config`/`RetryConfig`/`EndpointConfig`/`RobotsPolicy`/`RegexConfig`); DTOs (ComprehensiveFinancialsDTO, KeyStatisticsDTO, etc.) are now re-exported as `model.*` aliases for cross-layer reuse.
    - `svc/emit/`: Maps normalized data to `ampy-proto` Protobuf + validates.
    - `svc/twse/`: Taiwan Stock Exchange open data client + parsers (23 endpoints). Owns HTTP/fetch logic + `Registry`/`Dispatcher`/`Fetcher`/`Endpoint` meta; the 23 `*Response` envelope structs and `*Row` typed rows are re-exported as `model.*` aliases for cross-layer reuse.
- `model/`: Pure data types + normalization logic — the **lowest layer** of the dependency graph. Houses: 8 facade-aligned SDK DTOs (`Bar`/`BarBatch`/`Quote`/`MarketData`/`CompanyInfo`/`FundamentalsSnapshot`/`FundamentalsLine`/`NewsItem`), `ScaledDecimal` precision type + helpers, `Security` + `InferMIC` + `CreateSecurity` + `ExchangeToMIC`, 9 `Normalized*` types + 4 `Converted*` + `ConvertTo` methods + `FXConverter`/`FXMeta`/`MockFXConverter`, all `Normalize*` functions (`NormalizeBars`/`Quote`/`Fundamentals`/`MarketData`/`CompanyInfo`/`Holders`/`Insider`), `ToUTCDayBoundaries`, scrape DTOs (`FetchMeta`/`ScrapeNewsItem`/`NewsStats`/value types/`Scaled` alias/`Currency` alias/`YahooNum`/`YahooInt`/`YahooString`/all comprehensive DTOs), TWSE `Response` envelope + 22 endpoint `*Response` + `*Row` types. **`svc/norm/` was merged into `model/`** in Phase 2; `svc/norm/` no longer exists.
- `utils/`: Shared infrastructure (`utils/*` may NOT import `svc/*`):
    - `utils/httpx/`: Resilient HTTP client — QPS rate limiting, exponential backoff, retry logic, circuit breaker. Single shared `http.Client`; session rotation removed.
    - `utils/bus/`: ampy-bus publisher (retry, chunking, envelope).
    - `utils/cache/`: Refresh-frequency cache (daily/monthly/quarterly).
    - `utils/obsv/`: OpenTelemetry tracing + Prometheus metrics (default-export to `inf`).
- `config/`: Top-level ampy-config YAML loader. NOT under `internal/`. Wraps `AmpyFin/ampy-config v1.1.4` with business-specific types (HTTP/bus/scrape/FX adapters).
- `cmd/`: CLI composition root. `main.go` calls each sub-package's `Register(RootCmd)`:
    - `cmd/{root,client,global,build,exitcodes}.go`: helpers + persistent flags + shared client builder (`CreateClient()` returns `*facade.Client`).
    - `cmd/admin/`: `config-effective`, `version`.
    - `cmd/dispatch/`: `batch` (sub-package name is `dispatch` but no top-level `yfin dispatch` command exists).
    - `cmd/fundamentals/`: `fundamentals`, `comprehensive-stats`, `comprehensive-profile`.
    - `cmd/market/`: `pull`, `quote`.
    - `cmd/scrape/`: `scrape` (4 mutually-exclusive modes: `--check` / `--preview-json` / `--preview-news` / `--preview-proto`).
    - `cmd/twse/`: `twse` (23 endpoints with `--endpoint` / `--date` / `--stock` / `--month` flags).
    - `cmd/soak/`: Standalone binary `soak` (invoked via `go run ./cmd/soak`, NOT `yfin soak`).
    - `cmd/samples/`, `cmd/tools/`: scripts + golden fixtures.

> CLI tree is **flat** (`yfin pull`, `yfin quote`, ...) — there are no nested groups. `cmd/market/` is the Go sub-package directory name, not a user-facing group.
>
> **Architecture contract: `cmd → facade → svc → model`** (model is the lowest layer; svc depends on model; facade depends on svc + model; cmd depends on facade + model). The yfin CLI's runtime code never calls `svc/*` directly — every fetch goes through `cmd.CreateClient()` (returns `*facade.Client`) which wraps `facade.NewClientWithConfig()`. The CLI flag overrides (`--qps` / `--retry-max` / `--timeout`) are applied in `cmd.CreateClient()` on top of the ampy-config HTTP settings. Cmd sub-packages (`market` / `fundamentals` / `scrape` / `dispatch` / `twse`) only interact with `svc/yahoo` / `svc/scrape` / `svc/twse` through facade's plain-SDK methods or facade-level wrapper functions (`Facade.Client.ScrapeFetch` / `Facade.TwseDispatch` / `Facade.YahooDispatch` etc.). `model/*` and `svc/emit` type imports in cmd ARE allowed (DTOs as formatter parameters, proto emission); `svc/emit` is invoked only for proto emission (a transport concern). External packages (stock, data) call `facade.NewClient()` directly or import `model/` for raw types.

> **Why yahoo types stay in `svc/yahoo` (NOT in `model/`)** — `model/normalize.go` consumes `yahoo.Bar`/`yahoo.Quote`/`yahoo.Fundamentals`/etc. as raw inputs. If yahoo types moved to model/, `svc/yahoo` would import model/ (for its `Decode*` return types) and `model/` would import `svc/yahoo` (for normalize inputs) — Go forbids this cycle. The asymmetry is deliberate: `model/ → svc/yahoo` is allowed; `svc/yahoo → model/` is not.

---

## 🔧 Code Conventions & Decisions

1. **No Session Rotation**: Session rotation was removed to simplify HTTP connection reuse and state management. The HTTP client relies on a single shared `http.Client` with robust rate-limiting, retries, and circuit breakers. `--sessions` persistent flag remains for backward compatibility but its effect is vestigial.
2. **Facade boundary (legacy)**: External projects (stock, data, ...) may import from `facade/` for the legacy back-compat surface OR from `model/` directly for raw types. The recommended path is `model/` since `facade/` types are now just aliases. **Avoid `ScaledDecimal` in external surfaces** — use float64 via `model.Bar`/`model.Quote`/etc.
3. **Decimals**: Prices/amounts are internally stored as `ScaledDecimal` for precision in `model.Normalized*` types. External-facing `model.Bar`/`model.Quote`/etc. expose float64. Conversion between the two is `model.FromScaledDecimal(sd)`.
4. **Timezones**: All timestamps must be handled in UTC. `model.ToUTCDayBoundaries()` is the canonical epoch→day conversion.
5. **`cmd → facade → svc → model` strict path**: The yfin CLI's runtime code (`cmd/**/*.go`) MUST NOT call any `svc/*` package directly. All data fetching goes through `cmd.CreateClient() → *facade.Client`; all dispatch (YahooDispatch / ScrapeFetch / TwseDispatch / Parse*) goes through facade-level functions in `facade/`. `model/*` and `svc/emit` imports in cmd ARE allowed (DTOs as formatter parameters, proto emission). New code in `cmd/` that needs a `svc/*` capability should expose a facade wrapper first rather than importing `svc/*`.
6. **model/ lowest-layer rule**: `model/` imports `svc/yahoo` only (for raw input shapes consumed by `Normalize*`). It does NOT import `svc/scrape`, `svc/twse`, `svc/emit`, or `facade`. If new normalize logic needs raw data not in `svc/yahoo`, define a minimal new raw-shape struct inside `model/` rather than adding new imports.

### Exit Codes (cmd/exitcodes.go)

`0=Success, 1=GeneralError, 2=PaidFeatureRequired, 3=ConfigError, 4=PublishingError`.

### Persistent Flags (cmd/global.go)

`--config`, `--log-level`, `--run-id`, `--concurrency`, `--qps`, `--retry-max`, `--sessions` (vestigial), `--timeout`, `--observability-disable-tracing`, `--observability-disable-metrics`.

### Dependencies

| Dep                                                | ver         |
| -------------------------------------------------- | ----------- |
| `github.com/bizshuk/yfin` (module)                 | `go 1.26.0` |
| `github.com/bizshuk/gosdk`                         | `v1.1.0`    |
| `github.com/AmpyFin/ampy-bus`                      | `v1.2.0`    |
| `github.com/AmpyFin/ampy-config/go/ampyconfig`     | `v1.1.4`    |
| `github.com/AmpyFin/ampy-observability/go/ampyobs` | `v0.0.3`    |
| `github.com/AmpyFin/ampy-proto/v2`                 | `v2.1.1`    |
