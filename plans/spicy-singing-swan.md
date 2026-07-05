# Reshape yfinance-go: SDK-first (facade) + gosdk/cobra CLI

## Context

`yfinance-go` serves two audiences: external Go programs that embed it as an `SDK`
(the major use case — `stock`, `data`, examples) and a `simple CLI` (`yfin`).
Today those roles are tangled:

- The SDK entry `Client` lives in the repo-root `client.go` as `package yfinance`.
  A Go directory holds one package, so the root cannot also host the CLI entry
  `main.go`; the CLI lives in `cmd/yfin/` as a 3033-line `package main`.
- An in-progress refactor extracted the soak orchestrator from `internal/soak` into
  a standalone `bin/soak/`, but left `cmd/yfin/main.go:23` importing the deleted
  `internal/soak`. The tree does not build right now.
- The `facade/` package (plain reflection-free structs) exists but is unused; the
  `Client` still returns `internal/norm` + ampy-proto types, leaking internals.

Goal: split into three clean layers — `facade/` = SDK (plain structs), root
`main.go` + `cmd/` = CLI, `cmd/soak/` = soak binary — wired with gosdk
(`config`/slog/`metric`) + cobra, keeping ampy-config for domain typing.

## Target layout

```tree
/main.go                     # package main — thin entry: cmd.Execute()
/cmd/                        # package cmd — yfin subcommands (from cmd/yfin/*, minus soak)
│   root.go pull.go quote.go fundamentals.go scrape.go config.go version.go
│   comprehensive.go batch.go dispatch.go twse.go
│   fetch.go                 # NEW: svc/yahoo+norm helpers for the CLI proto/bus path
/cmd/soak/                   # package main — standalone soak binary (from bin/soak/*)
/facade/                     # SDK — plain structs (external contract)
│   client.go bars.go quote.go company_info.go
│   market_data.go fundamentals.go news.go   # NEW plain types + converters
/internal/ svc/ utils/                        # unchanged
```

## Principle

Each step below is self-contained and ends at a green `go build ./...`. Land them in
order; do not start the next until the current one's verification passes. Steps 1-5
are low risk; step 6 (the return-type flip) is the crux and comes last so everything
under it is already stable.

## Step 1 — Decouple soak, get the build green

Focus: finish the soak extraction so the module compiles again. No SDK/CLI reshape yet.

- `git mv bin/soak cmd/soak` (stays `package main`); delete the empty `bin/`.
- In `cmd/yfin/main.go`: remove the `internal/soak` import, `soakCmd`, `runSoak`,
  `SoakConfig`, and the `rootCmd.AddCommand(soakCmd)` line. Soak now lives only in
  `cmd/soak`.
- Fix `bin/soak` references in `scripts/`, `run_tests.sh`, docs.

Verify:

- `go build ./...` succeeds (currently fails on `internal/soak`).
- `go build -o soak ./cmd/soak && ./soak --help` works.
- `grep -rn "internal/soak\|bin/soak" .` is empty.

## Step 2 — Relocate `client.go` into `facade/` (rename only)

Focus: move the file and package, no behavior change. Returns stay `norm.*`/proto
for now — the plain-struct flip is Step 6.

- `git mv client.go facade/client.go`; change `package yfinance` → `package facade`;
  drop the self `yfinance` import; keep the qualified internal imports.
- Repoint the current `Client` users to `facade.Client` / `facade.NewClient*`:
  `cmd/yfin/main.go`, plus consumers `examples/standalone/api_usage/api_usage.go`,
  `examples/standalone/historical_data/...`, `examples/standalone/print_*`,
  `examples/library/scrape_fallback.go`, `scripts/fetch_nvda_data.go`.

Verify:

- `go build ./...` and `go vet ./...` green; example mains compile.

## Step 3 — Root `main.go` + `cmd` package

Focus: CLI entry relocation. Root is free of `package yfinance` after Step 2, so
`package main` can live there.

- Create `/main.go` (`package main`): `func main(){ if err := cmd.Execute(); err != nil { os.Exit(1) } }`.
- `git mv cmd/yfin/*.go cmd/`; change each file `package main` → `package cmd`.
- Rename `cmd/yfin/main.go` → `cmd/root.go`; replace `func main()` with
  `func Execute() error { return rootCmd.Execute() }`. Keep the exit-code constants
  and the leaf `RunE`s.
- Remove the empty `cmd/yfin/`.

Verify:

- `go build -o yfin . && ./yfin --help` lists subcommands (no `soak`); existing
  `cmd` `*_test.go` pass.

## Step 4 — gosdk config + logging + metric (layered)

Focus: app-level config/observability via gosdk, without touching ampy-config domain
typing.

- `go get github.com/bizshuk/gosdk` (pulls `spf13/viper`); `go mod tidy`.
- In `cmd/root.go` `rootCmd.PersistentPreRunE`:
    - `config.Default(config.WithAppName("yfin"))` — loads `~/.config/yfin` + binds
      `APP_*` env (gosdk `config/config.go:35`).
    - init slog default honoring `--log-level` (gosdk `log` sets `slog.Default`).
- In `Execute()`: `metric.CobraCMDHook(rootCmd)` once (pattern:
  `gosdk/cmd/cobrasample/main.go`).
- Keep `internal/config.NewLoader` for HTTP/bus/scrape typed structs; gosdk only adds
  app-level overrides (honors the `~/.config/<app>` convention — no custom path).

Verify:

- `go build ./...` green; `APP_LOG_LEVEL=debug ./yfin version` respects env binding.

## Step 5 — Add facade plain types + converters (additive)

Focus: new plain structs and unit-tested converters. Nothing consumes them yet, so
this cannot break callers.

- Reuse `norm.FromScaledDecimal` (`internal/norm/decimal.go:59`); mirror the
  existing `facade/bars.go` `From*` style.
- `facade/market_data.go`: `MarketData` + `FromMarketData(*norm.NormalizedMarketData)`.
- `facade/fundamentals.go`: `FundamentalsSnapshot{Symbol,MIC,Source,AsOf,Lines}`,
  `FundamentalsLine{Key,Value float64,CurrencyCode,PeriodStart,PeriodEnd}`,
  `FromFundamentalsSnapshot(*norm.NormalizedFundamentalsSnapshot)` and
  `fromProtoFundamentals(*fundamentalsv1.FundamentalsSnapshot)`.
- `facade/news.go`: `NewsItem{Title,URL,Source,PublishedAt,Summary,Symbols}`,
  `fromProtoNews([]*newsv1.NewsItem)`.

Verify:

- `go test ./facade/...` passes (add table tests for `fromProtoFundamentals` and
  `FromMarketData`).

## Step 6 — Flip facade.Client to plain structs + rewire CLI fetch path

Focus: the crux. Make the SDK return plain structs; keep the CLI on the full-precision
proto/bus path via its own helpers.

- Convert `facade.Client` return types (internals unchanged — still fetch→norm and
  scrape→emit→proto, then convert to plain via Step 5 before returning):
    - `Fetch{Daily,Intraday,Weekly,Monthly}Bars` → `*facade.BarBatch`
    - `FetchQuote` → `*facade.Quote`; `FetchCompanyInfo` → `*facade.CompanyInfo`
    - `FetchMarketData` → `*facade.MarketData`
    - `FetchFundamentalsQuarterly` + all `Scrape{Financials,BalanceSheet,CashFlow,
      KeyStatistics,Analysis,AnalystInsights}` → `*facade.FundamentalsSnapshot`
    - `ScrapeNews` → `[]facade.NewsItem`; `ScrapeAllFundamentals` →
      `[]*facade.FundamentalsSnapshot`
- Add `cmd/fetch.go`: thin `svc/yahoo` + `internal/norm` helpers (glue lifted from
  old `client.go`) — `fetchDailyBarsNorm`, `fetchQuoteNorm`, `fetchFundamentalsNorm`,
  `fetchMarketDataNorm` returning `*norm.*`. Repoint `createClient`,
  `processSymbol`, `processQuote`, `processFundamentals` to these; the
  `print*`/`handle*Bus*`/`handleLocalExport` functions keep their `norm.*` signatures.
- Update the external consumers (Step 2 list) to print plain-struct fields.

Verify:

- `go build ./...`, `go vet ./...`, `go test ./...` green.
- `./yfin quote --tickers AAPL --preview` still previews (proto path intact).
- `go run ./examples/standalone/api_usage` compiles and prints plain structs.

## Step 7 — Docs + build sync

Focus: reflect the new shape.

- `CLAUDE.md`: build commands → `go build -o yfin .` and `go build -o soak ./cmd/soak`;
  note `facade` is the SDK entry and the `cmd` / `cmd/soak` split.
- `README.md`: SDK snippet → `facade.NewClient()` + plain structs.
- Rename this plan to the workspace convention `plans/2026-07-05-sdk-cli-reshape.md`.

## Notes / risks

- `facade.Client` returning plain structs is a breaking API change (import path +
  return types) for external consumers. Acceptable pre-1.0 given the SDK-first goal.
- Keep `go.uber.org/zap` (used by `utils/obsv`); gosdk only sets `slog.Default` for
  CLI logging. Zap→slog migration is out of scope.
- A non-`main` `cmd` library with a sibling `cmd/soak` `main` is legal and idiomatic.

## Subagent dispatch prompts (steps 2–6)

When the previous step reports green, dispatch the next `general-purpose` agent
with the matching prompt below. Run them in the background in order; the harness
notifies when each finishes. Verify before starting the next step.

### Step 2 — Relocate `client.go` into `facade/`

```
You are implementing Step 2 of /Users/shuk/projects/yfinance-go/plans/spicy-singing-swan.md
in the working directory /Users/shuk/projects/yfinance-go.

Step 1 already landed (build is green, soak lives in cmd/soak). Your scope is ONLY
Step 2: rename-only relocation of client.go. Do NOT change method return types —
the plain-struct flip is Step 6. Do not start later steps.

Required actions:
1. `git mv client.go facade/client.go`
2. In facade/client.go change `package yfinance` to `package facade`. Drop any
   self-import (`github.com/bizshuk/yfin`) since facade doesn't need its own
   package alias. Keep the qualified internal imports (emit, norm, scrape, svc/yahoo,
   utils/httpx) as-is.
3. Repoint the current Client users:
   - cmd/yfin/main.go: replace `github.com/bizshuk/yfin` with
     `github.com/bizshuk/yfin/facade`; `yfinance.NewClient*` → `facade.NewClient*`;
     `*yfinance.Client` → `*facade.Client`.
   - examples/standalone/api_usage/api_usage.go
   - examples/standalone/historical_data/historical_data_example.go
   - examples/standalone/print_all_data_types/print_all_data_types.go
   - examples/standalone/print_data_contents/print_data_contents.go
   - examples/library/scrape_fallback.go
   - scripts/fetch_nvda_data.go
   For examples that have compiled artifacts next to source (api_usage, scrape_fallback,
   historical_data), leave the binary alone — recompile will happen on next `go run`.
4. Delete the repo-root `client.go` if it still exists (the git mv above already moved it).

Verification (must pass):
- `go build ./...`  — must succeed.
- `go vet ./...` — must be clean.
- `grep -rn "\"github.com/bizshuk/yfin\"" --include=*.go .` — must be empty
  (the facade path is allowed; only the bare root import must be gone).
- `go run ./examples/standalone/api_usage` — must compile (network call may fail
  offline; check `go build ./examples/standalone/api_usage` instead if `go run`
  errors on a 401).

Reporting: short summary + the exact verification command output for each check.
Do not commit. Do not touch the plan file.
```

### Step 3 — Root `main.go` + `cmd` package

```
You are implementing Step 3 of /Users/shuk/projects/yfinance-go/plans/spicy-singing-swan.md
in the working directory /Users/shuk/projects/yfinance-go.

Steps 1–2 already landed. Root no longer hosts `package yfinance` (client.go moved
to facade/), so root can become `package main`. Your scope is ONLY Step 3: relocate
the CLI entry, don't add gosdk yet (that's Step 4). Do not start later steps.

Required actions:
1. Create `/main.go` (package main):
       func main() { if err := cmd.Execute(); err != nil { os.Exit(1) } }
   Import the new cmd package: `github.com/bizshuk/yfin/cmd`.
2. `git mv cmd/yfin/*.go cmd/` (pull every .go file into the new cmd/ directory;
   skip any `.gitkeep` or non-go files).
3. In every file under cmd/, change `package main` to `package cmd`.
4. Rename cmd/yfin/main.go → cmd/root.go (use git mv). Replace its `func main()`
   with:
       func Execute() error { return rootCmd.Execute() }
   Keep the `version`/`commit`/`date` vars, all exit-code constants (rename to
   `ErrSuccess`/`ErrGeneral`/etc. if needed, but keep the numeric values the same),
   and every leaf `RunE`. The leaf RunEs currently `os.Exit(...)` directly; leave
   them that way for now — we'll tidy that in a later step.
5. Remove the now-empty cmd/yfin/ directory.
6. Fix any test files under cmd/ that still reference `cmd/yfin` paths.

Verification (must pass):
- `go build -o /tmp/yfin-bin . && /tmp/yfin-bin --help` — must list subcommands
  including pull, quote, fundamentals, scrape, comprehensive-stats,
  comprehensive-profile, config, version, batch, twse. Must NOT include `soak`.
- `go test ./cmd/...` — must pass (batch_test.go, dispatch_test.go, twse_test.go).
- `grep -rn "cmd/yfin\|package main" cmd/` — must be empty (every file is package cmd).

Reporting: short summary + exact verification output.
```

### Step 4 — gosdk config + logging + metric (layered)

```
You are implementing Step 4 of /Users/shuk/projects/yfinance-go/plans/spicy-singing-swan.md
in the working directory /Users/shuk/projects/yfinance-go.

Steps 1–3 already landed. Your scope is ONLY Step 4: add gosdk wiring on top of
ampy-config. Don't replace ampy-config (internal/config) — it still owns the
domain HTTP/bus/scrape typing.

Required actions:
1. Add gosdk to go.mod: `go get github.com/bizshuk/gosdk` then `go mod tidy`.
   Use the version currently on the master branch at /Users/shuk/projects/gosdk
   (run `go list -m github.com/bizshuk/gosdk` first to see what's already there,
   prefer that pin over @latest).
2. In cmd/root.go (which is now `package cmd` and exposes `func Execute() error`),
   wire gosdk inside `rootCmd.PersistentPreRunE`:
       func(cmd *cobra.Command, args []string) error {
           config.Default(config.WithAppName("yfin"))   // ~/.config/yfin + APP_* env
           // init slog honoring --log-level if gosdk/log exposes a helper;
           // otherwise just leave slog.Default as gosdk initialized it.
           return nil
       }
   Reference APIs:
   - gosdk/config: config.Default(opts ...ConfigOption), config.WithAppName(string).
     See /Users/shuk/projects/gosdk/config/config.go:35 and config/option.go:47.
   - gosdk/log: check /Users/shuk/projects/gosdk/log/log.go and level.go for the
     actual exported function (it may be log.Init or similar — read the file).
3. In `Execute()`, after `rootCmd` is built, call `metric.CobraCMDHook(rootCmd)`
   exactly once. See /Users/shuk/projects/gosdk/cmd/cobrasample/main.go for the
   canonical pattern.
4. Keep the existing --config / --log-level / --run-id persistent flags and the
   ampy-config loader. gosdk adds app-level overrides only — don't replace.

Verification (must pass):
- `go build ./...` — must succeed.
- `APP_LOG_LEVEL=debug /tmp/yfin-bin version` — must respect env binding (the
  debug log line should appear; if gosdk doesn't surface log level through
  APP_LOG_LEVEL, document that and move on, but the build must still pass).
- `go vet ./...` — clean.

Reporting: short summary + verification output. If gosdk's exact log-init
function name differs from what you assumed, cite the file:line you found it at.
```

### Step 5 — Add facade plain types + converters (additive)

```
You are implementing Step 5 of /Users/shuk/projects/yfinance-go/plans/spicy-singing-swan.md
in the working directory /Users/shuk/projects/yfinance-go.

Steps 1–4 already landed. Your scope is ONLY Step 5: add new plain-struct types +
converters in facade/. Do NOT change facade.Client's return types — Step 6 does
that. Nothing will call the new converters yet, so this step cannot break callers.

Required actions:
1. facade/market_data.go (new file, package facade):
   - `type MarketData struct` mirroring fields of `norm.NormalizedMarketData`:
     Symbol, MIC, RegularMarketPrice/High/Low/Volume (float64/int64, *float64/*int64
     nullable), FiftyTwoWeekHigh/Low, PreviousClose, CurrencyCode, EventTime.
   - `func FromMarketData(*norm.NormalizedMarketData) *MarketData` using
     `norm.FromScaledDecimal` (internal/norm/decimal.go:59).
2. facade/fundamentals.go (new file, package facade):
   - `type FundamentalsLine struct{ Key string; Value float64; CurrencyCode string;
     PeriodStart, PeriodEnd time.Time }`
   - `type FundamentalsSnapshot struct{ Symbol, MIC, Source string; AsOf time.Time;
     Lines []FundamentalsLine }`
   - `func FromFundamentalsSnapshot(*norm.NormalizedFundamentalsSnapshot) *FundamentalsSnapshot`
   - `func fromProtoFundamentals(*fundamentalsv1.FundamentalsSnapshot) *FundamentalsSnapshot`
     (lower-case = unexported, but in same package as future callers if needed;
     ampy-proto fundamentals package:
     github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1).
3. facade/news.go (new file, package facade):
   - `type NewsItem struct{ Title, URL, Source, Summary string; PublishedAt time.Time;
     Symbols []string }`
   - `func fromProtoNews([]*newsv1.NewsItem) []NewsItem`
4. facade/fundamentals_test.go (new): table test for `fromProtoFundamentals` round-trip
   (given a populated proto, expect the same key/value/period on the struct).
5. facade/market_data_test.go (new): table test for `FromMarketData` (given a
   populated norm type with a non-nil ScaledDecimal, expect the float64 scale
   conversion).

Verification (must pass):
- `go test ./facade/...` — must pass.
- `go vet ./...` — clean.
- `go build ./...` — clean.

Reporting: short summary + the test output. If you have to choose how to map
proto `FundamentalsSnapshot.Line` (which is a `map<string, LineValue>`) onto a
slice, document your decision — flat slice loses key ordering but is the simplest.
```

### Step 6 — Flip facade.Client to plain structs + rewire CLI fetch path

```
You are implementing Step 6 of /Users/shuk/projects/yfinance-go/plans/spicy-singing-swan.md
in the working directory /Users/shuk/projects/yfinance-go.

Steps 1–5 already landed. Your scope is ONLY Step 6: the crux step that flips
facade.Client to return plain structs, and adds cmd/fetch.go so the CLI keeps
full ampy-proto/norm precision for bus publishing.

Required actions:
1. In facade/client.go, change every Fetch*/Scrape* method's return type:
   - Fetch{Daily,Intraday,Weekly,Monthly}Bars → *facade.BarBatch (via FromBarBatch)
   - FetchQuote → *facade.Quote (via FromQuote)
   - FetchCompanyInfo → *facade.CompanyInfo (via FromCompanyInfo)
   - FetchMarketData → *facade.MarketData (via FromMarketData, from Step 5)
   - FetchFundamentalsQuarterly → *facade.FundamentalsSnapshot
     (via FromFundamentalsSnapshot)
   - Scrape{Financials,BalanceSheet,CashFlow,KeyStatistics,Analysis,
     AnalystInsights} → *facade.FundamentalsSnapshot (via fromProtoFundamentals)
   - ScrapeNews → []facade.NewsItem (via fromProtoNews)
   - ScrapeAllFundamentals → []*facade.FundamentalsSnapshot
   The internals (yahooClient.Fetch*, scrape.Parse*, emit.Map*) stay unchanged;
   the conversion to plain structs is the last step before returning.
2. Create cmd/fetch.go (package cmd): thin helpers using svc/yahoo + internal/norm
   + internal/emit + internal/scrape — lifted from the old client.go internals —
   so the CLI can still get the full norm/proto values for bus publishing:
       fetchDailyBarsNorm(ctx, symbol, start, end, adjusted) (*norm.NormalizedBarBatch, error)
       fetchQuoteNorm(ctx, symbol) (*norm.NormalizedQuote, error)
       fetchFundamentalsNorm(ctx, symbol) (*norm.NormalizedFundamentalsSnapshot, error)
       fetchMarketDataNorm(ctx, symbol) (*norm.NormalizedMarketData, error)
   These don't go through facade.Client; they call svc/yahoo directly.
3. In cmd/root.go (and the other cmd/ files), repoint the call sites:
   - `createClient` currently returns *facade.Client — change to return
     a small `*cliClient` struct holding *svc/yahoo.Client, or just the *yahoo.Client.
   - processSymbol: replace `client.FetchDailyBars(...)` with
     `fetchDailyBarsNorm(...)` so the bars keep ScaledDecimal precision.
   - processQuote: same for FetchQuote → fetchQuoteNorm.
   - processFundamentals: same for FetchFundamentalsQuarterly → fetchFundamentalsNorm.
   - print*/handle*Bus*/handleLocalExport keep their norm.* signatures unchanged.
4. Update the example consumers (from Step 2 list) so their `client.Fetch*` calls
   receive *facade.BarBatch etc. and they print plain-struct fields
   (e.g. batch.Symbol / batch.Bars[i].Close instead of bar.Close.Scaled/Close.Scale).

Verification (must pass):
- `go build ./...` — must succeed.
- `go vet ./...` — clean.
- `go test ./...` — full suite (facade, cmd, svc/twse, internal/*).
- `/tmp/yfin-bin quote --tickers AAPL --preview` — must still preview
  (proto path intact; runtime may fail offline — accept a connection error, but
  the code must compile and the cobra plumbing must work).
- `go build ./examples/standalone/api_usage` — must compile.

Reporting: short summary + full test output + which call sites you had to retarget.
```
