# Per-host HTTP Profiles Implementation Plan

> For agentic workers: use `test-driven-development` and execute this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

`Status: Completed`

`Goal:` Finish the deferred portion of the external-client unification by making `utils/httpx.Client` select rate limits, request defaults, endpoint tags, and response-body caps per upstream host.

`Architecture:` Keep one generic `httpx.Client` contract. Register immutable `HostProfile` values before the first request, select profiles with `req.URL.Hostname()`, and preserve the existing `Config` fields as the fallback for unregistered hosts. Yahoo-specific circuit groups remain independent from host profiles.

`Tech stack:` Go 1.26.0, stdlib `net/http`, `sync`, existing `utils/httpx`, `httptest`, and `testify`.

## Global Constraints

- Keep `utils/httpx` domain-neutral; it must not contain Yahoo or TWSE path knowledge.
- Preserve `Caller.Get(ctx, target, query)` and `Client.Do(ctx, req)` signatures.
- Untagged and unregistered callers retain current `Config` behavior.
- Circuit-breaker identity remains `{authority, optional logical group}` and is not merged with profile identity.
- Profile registration happens during composition before concurrent requests begin.
- Do not add YAML settings or third-party dependencies.
- Keep `cmd → facade → svc → model`; `cmd/*` must not import `svc/*`.

## File Structure

- Modify `utils/httpx/profile.go`: host normalization, registration, lookup, and request-default helpers.
- Create `utils/httpx/limiter_registry.go`: lazy per-host limiter ownership.
- Modify `utils/httpx/client.go`: profile registry fields and per-request limiter/profile selection.
- Modify `utils/httpx/caller.go`: profile-specific body cap and endpoint metadata.
- Modify `utils/httpx/profile_test.go`: registration and normalization contract.
- Create `utils/httpx/limiter_registry_test.go`: independent-host throttling tests.
- Modify `utils/httpx/client_test.go`: headers, endpoint tags, and fallback behavior.
- Modify `facade/client.go`: register Yahoo host profiles on the shared client.
- Modify `facade/client_yahoo_test.go`: composition-level profile contract.
- Update `CLAUDE.md` and `docs/operations/performance.md`: active runtime behavior.

---

### Task 1: Host-profile registry contract

`Files:`

- Modify `utils/httpx/profile.go`.
- Create `utils/httpx/profile_test.go`.
- Modify `utils/httpx/client.go`.

`Interfaces:`

- Produces `func (c *Client) UseProfile(profile HostProfile)`.
- Produces `func normalizeProfileHost(host string) string`.
- Produces `func (c *Client) profileForHost(host string) (HostProfile, bool)`.

- [x] `Step 1: Write failing normalization and lookup tests`

Cover lowercase normalization, trailing-dot removal, `hostname:port` input, replacement before first use, unknown-host fallback, and defensive cloning of `http.Header`.

- [x] `Step 2: Run the focused tests and verify RED`

```bash
go test ./utils/httpx -run 'Test(UseProfile|NormalizeProfileHost|ProfileForHost)' -count=1
```

Expected: build failure because `UseProfile` and lookup helpers do not exist.

- [x] `Step 3: Implement the registry`

Add an `RWMutex` and `map[string]HostProfile` to `Client`. `UseProfile` validates `profile.Host`, normalizes with `url.Parse("//"+host).Hostname()`, clones headers, and stores the value. Lookup must return a cloned value so callers cannot mutate shared headers.

- [x] `Step 4: Run focused and complete httpx tests`

```bash
go test ./utils/httpx -run 'Test(UseProfile|NormalizeProfileHost|ProfileForHost)' -count=1
go test ./utils/httpx -count=1
```

Expected: both commands pass.

### Task 2: Per-host limiter registry

`Files:`

- Create `utils/httpx/limiter_registry.go`.
- Create `utils/httpx/limiter_registry_test.go`.
- Modify `utils/httpx/client.go`.

`Interfaces:`

- Produces `type rateLimiterRegistry` with `forHost(host string, qps float64, burst int) *RateLimiter`.
- Changes `NewRateLimiter(qps float64, burst int) *RateLimiter` so fractional scrape limits such as `0.7` remain valid.

- [x] `Step 1: Write failing fractional and host-isolation tests`

Use a deterministic clock seam to prove `0.7 QPS` is not truncated to zero and that consuming the burst for host A does not consume host B's tokens.

- [x] `Step 2: Run the limiter tests and verify RED`

```bash
go test ./utils/httpx -run 'TestRateLimiterRegistry|TestRateLimiter_FractionalQPS' -count=1
```

Expected: build failure because the registry and fractional constructor do not exist.

- [x] `Step 3: Implement lazy per-host limiters`

Normalize keys with `req.URL.Hostname()`. Registered profiles provide `QPS` and `Burst`; a zero profile value falls back to `Client.config`. Reject non-positive effective values during construction instead of allowing divide-by-zero waits.

- [x] `Step 4: Select the limiter inside Client.Do`

Replace the single `c.rateLimiter.Wait(ctx)` call with registry lookup based on the request host and selected profile. Do not change breaker selection or retry placement.

- [x] `Step 5: Run focused and complete httpx tests`

```bash
go test ./utils/httpx -run 'TestRateLimiter' -count=1
go test ./utils/httpx -count=1
```

Expected: all limiter and httpx tests pass.

### Task 3: Apply profile request defaults and body caps

`Files:`

- Modify `utils/httpx/profile.go`.
- Modify `utils/httpx/client.go`.
- Modify `utils/httpx/caller.go`.
- Modify `utils/httpx/client_test.go`.
- Modify `utils/httpx/caller_test.go`.

`Interfaces:`

- Profile `UserAgent` and `Headers` override client defaults but remain lower precedence than explicitly registered request middleware.
- Profile `EndpointFn(path)` sets `Meta.Endpoint` before observability begins.
- Profile `MaxBody` overrides `Config.MaxBodyBytes`; zero retains the existing `HostProfile` contract and means unlimited. An unregistered host uses the client fallback.

- [x] `Step 1: Write failing precedence and body-cap tests`

Use two `httptest.Server` instances with different profiles. Assert independent user agents, headers, endpoint tags, and body limits, including `MaxBody=0` as unlimited. Add one unregistered host assertion for exact backward compatibility.

- [x] `Step 2: Run focused tests and verify RED`

```bash
go test ./utils/httpx -run 'TestClient_Profile|TestGet_ProfileBodyCap' -count=1
```

Expected: assertions fail because profiles are not selected by `Do` or `Get`.

- [x] `Step 3: Apply request settings before middleware`

Seed `Meta.Endpoint` through the profile function, apply client defaults, then profile defaults, then run request middleware. Existing caller-set headers must not be erased unless the profile explicitly owns that header.

- [x] `Step 4: Apply the effective body cap in Client.Get`

After target resolution, select the profile once and pass its effective body cap to `readBody`. Keep gzip decoding and `ErrBodyTooLarge` behavior unchanged.

- [x] `Step 5: Run focused and complete package tests`

```bash
go test ./utils/httpx -run 'TestClient_Profile|TestGet_ProfileBodyCap' -count=1
go test ./utils/httpx -count=1
```

Expected: both commands pass.

### Task 4: Register Yahoo profiles and verify the full contract

`Files:`

- Modify `facade/client.go`.
- Modify `facade/client_yahoo_test.go`.
- Modify `CLAUDE.md`.
- Modify `docs/operations/performance.md`.

`Interfaces:`

- `facade.NewClient` and `facade.NewClientWithConfig` register profiles for `query1.finance.yahoo.com`, `query2.finance.yahoo.com`, `finance.yahoo.com`, and `fc.yahoo.com` before constructing service clients.
- Query hosts inherit the configured Yahoo QPS/burst; `finance.yahoo.com` uses the existing scrape defaults `0.7 QPS`, burst `1`, and `8 MiB` body cap.

- [x] `Step 1: Write a composition-level failing test`

Construct a facade client over instrumented HTTP endpoints and prove query and web requests select different profiles without exposing `svc/*` types across the facade boundary.

- [x] `Step 2: Register the fixed host profiles`

Keep profile definitions in `facade`, where service composition already occurs. Do not add finance-specific constants to `utils/httpx`.

- [x] `Step 3: Update active documentation`

Document that connection pooling is shared while limiter, body cap, headers, and endpoint tags are selected per host. Preserve the separate Yahoo circuit-group explanation.

- [x] `Step 4: Run deterministic acceptance`

```bash
go test ./utils/httpx ./facade ./svc/yahoo ./svc/scrape ./svc/twse -count=1
go test -race ./... -count=1
go build ./...
go vet ./...
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./cmd/... | grep 'github.com/bizshuk/yfin/svc/'
git diff --check
```

Expected: tests, race, build, vet, and diff checks pass; import grep emits no output.

## Final Acceptance Criteria

- [x] Fractional QPS is preserved.
- [x] Two hosts sharing one client have independent limiter state.
- [x] Profile headers, endpoint tags, and body caps are selected by hostname.
- [x] Unregistered hosts retain current `Config` behavior.
- [x] Yahoo circuit groups remain independently keyed and tested.
- [x] No Yahoo-specific host/path profile logic exists in `utils/httpx`.
- [x] Active docs match the implemented profile behavior.
