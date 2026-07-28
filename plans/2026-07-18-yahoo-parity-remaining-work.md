# Yahoo 30-command Parity Remaining Work Implementation Plan

> For agentic workers: use `test-driven-development` and execute this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

`Status: Active`

`Goal:` Make batch artifacts truthful, replace the remaining avoidable HTML paths with stable Yahoo JSON surfaces, and pass the fixed 30-command semantic parity gate for both `AAPL` and `2330.TW` without overstating upstream anti-bot limitations.

`Architecture:` Preserve `cmd → facade → svc → model`. Batch owns artifact lifecycle; `svc/yahoo` owns HTTP request and decode contracts; facade exposes only model/facade types; parity scripts remain an external acceptance oracle. No production Python, `curl_cffi`, TLS-impersonation library, or browser runtime is introduced.

`Current live baseline:` The last formal run recorded on 2026-07-17 produced `8 success / 22 failed` for AAPL. The eight healthy commands were `history`, `actions`, `income`, `balance`, `cashflow`, `news`, `isin`, and `metadata`. The 22 remaining results were 13 quoteSummary/auth-family failures, 7 analysis/analyst HTML `503` failures, `options` `401`, and `earnings-dates` with no rows. This is historical evidence and must be refreshed before claiming current parity.

`Tech stack:` Go 1.26.0, Cobra, stdlib HTTP/JSON/filesystem packages, existing `utils/httpx`, Python yfinance oracle scripts, Bash, and `testify`.

## Global Constraints

- Do not add production Python, `curl_cffi`, TLS impersonation, Playwright, or a new HTTP transport dependency.
- Keep `facade.Client` as the only fetch handle used by `cmd/dispatch`.
- Keep `model` as the lowest layer.
- Keep runtime artifacts under `~/.config/yfin/data/raw/`.
- Preserve the fixed 30-command manifest and command-specific JSON semantics.
- Treat deterministic tests and live Yahoo parity as separate gates.
- A live command is successful only when its current JSON artifact exists and its stale `_failed` artifact does not.
- Do not run `2330.TW` acceptance until the AAPL gate exits zero.
- If Yahoo still requires a browser-equivalent fingerprint for a command, record it as an explicit blocker; do not weaken the gate or mark it complete.

## Failure Inventory

| Family | Commands | Last observed failure |
| --- | --- | --- |
| quoteSummary/auth | `info`, three holder commands, three insider commands, `recommendations`, `recommendations-summary`, `upgrades`, `calendar`, `sec-filings`, `sustainability` | crumb `429` or auth-family circuit-open |
| analysis JSON candidates | `earnings-history`, `eps-trend`, `eps-revisions`, `earnings-estimates`, `revenue-estimates`, `growth-estimates`, `price-targets` | Yahoo HTML `503` |
| options | `options` | `401 Unauthorized` |
| earnings calendar | `earnings-dates` | HTML parser returned no rows |

## File Structure

- Modify `cmd/dispatch/batch.go`: remove stale error artifacts after non-failure terminal states.
- Modify `cmd/dispatch/batch_test.go` and `batch_integration_test.go`: artifact-lifecycle tests.
- Modify `scripts/verify-yf-parity.sh`: always run classification and comparison even when Go exits non-zero.
- Create `skills/scripts/classify_yfin_failures.py`: current-run family summary.
- Create `skills/scripts/test_classify_yfin_failures.py`: deterministic classifier tests.
- Create `svc/yahoo/analysis_quotesummary.go` and tests: quoteSummary analysis modules and projections.
- Modify `facade/client_yahoo.go` and tests: plain analysis surfaces.
- Modify `cmd/dispatch/dispatch.go` and tests: cut over seven analysis commands.
- Create `svc/yahoo/earnings_calendar.go` and tests: visualization JSON probe/decoder kept separate from the active HTML path until freshness is proven.
- Modify `svc/yahoo/options.go` and tests: authenticated option-chain request contract if the live probe proves it works.
- Update `plans/2026-07-18-yahoo-parity-remaining-work.md`: append exact live evidence after each formal run.
- Update `README.md`, `CLAUDE.md`, `docs/cli/commands.md`, and `docs/operations/error-handling.md` only after behavior changes.

---

### Task 1: Truthful `_failed` artifact lifecycle

`Files:`

- Modify `cmd/dispatch/batch.go`.
- Modify `cmd/dispatch/batch_test.go`.
- Modify `cmd/dispatch/batch_integration_test.go`.

`Interfaces:`

- Produces `func clearCommandFailure(rawDir, ticker, command string) error`.
- A `success`, `skipped`, or `not_found` result removes `_failed/<ticker>.<command>.err`.
- `os.ErrNotExist` is success; every other removal error is surfaced in `tickerResult.Errors`.

- [ ] `Step 1: Write failing stale-error tests`

Pre-create `_failed/AAPL.good.err`, run a successful fetch, and assert the JSON artifact exists while the stale error does not. Repeat for `skipped` and `not_found`. For the removal-error case, replace the error path with a non-empty directory so `os.Remove` deterministically returns an error on every supported platform.

- [ ] `Step 2: Run focused tests and verify RED`

```bash
go test ./cmd/dispatch -run 'TestRunBatchForTicker_(ClearsStaleFailure|SkipClearsStaleFailure|NotFoundClearsStaleFailure|FailureCleanupError)' -count=1
```

Expected: stale error assertions fail because `runBatchForTicker` never removes them.

- [ ] `Step 3: Implement cleanup at terminal-state boundaries`

Call `clearCommandFailure` only after a successful JSON rename, after a cache skip is confirmed, or after a current not-found classification. Do not clear the error before a fetch begins; an interrupted request must not erase the last known failure. A cleanup error changes the command status to `failed` and is stored in `tickerResult.Errors`.

- [ ] `Step 4: Run dispatch and lifecycle integration tests`

```bash
go test ./cmd/dispatch -count=1
```

Expected: all tests pass, including existing failure-artifact behavior.

### Task 2: Current-run parity classification

`Files:`

- Create `skills/scripts/classify_yfin_failures.py`.
- Create `skills/scripts/test_classify_yfin_failures.py`.
- Modify `scripts/verify-yf-parity.sh`.

`Interfaces:`

- Classifier consumes `--go-root`, `--ticker`, `--day`, and the fixed commands from `skills/scripts/config.py`.
- Classifier prints exact `success`, `failed`, `missing`, and family counts and exits non-zero unless all 30 commands have valid JSON with no matching error.
- The shell runner captures Python, Go, classifier, and semantic-comparator statuses independently, then returns non-zero if any stage fails.

- [ ] `Step 1: Write classifier fixture tests`

Create temporary fixtures for the recorded `8/22` split, a stale-error-plus-success conflict, invalid JSON, and full `30/30`. Assert stable family counts and exit codes.

- [ ] `Step 2: Run tests and verify RED`

```bash
python3 skills/scripts/test_classify_yfin_failures.py
```

Expected: import failure because the classifier does not exist.

- [ ] `Step 3: Implement the classifier`

Classify by the four fixed families in this plan. A success artifact plus `_failed` file is `conflict`, not success. Read only the requested UTC date.

- [ ] `Step 4: Make the shell gate run every diagnostic stage`

Temporarily disable immediate exit only around the Go batch and comparison commands, capture statuses, restore `set -e`, print the four stage codes, and exit `1` if any is non-zero.

- [ ] `Step 5: Run deterministic script tests`

```bash
python3 skills/scripts/test_classify_yfin_failures.py
bash -n scripts/verify-yf-parity.sh
```

Expected: both pass.

### Task 3: Replace seven analysis HTML commands

`Files:`

- Create `svc/yahoo/analysis_quotesummary.go`.
- Create `svc/yahoo/analysis_quotesummary_test.go`.
- Modify `facade/client_yahoo.go`.
- Modify `facade/client_yahoo_test.go`.
- Modify `cmd/dispatch/dispatch.go`.
- Modify `cmd/dispatch/dispatch_test.go`.

`Interfaces:`

- `earnings-history` decodes quoteSummary module `earningsHistory`.
- `earnings-estimates`, `revenue-estimates`, `eps-trend`, and `eps-revisions` project their matching nested fields from `earningsTrend`.
- `growth-estimates` combines the stock growth field from `earningsTrend` with `industryTrend`, `sectorTrend`, and `indexTrend` estimates.
- `price-targets` projects `financialData` target-price fields.
- Facade methods return `model` or facade-owned DTOs only.

- [ ] `Step 1: Add fixture-based decoder and projection tests`

Cover missing modules, explicit zero values, nullable raw fields, deterministic row order, and distinct payloads for the six analysis dimensions sharing upstream modules.

- [ ] `Step 2: Run service tests and verify RED`

```bash
go test ./svc/yahoo -run 'TestDecode(Analysis|EarningsTrend|PriceTargets)' -count=1
```

Expected: build failure because the decoders do not exist.

- [ ] `Step 3: Implement request and decode contracts`

Use existing quoteSummary request construction, crumb handling, and `yahoo-auth` circuit group. Do not parse HTML or reuse scrape DTOs across the service boundary.

- [ ] `Step 4: Add facade methods and dispatch cutover`

Move only the seven listed registry entries. Add a source-contract test rejecting `ScrapeAnalysisDimension` and `ScrapeAnalystInsights` in `cmd/dispatch/dispatch.go`.

- [ ] `Step 5: Run focused and complete deterministic tests`

```bash
go test ./svc/yahoo ./facade ./cmd/dispatch -count=1
go test ./... -count=1
```

Expected: all tests pass. Live success is not claimed until Task 6.

### Task 4: Earnings-calendar capability gate

`Files:`

- Create `svc/yahoo/earnings_calendar.go`.
- Create `svc/yahoo/earnings_calendar_test.go`.
- Create `scripts/probe-yahoo-earnings-calendar.sh`.
- Modify `svc/yahoo/earnings_dates.go`, `facade/yahoo.go`, and `cmd/dispatch/dispatch_test.go` only if the live freshness gate passes.

`Interfaces:`

- Add a side-by-side decoder for Yahoo's `POST /v1/finance/visualization` JSON surface using `entityIdType=earnings`, a ticker equality filter, `startdatetime` descending sort, and the six yfinance fields.
- The current Python yfinance source states that this surface stopped updating in 2025; it cannot replace HTML unless AAPL and 2330.TW freshness matches the Python oracle during the same run.
- POST request bodies must remain replayable through retries.

- [ ] `Step 1: Write request-contract and decoder tests`

Assert method, endpoint, exact JSON filter/sort/fields, content type, response mapping, empty result, nullable EPS fields, and retry body replay.

- [ ] `Step 2: Run focused tests and verify RED`

```bash
go test ./svc/yahoo -run 'Test(Fetch|Decode)EarningsCalendar' -count=1
```

Expected: build failure because the visualization client does not exist.

- [ ] `Step 3: Implement the isolated visualization probe`

Use the existing `httpx.Client.Do` transport and `yahoo-web` group. Keep the production `FetchEarningsDates` path unchanged at this checkpoint.

- [ ] `Step 4: Compare live freshness before cutover`

The probe prints the newest and oldest event timestamps from visualization and the Python oracle for both tickers without printing cookies. Cut over only when both datasets cover the same current event window. If visualization is stale or HTML remains fingerprint-blocked, record `earnings-dates` as blocked and make no production cutover.

- [ ] `Step 5: Run focused and complete Yahoo tests`

```bash
go test ./svc/yahoo ./facade ./cmd/dispatch -count=1
```

Expected: all tests pass.

### Task 5: Auth and options capability gate

`Files:`

- Create `scripts/probe-yahoo-auth.sh`.
- Create `svc/yahoo/auth_live_test.go` with `//go:build integration`.
- Modify `svc/yahoo/options.go` and tests only if the authenticated stdlib request succeeds.
- Append probe evidence to this plan.

`Interfaces:`

- Probe `quoteSummary`, `options`, and crumb/cookie bootstrap with the same configured stdlib transport used by production.
- Record HTTP status, authority, family, and whether Python yfinance succeeds for the same ticker and time window.
- Never print cookie or crumb values.

- [ ] `Step 1: Add an opt-in integration probe`

Require `YFIN_LIVE_YAHOO=1`; otherwise skip. Test AAPL sequentially with retry max `1` so breaker amplification does not hide root statuses.

- [ ] `Step 2: Run the formal stdlib probe`

```bash
YFIN_LIVE_YAHOO=1 go test -tags=integration ./svc/yahoo -run TestLiveYahooAuthCapability -count=1 -v
```

Expected: exact statuses are printed without secrets.

- [ ] `Step 3: Apply only evidence-supported options changes`

If the same cookie/crumb contract makes options return `2xx`, add it with fixture tests and retain `yahoo-options` isolation. If it remains `401` while Python succeeds, record `options` as fingerprint-blocked and make no speculative production change.

- [ ] `Step 4: Record the quoteSummary boundary`

If stdlib quoteSummary remains `429` while Python yfinance succeeds, leave all 13 auth commands explicitly blocked. Resolving that condition requires a separately approved transport-policy change because this plan prohibits fingerprint impersonation dependencies.

### Task 6: Formal AAPL then 2330.TW acceptance

`Files:`

- Modify `README.md`, `CLAUDE.md`, `docs/cli/commands.md`, and `docs/operations/error-handling.md`.
- Append exact results to this plan.

- [ ] `Step 1: Run deterministic gates`

```bash
go test ./... -count=1
go test -race ./... -count=1
go build ./...
go vet ./...
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./cmd/... | grep 'github.com/bizshuk/yfin/svc/'
git diff --check
```

Expected: test, race, build, vet, and diff checks pass; import grep emits no output.

- [ ] `Step 2: Run AAPL parity`

```bash
./scripts/verify-yf-parity.sh AAPL
```

Expected for completion: `30 success`, zero failed/missing/conflict, semantic comparator exit `0`.

- [ ] `Step 3: Run 2330.TW parity only after AAPL passes`

```bash
./scripts/verify-yf-parity.sh 2330.TW
```

Expected for completion: `30 success`, zero failed/missing/conflict, semantic comparator exit `0`.

- [ ] `Step 4: Synchronize active documentation`

Document only the endpoints and live support proven by both runs. If Task 5 remains fingerprint-blocked, keep this plan active and do not claim full parity.

## Final Acceptance Criteria

- [ ] Current success, skipped, and not-found results cannot coexist with stale error artifacts.
- [ ] The runner always reports all stage statuses.
- [ ] Seven analysis commands no longer use Yahoo HTML scrape paths.
- [ ] Earnings dates use JSON and treat an empty result consistently.
- [ ] Options behavior is evidence-backed.
- [ ] AAPL 30-command semantic parity exits `0`.
- [ ] 2330.TW 30-command semantic parity exits `0`.
- [ ] Deterministic gates pass with zero `cmd → svc` direct imports.
