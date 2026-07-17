# Yahoo Endpoint Migration Implementation Plan

> `For agentic workers:` REQUIRED SUB-SKILL: use `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

`Goal:` Replace the four stdlib-compatible HTML scrape failures with Yahoo fundamentals-timeseries and news XHR endpoints without changing CLI output types or adding dependencies.

`Architecture:` Implement endpoint-specific fetch/decode in `svc/yahoo`, expose model-only facade methods, and switch only four dispatch registry entries. Construct news POST requests with a replayable stdlib body and retain a regression test for retry parity.

`Tech Stack:` Go 1.26, stdlib `net/http` and `encoding/json`, Testify, `httptest`.

## Global Constraints

- Preserve `cmd -> facade -> svc -> model`; `cmd/*` must not import `svc/*`.
- Preserve the existing `model.FundamentalsSnapshot` and `model.NewsItem` JSON schemas.
- Do not add third-party dependencies.
- Do not change the remaining 22 cookie/crumb/fingerprint failures.
- Keep existing scrape methods available for other callers.
- Follow RED -> GREEN for every production behavior change.
- Do not commit without explicit user authorization.

---

### Task 1: Verify POST Body Replay During Retry

`Files:`

- Modify: `utils/httpx/client_test.go`
- Modify: `utils/httpx/client.go`

- [x] `Step 1: Add a POST retry regression test`

Use an `httptest.Server` that returns `500` once and `200` next. Record both request bodies and assert they contain identical JSON.

- [x] `Step 2: Run the focused test and evaluate the result`

The test passed before production changes: Go 1.26's stdlib request/transport path replays a `bytes.Reader` body through `GetBody` across the existing outer retry loop.

- [x] `Step 3: Avoid an unnecessary production change`

Use `bytes.NewReader` for the news request and retain the passing regression test. No `httpx.Client.Do` change is required.

- [x] `Step 4: Run focused httpx tests and verify GREEN`

```bash
go test ./utils/httpx -run 'TestClientRetry|TestClientRetryReplaysPOSTBody' -count=1
```

---

### Task 2: Annual Fundamentals-timeseries Service

`Files:`

- Create: `svc/yahoo/timeseries.go`
- Create: `svc/yahoo/timeseries_test.go`

- [x] `Step 1: Add failing decode and request-contract tests`

Cover statement type validation, exact type query, latest annual point selection, missing series omission, line order, source, currency, and period bounds.

- [x] `Step 2: Run focused tests and verify RED`

```bash
go test ./svc/yahoo -run 'Test(FetchFinancialStatement|DecodeFinancialStatement)' -count=1
```

- [x] `Step 3: Implement statement definitions, request builder, decoder, and model conversion`

Use absolute query2 URL, private response envelopes, and the stable model-key mappings in the approved spec.

- [x] `Step 4: Run focused service tests and verify GREEN`

```bash
go test ./svc/yahoo -run 'Test(FetchFinancialStatement|DecodeFinancialStatement)' -count=1
```

---

### Task 3: News XHR Service

`Files:`

- Create: `svc/yahoo/news.go`
- Create: `svc/yahoo/news_test.go`

- [x] `Step 1: Add failing POST contract and decode tests`

Assert method, query, JSON payload, ad filtering, URL fallback, provider, summary, RFC3339 timestamp, symbol, and invalid-item filtering.

- [x] `Step 2: Run focused tests and verify RED`

```bash
go test ./svc/yahoo -run 'Test(FetchNews|DecodeNews)' -count=1
```

- [x] `Step 3: Implement XHR request and model conversion`

Use the exact yfinance-compatible `latestNews` payload and return `[]model.NewsItem`.

- [x] `Step 4: Run focused service tests and verify GREEN`

```bash
go test ./svc/yahoo -run 'Test(FetchNews|DecodeNews)' -count=1
```

---

### Task 4: Facade and Dispatch Cutover

`Files:`

- Modify: `facade/client_yahoo.go`
- Modify: `cmd/dispatch/dispatch.go`
- Modify: `cmd/dispatch/dispatch_test.go`

- [x] `Step 1: Add facade methods returning only facade/model types`

Add `FetchIncomeStatement`, `FetchBalanceSheet`, `FetchCashFlowStatement`, and `FetchNews`; infer MIC for statement snapshots.

- [x] `Step 2: Switch only income, balance, cashflow, and news registry entries`

Keep the six analysis commands and analyst insights on the existing scrape surface.

- [x] `Step 3: Add a source-level dispatch contract test`

Assert the four registry entries reference the new facade methods and do not reference `ScrapeFinancials`, `ScrapeBalanceSheet`, `ScrapeCashFlow`, or `ScrapeNews`.

- [x] `Step 4: Run facade and dispatch tests`

```bash
go test ./facade ./cmd/dispatch -count=1
```

---

### Task 5: Documentation and Deterministic Verification

`Files:`

- Modify: `README.md`
- Modify: `CLAUDE.md`
- Modify active endpoint documentation found by consistency scan.

- [x] `Step 1: Update active architecture and endpoint claims`

Document that statements and news use Yahoo API/XHR while legacy scrape methods remain available.

- [x] `Step 2: Run consistency searches`

```bash
rg -n 'income|balance|cashflow|news|ScrapeFinancials|fundamentals-timeseries|xhr/ncp' README.md CLAUDE.md docs cmd facade svc
```

- [x] `Step 3: Run full deterministic gates`

```bash
go test ./... -count=1
go test -race ./utils/httpx ./svc/yahoo ./facade ./cmd/dispatch -count=1
go build -o /tmp/yfin-endpoint-migration-check .
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./cmd/... | grep svc/
git diff --check
```

---

### Task 6: Live AAPL Batch Acceptance

- [x] `Step 1: Run the fixed 30-command manifest`

```bash
go run . --config config/effective.yaml --retry-max 1 --timeout 12s batch --ticker AAPL --force
```

- [x] `Step 2: Verify the four migrated artifacts`

Confirm success artifacts exist for `income`, `balance`, `cashflow`, and `news`, and no corresponding `_failed` files remain.

- [x] `Step 3: Record the live result`

Expected minimum: `8 success / 22 failed`, with zero circuit-open failures. Record any Yahoo-side drift separately without widening this slice.

Live results on 2026-07-17:

- Formal effective config (`minimum_requests: 10`): `7 success / 23 failed`. All four migrated commands succeeded, but the existing query-host breaker opened after the concurrent crumb `429` group and rejected the otherwise healthy `metadata` command.
- Isolated endpoint diagnostic (`minimum_requests` temporarily raised to `1000`): `8 success / 22 failed / 0 circuit-open`.
- The temporary diagnostic value was immediately restored to the formal value `10`.
- Artifacts were valid JSON: income 21 lines, balance 5 lines, cashflow 5 lines, news 10 items.

The endpoint migration meets its four-command acceptance. The formal-config aggregate count remains one below the target because host-level breaker isolation does not separate failing quoteSummary/crumb traffic from healthy chart traffic on the same Yahoo authority; that is a separate breaker-keying policy decision.
