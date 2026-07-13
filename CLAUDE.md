# CLAUDE.md - Technical Context

## 🚀 Commands

### Build & Run

- **Build CLI**: `make build` or `go build -o yfin .`
- **Install CLI**: `go install .`
- **Run CLI**: `./yfin --help` or `go run .`

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
    - `svc/twse/`: Taiwan Stock Exchange open data client + parsers (23 endpoints). Owns HTTP/fetch logic + `Registry`/`Dispatcher`/`Fetcher`/`Endpoint` meta; the 23 `*Response` envelope structs and `*Row` typed rows are re-exported as `model.*` aliases for cross-layer reuse.
- `model/`: Pure data types + normalization logic — the **lowest layer** of the dependency graph. Houses: 8 facade-aligned SDK DTOs (`Bar`/`BarBatch`/`Quote`/`MarketData`/`CompanyInfo`/`FundamentalsSnapshot`/`FundamentalsLine`/`NewsItem`), `ScaledDecimal` precision type + helpers, `Security` + `InferMIC` + `CreateSecurity` + `ExchangeToMIC`, 9 `Normalized*` types + 4 `Converted*` + `ConvertTo` methods + `FXConverter`/`FXMeta`/`MockFXConverter`, all `Normalize*` functions (`NormalizeBars`/`Quote`/`Fundamentals`/`MarketData`/`CompanyInfo`/`Holders`/`Insider`), `ToUTCDayBoundaries`, scrape DTOs (`FetchMeta`/`ScrapeNewsItem`/`NewsStats`/value types/`Scaled` alias/`Currency` alias/`YahooNum`/`YahooInt`/`YahooString`/all comprehensive DTOs), TWSE `Response` envelope + 22 endpoint `*Response` + `*Row` types, and `scrape_convert.go` (DTO→model direct converters: `ScrapeFinancialsToSnapshot`/`BalanceSheet`/`CashFlow`/`KeyStatistics`/`Analysis`/`AnalystInsights` + `ScrapeNewsToItems`, replacing the former ampy-proto emit hop). **`svc/norm/` was merged into `model/`** in Phase 2; `svc/norm/` no longer exists.
- `utils/`: Shared infrastructure (`utils/*` may NOT import `svc/*`):
    - `utils/httpx/`: Resilient HTTP client — QPS rate limiting, exponential backoff, retry logic, circuit breaker. Single shared `http.Client`; session rotation removed.
    - `utils/cache/`: Refresh-frequency cache (daily/monthly/quarterly).
    - `utils/obsv/`: Structured logging (stdlib `slog`) + OpenTelemetry tracing (no-op tracer) + Prometheus metrics. Stands alone on stdlib + OTel + `prometheus/client_golang` (no external observability wrapper).
- `config/`: Top-level YAML loader (`os.ReadFile` + `yaml.Unmarshal` into `map[string]interface{}`, then env-var interpolation + struct mapping + validation). NOT under `internal/`. Owns business-specific types (HTTP/scrape/FX adapters) with no external config dependency.
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
> **Architecture contract: `cmd → facade → svc → model`** (model is the lowest layer; svc depends on model; facade depends on svc + model; cmd depends on facade + model). The yfin CLI's runtime code never calls `svc/*` directly — every fetch goes through `cmd.CreateClient()` (returns `*facade.Client`) which wraps `facade.NewClientWithConfig()`. The CLI flag overrides (`--qps` / `--retry-max` / `--timeout`) are applied in `cmd.CreateClient()` on top of the YAML HTTP settings. Cmd sub-packages (`market` / `fundamentals` / `scrape` / `dispatch` / `twse`) only interact with `svc/yahoo` / `svc/scrape` / `svc/twse` through facade's plain-SDK methods or facade-level wrapper functions (`Facade.Client.ScrapeFetch` / `Facade.TwseDispatch` / `Facade.YahooDispatch` etc.). `model/*` imports in cmd ARE allowed (DTOs as formatter parameters). Scrape DTOs convert to model directly via `model.Scrape*ToSnapshot` / `model.ScrapeNewsToItems` (in `model/scrape_convert.go`) — there is no ampy-proto emit hop. External packages (stock, data) call `facade.NewClient()` directly or import `model/` for raw types.

> **Why yahoo raw shapes are now in `model/`** — yahoo raw-shape structs (`ChartBar`/`RawQuote`/`ChartMeta`/`Fundamentals`/`IncomeStatement`/etc.) live in `model/yahoo_raw.go`. svc/yahoo imports `model/` (one-way only), decoders return `*model.ChartResponse`/`*model.QuoteResponse`/`*model.FundamentalsResponse` directly, and `yahoo.X` aliases in `svc/yahoo/*.go` keep callers compiling unchanged. There is no cycle because `model/` does not import `svc/yahoo`.

---

## 🔧 Code Conventions & Decisions

1. **No Session Rotation**: Session rotation was removed to simplify HTTP connection reuse and state management. The HTTP client relies on a single shared `http.Client` with robust rate-limiting, retries, and circuit breakers. `--sessions` persistent flag remains for backward compatibility but its effect is vestigial.
2. **Facade boundary (legacy)**: External projects (stock, data, ...) may import from `facade/` for the legacy back-compat surface OR from `model/` directly for raw types. The recommended path is `model/` since `facade/` types are now just aliases. **Avoid `ScaledDecimal` in external surfaces** — use float64 via `model.Bar`/`model.Quote`/etc.
3. **Decimals**: Prices/amounts are internally stored as `ScaledDecimal` for precision in `model.Normalized*` types. External-facing `model.Bar`/`model.Quote`/etc. expose float64. Conversion between the two is `model.FromScaledDecimal(sd)`.
4. **Timezones**: All timestamps must be handled in UTC. `model.ToUTCDayBoundaries()` is the canonical epoch→day conversion.
5. **`cmd → facade → svc → model` strict path**: The yfin CLI's runtime code (`cmd/**/*.go`) MUST NOT call any `svc/*` package directly. All data fetching goes through `cmd.CreateClient() → *facade.Client`; all dispatch (YahooDispatch / ScrapeFetch / TwseDispatch / Parse*) goes through facade-level functions in `facade/`. `model/*` imports in cmd ARE allowed (DTOs as formatter parameters; scrape DTO→model conversion lives in `model/scrape_convert.go`). New code in `cmd/` that needs a `svc/*` capability should expose a facade wrapper first rather than importing `svc/*`.
6. **model/ lowest-layer rule**: `model/` does NOT import `svc/*`, `facade/`, or `cmd/`. All raw-shape structs consumed by `Normalize*` (ChartBar, RawQuote, ChartMeta, Fundamentals, IncomeStatement, etc.) live in `model/yahoo_raw.go`. svc/yahoo decoders return `*model.ChartResponse`/`*model.QuoteResponse`/`*model.FundamentalsResponse` directly, so the cycle is broken.
7. **External SDK objects decode at upper layer**: When integrating a third-party SDK whose data shapes don't naturally fit `model/`, the decode/translate step happens in `facade/` (or a dedicated handler), NOT in `model/`. `model/` only owns shapes that result from such decode. The rationale: `model/` should stay import-cycle-free at the bottom of the stack; pushing decode upstream keeps the layer rule simple.

### Exit Codes (cmd/exitcodes.go)

`0=Success, 1=GeneralError, 2=PaidFeatureRequired, 3=ConfigError`.

### Persistent Flags (cmd/global.go)

`--config`, `--log-level`, `--run-id`, `--concurrency`, `--qps`, `--retry-max`, `--sessions` (vestigial), `--timeout`, `--observability-disable-tracing`, `--observability-disable-metrics`.

### Dependencies

| Dep                                                | ver         |
| -------------------------------------------------- | ----------- |
| `github.com/bizshuk/yfin` (module)                 | `go 1.26.0` |
| `github.com/bizshuk/gosdk`                         | `v1.1.0`    |
