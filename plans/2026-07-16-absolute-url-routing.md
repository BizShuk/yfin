# Absolute URL Routing Implementation Plan

> `For agentic workers:` implement inline with `test-driven-development`; every production change requires a test that was observed failing first.

`Goal:` Preserve the requested host when `svc/scrape` sends an absolute Yahoo Finance URL through a shared `httpx.Client` whose configured `BaseURL` points at a different Yahoo host.

`Architecture:` Keep the existing `httpx.Caller` interface and shared client. Extend `Client.Get` so a relative target resolves against `Config.BaseURL`, while an absolute target is used verbatim; then make `svc/scrape.Client.Fetch` pass the complete parsed URL instead of discarding its scheme and host.

`Tech stack:` Go 1.26, stdlib `net/http`, `net/url`, `httptest`, Testify.

## Global constraints

- Preserve `cmd -> facade -> svc -> model`.
- Do not add third-party dependencies.
- Do not change retry, circuit-breaker, cookie, crumb, or endpoint-selection behavior in this slice.
- Keep `httpx.Caller` as a one-method interface.

---

### Task 1: Absolute target resolution in `httpx.Client.Get`

`Files:`

- Modify: `utils/httpx/caller_test.go`
- Modify: `utils/httpx/caller.go`

`Interface:`

- Consumes: `(*httpx.Client).Get(ctx context.Context, target string, query url.Values)`
- Produces: relative targets use `Config.BaseURL`; absolute targets retain their own scheme and host.

- [x] Step 1: Add `TestGet_AbsoluteURLOverridesBaseURL` using two `httptest.Server` instances. Configure the client with the first server as `BaseURL`, call `Get` with the second server's absolute URL, and assert only the second server receives the request.
- [x] Step 2: Run `go test ./utils/httpx -run TestGet_AbsoluteURLOverridesBaseURL -count=1` and verify it fails because the configured base URL is incorrectly prepended to the absolute target.
- [x] Step 3: Update `Client.Get` to parse `target`; prepend `Config.BaseURL` only when the parsed target is not absolute. Update the `Caller` and `Get` comments to state the contract.
- [x] Step 4: Run `go test ./utils/httpx -run 'TestGet_(AbsoluteURLOverridesBaseURL|PopulatesMetaOnSuccess)' -count=1` and verify both tests pass.

### Task 2: Preserve the scraper URL at the caller boundary

`Files:`

- Modify: `svc/scrape/client_test.go`
- Modify: `svc/scrape/client.go`

`Interface:`

- Consumes: the absolute-target behavior from Task 1.
- Produces: `Client.Fetch(ctx, urlStr)` delegates the complete parsed URL to `httpx.Caller.Get`.

- [x] Step 1: Change `TestFetch_DelegatesToCallerOnce` to assert the caller receives `https://finance.yahoo.com/quote/AAPL`, not `/quote/AAPL`.
- [x] Step 2: Run `go test ./svc/scrape -run TestFetch_DelegatesToCallerOnce -count=1` and verify it fails with the old path-only value.
- [x] Step 3: Pass `u.String()` to `c.caller.Get`; retain `u.Query()` and existing robots-policy behavior.
- [x] Step 4: Run `go test ./svc/scrape ./utils/httpx -count=1` and verify both packages pass.

### Task 3: Regression verification

`Files:`

- No production files beyond Tasks 1-2.

- [x] Step 1: Run `gofmt -w utils/httpx/caller.go utils/httpx/caller_test.go svc/scrape/client.go svc/scrape/client_test.go`.
- [x] Step 2: Run `go test ./utils/httpx ./svc/scrape ./facade ./cmd/dispatch -count=1`.
- [x] Step 3: Run `go test ./...`.
- [x] Step 4: Run `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./cmd/... | grep svc/` and verify it has no output.
- [x] Step 5: Review `git diff --check` and `git diff`; do not commit without explicit authorization.
