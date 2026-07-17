# Yahoo Circuit Groups Implementation Plan

> `For agentic workers:` REQUIRED SUB-SKILL: use `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

`Goal:` Add explicit Yahoo endpoint-family circuit groups so protected auth failures cannot open healthy chart, timeseries, news, or options circuits on the same authority.

`Architecture:` Add a context-only circuit-group value to `utils/httpx`, key the existing registry by `{authority, group}`, and let `svc/yahoo` / `svc/scrape` annotate requests with fixed bounded family names. Ungrouped callers preserve the current host-level key and all retry, outcome, rate, cookie, and configuration semantics remain unchanged.

`Tech Stack:` Go 1.26, stdlib `context` / `net/http`, Testify, `httptest`.

## Global Constraints

- Preserve `cmd -> facade -> svc -> model`; `cmd/*` must not import `svc/*`.
- Do not add dependencies, YAML fields, or `httpx.Config` options.
- Do not encode symbol, module, or full URL in a breaker key.
- Do not change thresholds, rolling windows, retry, backoff, rate limiting, cookie jar, or outcome classification.
- Do not infer Yahoo families inside `utils/httpx`; services annotate them explicitly.
- Preserve host-level behavior for ungrouped requests.
- Follow RED -> GREEN for every production behavior change.
- Do not commit without explicit user authorization; use diff/test checkpoints.

## File Structure

| File | Responsibility |
| --- | --- |
| `utils/httpx/circuit_group.go` | Context API and group normalization only |
| `utils/httpx/circuit_registry.go` | `{authority, group}` registry identity |
| `utils/httpx/client.go` | Select the request breaker before retries |
| `svc/yahoo/circuit_groups.go` | Fixed Yahoo family names and context wrapper |
| Yahoo endpoint files | Select one family before request construction |
| `svc/scrape/circuit_group.go` | Fixed Yahoo web family wrapper |
| Active docs | Describe isolation and formal live evidence |

---

### Task 1: Context Contract and Registry Identity

`Files:`

- Create: `utils/httpx/circuit_group.go`
- Create: `utils/httpx/circuit_group_test.go`
- Modify: `utils/httpx/circuit_registry.go:9-13,43-46,65-76`
- Modify: `utils/httpx/circuit_registry_test.go:1-53`

`Interfaces:`

- Produces: `WithCircuitGroup(ctx context.Context, group string) context.Context`.
- Produces: `circuitGroupFromContext(ctx context.Context) string`.
- Produces: `(*circuitBreakerRegistry).forRequest(authority, group string) *CircuitBreaker`.
- Preserves: `forHost(host)` as the empty-group adapter.

- [x] `Step 1: Write context normalization tests`

Create `utils/httpx/circuit_group_test.go`:

```go
package httpx

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestCircuitGroupContextNormalizesValue(t *testing.T) {
    ctx := WithCircuitGroup(context.Background(), "  Yahoo-Auth  ")
    assert.Equal(t, "yahoo-auth", circuitGroupFromContext(ctx))
}

func TestCircuitGroupContextIgnoresBlankValue(t *testing.T) {
    base := context.Background()
    ctx := WithCircuitGroup(base, "   ")
    assert.Equal(t, "", circuitGroupFromContext(ctx))
    assert.Equal(t, base, ctx)
}
```

- [x] `Step 2: Write registry group-isolation tests`

Append to `utils/httpx/circuit_registry_test.go`:

```go
func TestCircuitBreakerRegistryReusesNormalizedGroup(t *testing.T) {
    registry := newCircuitBreakerRegistry(&Config{
        CircuitWindow: time.Minute, FailureThreshold: 1, ResetTimeout: time.Minute,
    })
    assert.Same(t,
        registry.forRequest("QUERY2.finance.yahoo.com", " Yahoo-Auth "),
        registry.forRequest("query2.finance.yahoo.com", "yahoo-auth"),
    )
}

func TestCircuitBreakerRegistryIsolatesGroupsOnSameHost(t *testing.T) {
    registry := newCircuitBreakerRegistry(&Config{
        CircuitWindow: time.Minute, FailureThreshold: 1, ResetTimeout: time.Minute,
    })
    auth := registry.forRequest("query2.finance.yahoo.com", "yahoo-auth")
    chart := registry.forRequest("query2.finance.yahoo.com", "yahoo-chart")
    ungrouped := registry.forHost("query2.finance.yahoo.com")

    require.NotSame(t, auth, chart)
    require.NotSame(t, auth, ungrouped)
    assert.Same(t, ungrouped, registry.forRequest("query2.finance.yahoo.com", ""))
    auth.RecordFailure()
    assert.Equal(t, StateOpen, auth.State())
    assert.Equal(t, StateClosed, chart.State())
    assert.Equal(t, StateClosed, ungrouped.State())
}
```

- [x] `Step 3: Run focused tests and verify RED`

```bash
go test ./utils/httpx -run 'TestCircuit(GroupContext|BreakerRegistry.*Group)' -count=1
```

Expected: compile failure for the three missing APIs.

- [x] `Step 4: Implement the context-only group API`

Create `utils/httpx/circuit_group.go`:

```go
// circuit_group.go — request-context contract for bounded logical circuit
// groups. Values never leave the process as HTTP headers.
package httpx

import (
    "context"
    "strings"
)

type circuitGroupContextKey struct{}

func WithCircuitGroup(ctx context.Context, group string) context.Context {
    group = normalizeCircuitGroup(group)
    if group == "" {
        return ctx
    }
    return context.WithValue(ctx, circuitGroupContextKey{}, group)
}

func circuitGroupFromContext(ctx context.Context) string {
    group, _ := ctx.Value(circuitGroupContextKey{}).(string)
    return normalizeCircuitGroup(group)
}

func normalizeCircuitGroup(group string) string {
    return strings.ToLower(strings.TrimSpace(group))
}
```

- [x] `Step 5: Change the registry to a struct key`

Replace the registry key/map and lookup with:

```go
type circuitBreakerKey struct {
    authority string
    group     string
}

type circuitBreakerRegistry struct {
    mu       sync.Mutex
    breakers map[circuitBreakerKey]*CircuitBreaker
    new      func() *CircuitBreaker
}

func (r *circuitBreakerRegistry) forRequest(authority, group string) *CircuitBreaker {
    key := circuitBreakerKey{
        authority: strings.ToLower(strings.TrimSpace(authority)),
        group:     normalizeCircuitGroup(group),
    }
    r.mu.Lock()
    defer r.mu.Unlock()
    breaker, ok := r.breakers[key]
    if !ok {
        breaker = r.new()
        r.breakers[key] = breaker
    }
    return breaker
}

func (r *circuitBreakerRegistry) forHost(host string) *CircuitBreaker {
    return r.forRequest(host, "")
}
```

Change construction to `breakers: make(map[circuitBreakerKey]*CircuitBreaker)`.

- [x] `Step 6: Run registry/context tests and verify GREEN`

```bash
gofmt -w utils/httpx/circuit_group.go utils/httpx/circuit_group_test.go utils/httpx/circuit_registry.go utils/httpx/circuit_registry_test.go
go test ./utils/httpx -run 'TestCircuit(GroupContext|BreakerRegistry)' -count=1
```

Expected: new tests and existing host/config snapshot tests pass.

- [x] `Step 7: Review checkpoint`

```bash
git diff --check
git diff -- utils/httpx/circuit_group.go utils/httpx/circuit_group_test.go utils/httpx/circuit_registry.go utils/httpx/circuit_registry_test.go
```

---

### Task 2: Client.Do Selects the Request Group

`Files:`

- Modify: `utils/httpx/client.go:111-145`
- Modify: `utils/httpx/client_test.go:183-207`

`Interfaces:`

- Consumes: `circuitGroupFromContext(req.Context())` and `forRequest`.
- Preserves: one outcome per logical `Do` and every retry/classification rule.

- [x] `Step 1: Write the failing same-host isolation test`

Append to `utils/httpx/client_test.go`:

```go
func TestClientCircuitBreakerIsScopedByRequestGroup(t *testing.T) {
    var authHits, chartHits int
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/auth":
            authHits++
            w.WriteHeader(http.StatusTooManyRequests)
        case "/chart":
            chartHits++
            w.WriteHeader(http.StatusOK)
        default:
            http.NotFound(w, r)
        }
    }))
    defer server.Close()

    config := DefaultConfig()
    config.MaxAttempts = 1
    config.FailureThreshold = 1
    config.QPS = 100
    config.Burst = 10
    client := NewClient(config)

    authCtx := WithCircuitGroup(context.Background(), "yahoo-auth")
    authReq, err := http.NewRequestWithContext(authCtx, http.MethodGet, server.URL+"/auth", nil)
    require.NoError(t, err)
    _, err = client.Do(context.Background(), authReq)
    require.Error(t, err)

    secondAuthReq, err := http.NewRequestWithContext(authCtx, http.MethodGet, server.URL+"/auth", nil)
    require.NoError(t, err)
    _, err = client.Do(context.Background(), secondAuthReq)
    require.ErrorIs(t, err, ErrCircuitOpen)

    chartCtx := WithCircuitGroup(context.Background(), "yahoo-chart")
    chartReq, err := http.NewRequestWithContext(chartCtx, http.MethodGet, server.URL+"/chart", nil)
    require.NoError(t, err)
    resp, err := client.Do(context.Background(), chartReq)
    require.NoError(t, err)
    require.NoError(t, resp.Body.Close())
    assert.Equal(t, 1, authHits)
    assert.Equal(t, 1, chartHits)
}
```

The ungrouped `Do` argument intentionally proves `req.Context()` is the source.

- [x] `Step 2: Run the focused test and verify RED`

```bash
go test ./utils/httpx -run TestClientCircuitBreakerIsScopedByRequestGroup -count=1
```

Expected: chart receives `ErrCircuitOpen` under the current host-only selection.

- [x] `Step 3: Select breaker from request context`

Replace the host-only selection in `Client.Do` with:

```go
// Circuit state is isolated by upstream authority and optional logical
// request group. The request context is the group source.
group := circuitGroupFromContext(req.Context())
breaker := c.circuitBreakers.forRequest(req.URL.Host, group)
```

Do not change the outcome defer, retry loop, response classification, or limiter.

- [x] `Step 4: Run focused and complete httpx tests`

```bash
go test ./utils/httpx -run 'TestClient(CircuitBreaker|RetryFailureRecordsOneBreakerOutcome|HTTP404IsBreakerSuccess|CancellationIsBreakerNeutral)' -count=1
go test ./utils/httpx -count=1
```

Expected: all tests pass; existing ungrouped tests continue through `forHost`.

- [x] `Step 5: Review checkpoint`

```bash
git diff --check
git diff -- utils/httpx/client.go utils/httpx/client_test.go
```

---

### Task 3: Annotate Yahoo Endpoint Families

`Files:`

- Create: `svc/yahoo/circuit_groups.go`
- Create: `svc/yahoo/circuit_groups_test.go`
- Modify: `svc/yahoo/auth.go:56-91`
- Modify: `svc/yahoo/quotesummary.go:27-55`
- Modify: `svc/yahoo/client.go:43-68,121-140,227-324`
- Modify: `svc/yahoo/timeseries.go:107-142`
- Modify: `svc/yahoo/options.go:52-62`
- Modify: `svc/yahoo/news.go:51-79`
- Modify: `svc/yahoo/earnings_dates.go:128-138`

`Interfaces:`

- Consumes: `httpx.WithCircuitGroup`.
- Produces: `circuitGroupAuth`, `circuitGroupChart`, `circuitGroupTimeseries`, `circuitGroupOptions`, `circuitGroupNews`, `circuitGroupWeb`.
- Produces: `circuitContext(ctx context.Context, group string) context.Context`.
- Preserves: endpoint URLs, methods, queries, bodies, decoders, and cookies.

- [x] `Step 1: Write the failing Yahoo isolation and source-contract tests`

Create `svc/yahoo/circuit_groups_test.go`:

```go
package yahoo

import (
    "context"
    "errors"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "testing"

    "github.com/bizshuk/yfin/utils/httpx"
    "github.com/stretchr/testify/require"
)

func TestYahooAuthCircuitDoesNotBlockChart(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch {
        case r.URL.Path == "/":
            w.WriteHeader(http.StatusOK)
        case r.URL.Path == "/v1/test/getcrumb":
            w.WriteHeader(http.StatusTooManyRequests)
        case strings.HasPrefix(r.URL.Path, "/v8/finance/chart/AAPL"):
            _, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"symbol":"AAPL","currency":"USD"}}],"error":null}}`))
        default:
            http.NotFound(w, r)
        }
    }))
    defer server.Close()

    config := httpx.DefaultConfig()
    config.MaxAttempts = 1
    config.FailureThreshold = 1
    config.QPS = 100
    config.Burst = 10
    transport := httpx.NewClient(config)
    client := NewClientWithAuth(transport, server.URL,
        NewCrumbManager(transport, server.URL, server.URL))

    _, err := client.FetchQuoteSummary(context.Background(), "AAPL", []string{"assetProfile"})
    require.Error(t, err)
    metadata, err := client.FetchMetadata(context.Background(), "AAPL")
    require.NoError(t, err)
    require.Equal(t, "AAPL", metadata.Symbol)
    _, err = client.FetchQuoteSummary(context.Background(), "AAPL", []string{"assetProfile"})
    require.True(t, errors.Is(err, httpx.ErrCircuitOpen), "auth family must remain open: %v", err)
}

func TestYahooRequestSitesDeclareCircuitGroups(t *testing.T) {
    expected := map[string]map[string]int{
        "auth.go":           {"circuitGroupAuth": 2},
        "quotesummary.go":   {"circuitGroupAuth": 1},
        "client.go":         {"circuitGroupChart": 5},
        "timeseries.go":     {"circuitGroupTimeseries": 1},
        "options.go":        {"circuitGroupOptions": 1},
        "news.go":           {"circuitGroupNews": 1},
        "earnings_dates.go": {"circuitGroupWeb": 1},
    }
    for file, groups := range expected {
        source, err := os.ReadFile(file)
        require.NoError(t, err)
        for group, count := range groups {
            require.GreaterOrEqual(t, strings.Count(string(source), group), count,
                "%s must declare %s", file, group)
        }
    }
}
```

- [x] `Step 2: Run focused tests and verify RED`

```bash
go test ./svc/yahoo -run 'TestYahoo(AuthCircuitDoesNotBlockChart|RequestSitesDeclareCircuitGroups)' -count=1
```

Expected: missing constants/helper or host breaker blocks metadata.

- [x] `Step 3: Create the fixed family contract`

Create `svc/yahoo/circuit_groups.go`:

```go
// circuit_groups.go — fixed bounded breaker-family names for Yahoo request
// surfaces. URL and response handling remain in endpoint files.
package yahoo

import (
    "context"

    "github.com/bizshuk/yfin/utils/httpx"
)

const (
    circuitGroupAuth       = "yahoo-auth"
    circuitGroupChart      = "yahoo-chart"
    circuitGroupTimeseries = "yahoo-timeseries"
    circuitGroupOptions    = "yahoo-options"
    circuitGroupNews       = "yahoo-news"
    circuitGroupWeb        = "yahoo-web"
)

func circuitContext(ctx context.Context, group string) context.Context {
    return httpx.WithCircuitGroup(ctx, group)
}
```

- [x] `Step 4: Annotate auth, quoteSummary, and chart requests`

Immediately before request construction in `bootstrapCookie`, `fetchCrumb`, and `doQuoteSummary`, add:

```go
ctx = circuitContext(ctx, circuitGroupAuth)
```

In these five `svc/yahoo/client.go` methods, add before request construction:

```go
ctx = circuitContext(ctx, circuitGroupChart)
```

The methods are `FetchDailyBars`, `fetchChartRaw`, `FetchIntradayBars`, `FetchWeeklyBars`, and `FetchMonthlyBars`. Keep passing that `ctx` to both `http.NewRequestWithContext` and `httpClient.Do`.

- [x] `Step 5: Annotate remaining Yahoo requests`

Add the exact assignment immediately before request construction:

```go
// svc/yahoo/timeseries.go
ctx = circuitContext(ctx, circuitGroupTimeseries)

// svc/yahoo/options.go
ctx = circuitContext(ctx, circuitGroupOptions)

// svc/yahoo/news.go
ctx = circuitContext(ctx, circuitGroupNews)

// svc/yahoo/earnings_dates.go
ctx = circuitContext(ctx, circuitGroupWeb)
```

Leave `svc/yahoo/isin.go` ungrouped because it uses a distinct external authority.

- [x] `Step 6: Run focused and complete Yahoo tests`

```bash
gofmt -w svc/yahoo/circuit_groups.go svc/yahoo/circuit_groups_test.go svc/yahoo/auth.go svc/yahoo/quotesummary.go svc/yahoo/client.go svc/yahoo/timeseries.go svc/yahoo/options.go svc/yahoo/news.go svc/yahoo/earnings_dates.go
go test ./svc/yahoo -run 'TestYahoo(AuthCircuitDoesNotBlockChart|RequestSitesDeclareCircuitGroups)' -count=1
go test ./svc/yahoo -count=1
```

Expected: auth stays open, metadata succeeds, every request site declares a family, and existing URL/decode tests pass.

- [x] `Step 7: Review checkpoint`

```bash
git diff --check
git diff -- svc/yahoo
```

Expected: endpoint production diffs contain one context assignment per request constructor plus the helper.

---

### Task 4: Annotate Yahoo HTML Scrape Requests

`Files:`

- Create: `svc/scrape/circuit_group.go`
- Modify: `svc/scrape/client.go:72-101`
- Modify: `svc/scrape/client_test.go:1-131`

`Interfaces:`

- Consumes: `httpx.WithCircuitGroup`.
- Produces: `yahooWebCircuitContext(ctx)`, fixed to `yahoo-web`.
- Preserves: `Caller`, robots policy, absolute URL routing, FetchMeta, and parsers.

- [x] `Step 1: Write the failing scrape group test`

Append to `svc/scrape/client_test.go` and add `net/http` / `net/http/httptest` imports:

```go
func TestFetchUsesYahooWebCircuitGroup(t *testing.T) {
    hits := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        hits++
        w.WriteHeader(http.StatusServiceUnavailable)
    }))
    defer server.Close()

    config := DefaultConfig()
    config.RobotsPolicy = string(RobotsIgnore)
    config.HTTP.MaxAttempts = 1
    config.HTTP.FailureThreshold = 1
    config.HTTP.FailureRateThreshold = 0
    config.HTTP.QPS = 100
    config.HTTP.Burst = 10
    transport := httpx.NewClient(config.HTTP)
    client, err := NewClientWithCaller(transport, config)
    require.NoError(t, err)

    _, _, err = client.Fetch(context.Background(), server.URL+"/quote/AAPL/analysis")
    require.Error(t, err)
    grouped := httpx.WithCircuitGroup(context.Background(), "yahoo-web")
    req, err := http.NewRequestWithContext(grouped, http.MethodGet,
        server.URL+"/quote/AAPL/financials", nil)
    require.NoError(t, err)
    _, err = transport.Do(grouped, req)
    require.ErrorIs(t, err, httpx.ErrCircuitOpen)
    assert.Equal(t, 1, hits, "second yahoo-web request must be rejected before transport")
}
```

- [x] `Step 2: Run the focused test and verify RED`

```bash
go test ./svc/scrape -run TestFetchUsesYahooWebCircuitGroup -count=1
```

Expected: `Fetch` opens the ungrouped breaker, so the grouped follow-up reaches the server and `hits` becomes 2.

- [x] `Step 3: Add the scrape group wrapper`

Create `svc/scrape/circuit_group.go`:

```go
// circuit_group.go — fixed breaker group for Yahoo HTML page fetches.
package scrape

import (
    "context"

    "github.com/bizshuk/yfin/utils/httpx"
)

const circuitGroupYahooWeb = "yahoo-web"

func yahooWebCircuitContext(ctx context.Context) context.Context {
    return httpx.WithCircuitGroup(ctx, circuitGroupYahooWeb)
}
```

- [x] `Step 4: Annotate Fetch after robots approval`

Immediately before `c.caller.Get` in `svc/scrape/client.go`, use:

```go
ctx = yahooWebCircuitContext(ctx)
body, meta, err := c.caller.Get(ctx, u.String(), u.Query())
```

Do not wrap robots.txt; it uses its own stdlib client outside `httpx`.

- [x] `Step 5: Run focused and complete scrape tests`

```bash
gofmt -w svc/scrape/circuit_group.go svc/scrape/client.go svc/scrape/client_test.go
go test ./svc/scrape -run 'TestFetch' -count=1
go test ./svc/scrape -count=1
```

Expected: group test and complete package pass; robots and absolute-target behavior remain unchanged.

- [x] `Step 6: Review checkpoint`

```bash
git diff --check
git diff -- svc/scrape/circuit_group.go svc/scrape/client.go svc/scrape/client_test.go
```

---

### Task 5: Documentation and Formal Acceptance

`Files:`

- Modify: `CLAUDE.md`
- Modify: `README.md`
- Modify: `docs/operations/error-handling.md`
- Modify: `plans/2026-07-17-yahoo-circuit-groups.md` checkbox/result sections

`Interfaces:`

- Consumes: Tasks 1-4.
- Produces: active docs and formal-config live evidence.
- Preserves: config values, CLI manifest, artifact schema, and dependency direction.

- [x] `Step 1: Update active architecture claims`

Use this `CLAUDE.md` responsibility text:

```markdown
- `utils/httpx/`: Resilient HTTP client — QPS rate limiting, exponential backoff, retry logic, and rolling-window circuit breakers. Breakers are keyed by authority plus an optional context-only logical group; ungrouped callers retain host-level behavior. Each logical request records one final outcome after retries.
```

Update `README.md` to mention explicit Yahoo endpoint-family isolation without claiming protected endpoints succeed. Add this operational rule to `docs/operations/error-handling.md`:

```markdown
Yahoo requests use bounded circuit groups (`yahoo-auth`, `yahoo-chart`, `yahoo-timeseries`, `yahoo-options`, `yahoo-news`, `yahoo-web`) under each authority. An auth or scrape circuit may open without rejecting a healthy chart or news request on the same host. `429` and `5xx` remain breaker failures inside their own family.
```

- [x] `Step 2: Run the consistency scan`

```bash
rg -n 'per-host|host-level|circuit breaker|circuit_open|yahoo-auth|yahoo-chart|minimum_requests' README.md CLAUDE.md docs utils/httpx svc/yahoo svc/scrape
```

Expected: no active claim says every request on one Yahoo host shares a breaker; historical specs may describe the prior state explicitly.

- [x] `Step 3: Run the complete deterministic test gate`

```bash
go test ./... -count=1
```

Expected: every package passes.

- [x] `Step 4: Run race, build, architecture, and whitespace gates`

```bash
go test -race ./utils/httpx ./svc/yahoo ./svc/scrape ./facade ./cmd/dispatch -count=1
go build -o /tmp/yfin-circuit-groups-check .
```

Then run:

```bash
if go list -f '{{.ImportPath}} {{join .Imports " "}}' ./cmd/... | grep 'github.com/bizshuk/yfin/svc/'; then
    exit 1
fi
git diff --check
```

Expected: race packages and build pass, architecture grep emits nothing, diff check passes.

- [x] `Step 5: Confirm formal configuration`

```bash
test "$(sed -n 's/^  minimum_requests: //p' config/effective.yaml)" = "10"
```

Expected: exit 0. Do not alter `minimum_requests` for acceptance.

- [x] `Step 6: Run the formal AAPL live gate`

```bash
go run . --config config/effective.yaml --retry-max 1 --timeout 12s batch --ticker AAPL --force
```

Expected aggregate: `Done. success=8 skipped=0 failed=22 not_found=0`. Non-zero process exit remains expected because the known 22 commands fail.

- [x] `Step 7: Verify healthy-family artifacts`

Determine the UTC artifact date from the run. Run `jq -e .` against these eight non-empty files:

```text
history/AAPL.<UTC-DATE>.json
actions/AAPL.<UTC-DATE>.json
income/AAPL.<UTC-DATE>.json
balance/AAPL.<UTC-DATE>.json
cashflow/AAPL.<UTC-DATE>.json
news/AAPL.<UTC-DATE>.json
isin/AAPL.<UTC-DATE>.json
metadata/AAPL.<UTC-DATE>.json
```

Verify metadata mtime belongs to this run. Inspect `_failed/AAPL.*.err`: circuit-open is allowed only for auth or web commands; it must not appear for history, actions, income, balance, cashflow, news, options, or metadata.

- [x] `Step 8: Record actual live evidence`

Append to this plan the exact aggregate, eight healthy artifacts, family-local circuit-open errors, and date. If Yahoo drift changes the aggregate, classify it without changing thresholds or expanding scope.

`2026-07-17 live evidence:`

- Formal config gate: `success=8 skipped=0 failed=22 not_found=0`; exit `1` is the expected batch failure status.
- Valid artifacts written during the run: `history`, `actions`, `income`, `balance`, `cashflow`, `news`, `isin`, and `metadata` at `AAPL.2026-07-17.json`.
- Circuit-open remained local to the auth family: `calendar`, `sec-filings`, and `sustainability`.
- No circuit-open occurred for `history`, `actions`, `income`, `balance`, `cashflow`, `news`, `options`, or `metadata`; `options` reached Yahoo and returned its own `HTTP 401: Unauthorized`.
- Remaining failures were 13 auth commands (`429` or auth-family circuit-open), 7 Yahoo web analysis commands (`503`), `options` (`401`), and `earnings-dates` (no rows).

- [x] `Step 9: Final review checkpoint`

```bash
git status --short
git diff --stat
git diff --check
```

Expected: only approved routing, breaker-group, tests, spec/plan, and active docs are changed; no commit exists.
