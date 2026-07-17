# Host-scoped Circuit Breaker Implementation Plan

> `For agentic workers:` REQUIRED SUB-SKILL: use `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

`Goal:` Replace the shared retry-amplified circuit breaker with a per-host, one-outcome-per-request, rolling-window breaker while preserving the legacy Go count-mode API.

`Architecture:` Add rate-mode configuration without removing `FailureThreshold int`; extract breaker state from `client.go` into a focused implementation, introduce a host-keyed registry, then move breaker recording outside the retry loop. The CLI uses a 30% failure rate with a 10-request minimum, while direct Go callers can continue to select absolute failure-count mode.

`Tech Stack:` Go 1.26, stdlib `net/http`, `sync`, `time`, `httptest`, Testify, YAML v3.

## Global Constraints

- Preserve `cmd -> facade -> svc -> model`; `cmd/*` must not import `svc/*`.
- Keep `httpx.Caller`, `Client.Do`, `Client.Get`, and `NewCircuitBreaker` signatures compatible.
- Do not add third-party dependencies.
- Count at most one breaker outcome per logical `Client.Do`, regardless of retry attempts.
- Key breaker state by `req.URL.Host`, including port.
- Do not change Yahoo cookie, crumb, TLS fingerprint, endpoint selection, retry count, backoff, or rate-limiter behavior.
- Do not commit without explicit user authorization; use review checkpoints instead.
- Follow RED -> GREEN for every production behavior change.

---

### Task 1: Rate-mode Configuration Contract

`Files:`

- Modify: `utils/httpx/client.go`
- Modify: `utils/httpx/config_test.go`
- Modify: `config/retry.go`
- Modify: `config/defaults.go`
- Modify: `config/effective.yaml`
- Modify: `config/adapters.go`
- Modify: `config/loader.go`
- Modify: `config/loader_test.go`
- Modify: `svc/scrape/types.go`

`Interfaces:`

- Produces: `httpx.Config.FailureRateThreshold float64` and `httpx.Config.MinimumRequests int`.
- Produces: YAML `circuit_breaker.minimum_requests`, default `10`, validation `>= 1`.
- Preserves: `httpx.Config.FailureThreshold int` as the legacy count-mode selector.

- [x] `Step 1: Write failing httpx default assertions`

Update `TestDefaultConfig` in `utils/httpx/config_test.go`:

```go
if config.FailureThreshold != 0 {
	t.Errorf("Expected legacy FailureThreshold 0, got %d", config.FailureThreshold)
}
if config.FailureRateThreshold != 0.30 {
	t.Errorf("Expected FailureRateThreshold 0.30, got %v", config.FailureRateThreshold)
}
if config.MinimumRequests != 10 {
	t.Errorf("Expected MinimumRequests 10, got %d", config.MinimumRequests)
}
```

Extend the literal in `TestConfigFields`:

```go
FailureRateThreshold: 0.40,
MinimumRequests:      20,
```

- [x] `Step 2: Write failing loader mapping and validation tests`

Extend `TestGetHTTPConfig` in `config/loader_test.go`:

```go
if httpConfig.FailureThreshold != 0 {
	t.Errorf("FailureThreshold = %d, want legacy mode disabled", httpConfig.FailureThreshold)
}
if httpConfig.FailureRateThreshold != 0.30 {
	t.Errorf("FailureRateThreshold = %v, want 0.30", httpConfig.FailureRateThreshold)
}
if httpConfig.MinimumRequests != 10 {
	t.Errorf("MinimumRequests = %d, want 10", httpConfig.MinimumRequests)
}
```

Add a validation test using existing helpers:

```go
func TestValidateCircuitMinimumRequests(t *testing.T) {
	path := "test-invalid-circuit-minimum.yaml"
	defer os.Remove(path)
	require.NoError(t, CreateEffectiveConfig(path))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := strings.Replace(string(data), "minimum_requests: 10", "minimum_requests: -1", 1)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	_, err = NewLoader(path).Load()
	require.ErrorContains(t, err, "circuit_breaker.minimum_requests must be >= 1")
}
```

- [x] `Step 3: Run focused tests and verify RED`

Run:

```bash
go test ./utils/httpx ./config -run 'Test(DefaultConfig|ConfigFields|GetHTTPConfig|ValidateCircuitMinimumRequests)$' -count=1
```

Expected: compile failure for missing rate/minimum fields or validation failure.

- [x] `Step 4: Add fields, defaults, YAML mapping, and validation`

Extend `httpx.Config` and `DefaultConfig()`:

```go
CircuitWindow        time.Duration
FailureThreshold     int
FailureRateThreshold float64
MinimumRequests      int
ResetTimeout         time.Duration
```

```go
CircuitWindow:        60 * time.Second,
FailureThreshold:     0,
FailureRateThreshold: 0.30,
MinimumRequests:      10,
ResetTimeout:         30 * time.Second,
```

Extend `config.CircuitBreakerConfig`:

```go
const defaultCircuitMinimumRequests = 10

type CircuitBreakerConfig struct {
	Window           int     `yaml:"window"`
	FailureThreshold float64 `yaml:"failure_threshold"`
	MinimumRequests  int     `yaml:"minimum_requests"`
	ResetTimeoutMs   int     `yaml:"reset_timeout_ms"`
}
```

Add `minimum_requests: 10` to `CreateEffectiveConfig` and
`config/effective.yaml`. After YAML unmarshal, apply the omitted-field default:

```go
if config.CircuitBreaker.MinimumRequests == 0 {
	config.CircuitBreaker.MinimumRequests = defaultCircuitMinimumRequests
}
```

Validate negative values:

```go
if config.CircuitBreaker.MinimumRequests < 1 {
	return fmt.Errorf("circuit_breaker.minimum_requests must be >= 1")
}
```

Change `assembleHTTPConfig` to direct rate mapping:

```go
CircuitWindow:        time.Duration(c.CircuitBreaker.Window) * time.Second,
FailureThreshold:     0,
FailureRateThreshold: c.CircuitBreaker.FailureThreshold,
MinimumRequests:      c.CircuitBreaker.MinimumRequests,
ResetTimeout:         time.Duration(c.CircuitBreaker.ResetTimeoutMs) * time.Millisecond,
```

Change both scrape default copies from `FailureThreshold: 5` to:

```go
FailureThreshold:     0,
FailureRateThreshold: 0.30,
MinimumRequests:      10,
```

- [x] `Step 5: Format and verify GREEN`

Run:

```bash
gofmt -w utils/httpx/client.go utils/httpx/config_test.go config/retry.go config/defaults.go config/adapters.go config/loader.go config/loader_test.go svc/scrape/types.go
go test ./utils/httpx ./config ./svc/scrape -count=1
```

Expected: all three packages pass.

- [x] `Step 6: Review checkpoint`

Run `git diff --check`; review only Task 1 files. Do not commit.

---

### Task 2: Rolling-window Circuit Breaker Core

`Files:`

- Create: `utils/httpx/circuit_breaker.go`
- Modify: `utils/httpx/circuit_breaker_test.go`
- Modify: `utils/httpx/client.go`

`Interfaces:`

- Preserves: `NewCircuitBreaker(window time.Duration, failureThreshold int, resetTimeout time.Duration) *CircuitBreaker`.
- Produces: `NewFailureRateCircuitBreaker(window time.Duration, failureRateThreshold float64, minimumRequests int, resetTimeout time.Duration) *CircuitBreaker`.
- Produces: `(*CircuitBreaker).record(kind circuitOutcomeKind)` and existing state methods.
- Produces: `(*CircuitBreaker).Samples() int` for tests/diagnostics.

- [x] `Step 1: Add a deterministic clock helper and failing behavior tests`

Use an injected clock instead of `time.Sleep`:

```go
func setCircuitClock(cb *CircuitBreaker, current *time.Time) {
	cb.now = func() time.Time { return *current }
}
```

Add these test scenarios in `circuit_breaker_test.go`:

```go
func TestCircuitBreakerRateModeRequiresMinimumSamples(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewFailureRateCircuitBreaker(time.Minute, 0.30, 10, 30*time.Second)
	setCircuitClock(cb, &now)
	for range 9 { cb.RecordFailure() }
	assert.Equal(t, StateClosed, cb.State())
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreakerRateModeUsesSuccessDenominator(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewFailureRateCircuitBreaker(time.Minute, 0.30, 10, 30*time.Second)
	setCircuitClock(cb, &now)
	for range 7 { cb.RecordSuccess() }
	for range 2 { cb.RecordFailure() }
	assert.Equal(t, StateClosed, cb.State())
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreakerExpiresOutcomesOutsideWindow(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewFailureRateCircuitBreaker(time.Minute, 0.50, 3, 30*time.Second)
	setCircuitClock(cb, &now)
	cb.RecordFailure()
	now = now.Add(time.Minute + time.Second)
	cb.RecordSuccess()
	cb.RecordSuccess()
	assert.Equal(t, 2, cb.Samples())
	assert.Equal(t, 0, cb.Failures())
}

func TestCircuitBreakerHalfOpenAllowsSingleProbe(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewCircuitBreaker(time.Minute, 1, 30*time.Second)
	setCircuitClock(cb, &now)
	cb.RecordFailure()
	now = now.Add(31 * time.Second)
	assert.True(t, cb.Allow())
	assert.False(t, cb.Allow())
	cb.RecordSuccess()
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreakerNeutralProbeReopens(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewCircuitBreaker(time.Minute, 1, 30*time.Second)
	setCircuitClock(cb, &now)
	cb.RecordFailure()
	now = now.Add(31 * time.Second)
	require.True(t, cb.Allow())
	cb.record(circuitOutcomeNeutral)
	assert.Equal(t, StateOpen, cb.State())
	assert.False(t, cb.Allow())
}
```

- [x] `Step 2: Run breaker tests and verify RED`

Run `go test ./utils/httpx -run 'TestCircuitBreaker' -count=1`.

Expected: missing constructor, clock, `Samples`, and neutral outcome failures.

- [x] `Step 3: Extract and implement the breaker core`

Move breaker types/methods from `client.go` into `circuit_breaker.go` and use:

```go
type circuitOutcomeKind uint8

const (
	circuitOutcomeNeutral circuitOutcomeKind = iota
	circuitOutcomeSuccess
	circuitOutcomeFailure
)

type circuitOutcome struct {
	at     time.Time
	failed bool
}

type CircuitBreaker struct {
	window               time.Duration
	failureThreshold     int
	failureRateThreshold float64
	minimumRequests      int
	resetTimeout         time.Duration
	state                 CircuitState
	outcomes              []circuitOutcome
	openedAt              time.Time
	probeInFlight         bool
	now                   func() time.Time
	mu                    sync.Mutex
}
```

Both constructors set `now: time.Now`. `pruneLocked` removes outcomes before
`now.Add(-window)`. `shouldOpenLocked` is:

```go
failures := 0
for _, outcome := range cb.outcomes {
	if outcome.failed { failures++ }
}
if cb.failureThreshold > 0 { return failures >= cb.failureThreshold }
if len(cb.outcomes) < cb.minimumRequests { return false }
return float64(failures)/float64(len(cb.outcomes)) >= cb.failureRateThreshold
```

`Allow` uses one exclusive lock. Closed prunes and permits. Open permits exactly
one request after `resetTimeout`, sets `StateHalfOpen` and `probeInFlight=true`.
Half-open rejects further requests while the probe is in flight.

`record` appends closed-state success/failure outcomes; opens when
`shouldOpenLocked` is true; half-open success closes and clears history;
half-open failure or neutral reopens and resets the probe timer. Keep
`RecordSuccess`/`RecordFailure` wrappers and implement `Samples`/`Failures` with
pruning under lock.

- [x] `Step 4: Format and verify GREEN`

Run:

```bash
gofmt -w utils/httpx/circuit_breaker.go utils/httpx/circuit_breaker_test.go utils/httpx/client.go
go test ./utils/httpx -run 'TestCircuitBreaker' -count=1
```

Expected: breaker tests pass without real-time sleeps.

- [x] `Step 5: Review checkpoint`

Confirm `client.go` no longer defines breaker state types and run
`git diff --check`. Do not commit.

---

### Task 3: Host-keyed Breaker Registry

`Files:`

- Create: `utils/httpx/circuit_registry.go`
- Create: `utils/httpx/circuit_registry_test.go`
- Modify: `utils/httpx/client.go`

`Interfaces:`

- Produces: `newCircuitBreakerRegistry(config *Config) *circuitBreakerRegistry`.
- Produces: `(*circuitBreakerRegistry).forHost(host string) *CircuitBreaker`.
- `Client` owns `circuitBreakers *circuitBreakerRegistry`.

- [x] `Step 1: Write failing registry tests`

Create `circuit_registry_test.go`:

```go
package httpx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreakerRegistryReusesHostBreaker(t *testing.T) {
	r := newCircuitBreakerRegistry(&Config{CircuitWindow: time.Minute, FailureThreshold: 1, ResetTimeout: time.Minute})
	assert.Same(t, r.forHost("query1.finance.yahoo.com"), r.forHost("QUERY1.finance.yahoo.com"))
}

func TestCircuitBreakerRegistryIsolatesHosts(t *testing.T) {
	r := newCircuitBreakerRegistry(&Config{CircuitWindow: time.Minute, FailureThreshold: 1, ResetTimeout: time.Minute})
	a := r.forHost("query1.finance.yahoo.com")
	b := r.forHost("finance.yahoo.com")
	require.NotSame(t, a, b)
	a.RecordFailure()
	assert.Equal(t, StateOpen, a.State())
	assert.Equal(t, StateClosed, b.State())
}
```

- [x] `Step 2: Run registry tests and verify RED`

Run `go test ./utils/httpx -run 'TestCircuitBreakerRegistry' -count=1`.

Expected: registry symbols do not exist.

- [x] `Step 3: Implement normalization and lazy registry creation`

Create a mutex-protected `map[string]*CircuitBreaker`. Normalize keys with
`strings.ToLower(host)` and retain ports. Normalize only breaker fields:

```go
func normalizeCircuitConfig(config *Config) {
	defaults := DefaultConfig()
	if config.CircuitWindow <= 0 { config.CircuitWindow = defaults.CircuitWindow }
	if config.ResetTimeout <= 0 { config.ResetTimeout = defaults.ResetTimeout }
	if config.FailureThreshold <= 0 && config.FailureRateThreshold <= 0 {
		config.FailureRateThreshold = defaults.FailureRateThreshold
	}
	if config.FailureThreshold <= 0 && config.MinimumRequests <= 0 {
		config.MinimumRequests = defaults.MinimumRequests
	}
}
```

The registry factory uses `NewCircuitBreaker` when `FailureThreshold > 0`;
otherwise it uses `NewFailureRateCircuitBreaker`.

- [x] `Step 4: Wire registry into Client without breaking the package`

Replace `circuitBreaker *CircuitBreaker` with:

```go
circuitBreakers *circuitBreakerRegistry
```

Call `normalizeCircuitConfig(config)` in `NewClient` and initialize
`circuitBreakers: newCircuitBreakerRegistry(config)`. Update existing `Do`
references to resolve `c.circuitBreakers.forHost(req.URL.Host)` temporarily;
Task 4 changes recording semantics.

- [x] `Step 5: Format and verify GREEN`

Run:

```bash
gofmt -w utils/httpx/circuit_registry.go utils/httpx/circuit_registry_test.go utils/httpx/client.go
go test ./utils/httpx -run 'TestCircuitBreakerRegistry' -count=1
```

Expected: registry tests and package compilation pass.

- [x] `Step 6: Review checkpoint`

Confirm registry locking only protects map creation; each breaker owns its state
lock. Run `git diff --check`. Do not commit.

---

### Task 4: Record One Final Outcome per Client.Do

`Files:`

- Modify: `utils/httpx/client.go`
- Modify: `utils/httpx/client_test.go`
- Modify: `utils/httpx/integration_test.go`

`Interfaces:`

- Consumes: `circuitBreakerRegistry.forHost(req.URL.Host)`.
- Consumes: `CircuitBreaker.record(circuitOutcomeKind)`.
- Produces: retry loop with no direct breaker mutations.

- [x] `Step 1: Add failing retry-amplification test`

```go
func TestClientRetryFailureRecordsOneBreakerOutcome(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	config := DefaultConfig()
	config.MaxAttempts = 5
	config.BackoffBaseMs = 1
	config.BackoffJitterMs = 0
	config.MaxDelayMs = 2
	config.FailureThreshold = 2
	client := NewClient(config)
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	_, err = client.Do(context.Background(), req)
	require.Error(t, err)
	breaker := client.circuitBreakers.forHost(req.URL.Host)
	assert.Equal(t, 1, breaker.Failures())
	assert.Equal(t, StateClosed, breaker.State())
}
```

- [x] `Step 2: Add failing host/classification/cancellation tests`

Add these complete tests:

```go
func TestClientCircuitBreakerIsScopedByHost(t *testing.T) {
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failing.Close()
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthy.Close()

	config := DefaultConfig()
	config.MaxAttempts = 1
	config.FailureThreshold = 1
	client := NewClient(config)
	failReq, err := http.NewRequest(http.MethodGet, failing.URL, nil)
	require.NoError(t, err)
	healthyReq, err := http.NewRequest(http.MethodGet, healthy.URL, nil)
	require.NoError(t, err)

	_, err = client.Do(context.Background(), failReq)
	require.Error(t, err)
	resp, err := client.Do(context.Background(), healthyReq)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func TestClientHTTP404IsBreakerSuccess(t *testing.T) {
	statuses := []int{http.StatusInternalServerError, http.StatusNotFound}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		status := statuses[0]
		statuses = statuses[1:]
		w.WriteHeader(status)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxAttempts = 1
	config.FailureThreshold = 0
	config.FailureRateThreshold = 0.75
	config.MinimumRequests = 2
	client := NewClient(config)
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	_, _ = client.Do(context.Background(), req)
	_, _ = client.Do(context.Background(), req)

	breaker := client.circuitBreakers.forHost(req.URL.Host)
	assert.Equal(t, 2, breaker.Samples())
	assert.Equal(t, 1, breaker.Failures())
	assert.Equal(t, StateClosed, breaker.State())
}

func TestClientCancellationIsBreakerNeutral(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxAttempts = 1
	client := NewClient(config)
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = client.Do(ctx, req)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, 0, client.circuitBreakers.forHost(req.URL.Host).Samples())
}
```

- [x] `Step 3: Run focused tests and verify RED`

Run:

```bash
go test ./utils/httpx -run 'TestClient(RetryFailureRecordsOneBreakerOutcome|CircuitBreakerIsScopedByHost|HTTP404IsBreakerSuccess|CancellationIsBreakerNeutral)$' -count=1
```

Expected: retries produce multiple failures or new outcome tests fail.

- [x] `Step 4: Move outcome recording outside the retry loop`

After span setup, resolve/gate one host breaker and defer one outcome:

```go
breaker := c.circuitBreakers.forHost(req.URL.Host)
if !breaker.Allow() {
	obsv.RecordRequest(endpoint, "error", "circuit_open")
	obsv.RecordSpanError(span, ErrCircuitOpen)
	return nil, ErrCircuitOpen
}
outcome := circuitOutcomeNeutral
defer func() { breaker.record(outcome) }()
```

Remove all `RecordFailure`/`RecordSuccess` calls from the retry loop. Before each
final return, set:

```go
// 2xx and any received response except 429/5xx:
outcome = circuitOutcomeSuccess

// terminal 429, 5xx, or transport error while ctx.Err() == nil:
outcome = circuitOutcomeFailure

// caller cancellation or pre-send failure:
outcome = circuitOutcomeNeutral
```

For transport errors:

```go
if ctx.Err() == nil { outcome = circuitOutcomeFailure }
```

Keep `Meta.Attempts`, response closing, metrics, middleware, and error wrapping
unchanged. Remove the final duplicate `RecordFailure()` after max attempts.

- [x] `Step 5: Update existing tests and verify GREEN`

Replace existing `client.circuitBreaker` test reads with:

```go
breaker := client.circuitBreakers.forHost(req.URL.Host)
```

Run:

```bash
gofmt -w utils/httpx/client.go utils/httpx/client_test.go utils/httpx/integration_test.go
go test ./utils/httpx -count=1
go test -race ./utils/httpx -count=1
```

Expected: all tests pass; no registry/outcome/probe race.

- [x] `Step 6: Review checkpoint`

Run `rg -n 'RecordFailure|RecordSuccess|circuitBreaker' utils/httpx/client.go`.
Expected: no direct record calls inside the retry loop and only the per-host
registry field. Do not commit.

---

### Task 5: Active Documentation and Full Verification

`Files:`

- Modify: `README.md`
- Modify: `CLAUDE.md`
- Modify: `docs/operations/configuration.md`
- Modify: `docs/operations/error-handling.md`
- Modify: other active API/operations docs only when the consistency scan finds a stale claim.

`Interfaces:`

- Documents: per-host breaker, rolling rate, minimum sample gate, one outcome after retries.
- Verifies: build, tests, race safety, architecture boundary, and whitespace.

- [x] `Step 1: Locate stale active documentation`

Run:

```bash
rg -n -C 3 'failure_threshold|circuit breaker|CircuitBreaker|shared breaker|failure count' README.md CLAUDE.md docs --glob '*.md'
```

Do not rewrite historical plans, specs, or audit reports.

- [x] `Step 2: Update examples and contracts`

Use this active YAML example:

```yaml
circuit_breaker:
    window: 50
    failure_threshold: 0.30
    minimum_requests: 10
    reset_timeout_ms: 30000
```

Add this validation row:

```markdown
| `circuit_breaker.minimum_requests >= 1` | 範圍 | `must be >= 1` |
```

Update `CLAUDE.md`'s `utils/httpx` description to say breakers are per-host and
record one final outcome after retries. Add this operational explanation after
the circuit config in `docs/operations/error-handling.md`:

```markdown
斷路器以 upstream host 分區。每次完整 HTTP 操作（包含所有 retry）只產生一個 outcome；
active window 至少累積 `minimum_requests` 筆後，失敗率達 `failure_threshold` 才開路。
HTTP 429、5xx 與 transport failure 視為失敗；一般 4xx 表示 upstream 可達，不歸類為
availability failure。
```

- [x] `Step 3: Run targeted regression tests`

Run:

```bash
go test ./utils/httpx ./config ./svc/scrape ./svc/twse ./facade ./cmd/dispatch -count=1
```

Expected: all packages pass.

- [x] `Step 4: Run full and race tests`

Run:

```bash
go test ./... -count=1
go test -race ./utils/httpx ./config ./svc/scrape ./svc/twse ./facade ./cmd/dispatch -count=1
```

Expected: all packages pass with no race report.

- [x] `Step 5: Build without repository artifacts`

Run:

```bash
go build -o /tmp/yfin-circuit-breaker-check .
```

Expected: exit 0 and no untracked `./yfin` binary.

- [x] `Step 6: Verify architecture and final diff`

Run:

```bash
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./cmd/... | grep 'github.com/bizshuk/yfin/svc/'
git diff --check
git status --short
git diff --stat
```

Expected: import grep has no output; whitespace passes; changes are limited to
the approved routing/breaker/config/test/doc scope.

- [x] `Step 7: Report remaining live limitations`

State that host routing and breaker amplification are fixed, but Yahoo
crumb/fingerprint 429 and endpoint migration remain separate. Do not claim all
28 live failures are resolved until a live batch rerun confirms its artifacts.
