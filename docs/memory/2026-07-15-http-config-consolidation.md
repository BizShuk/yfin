# HTTP Config Consolidation Hard Decision

`Status: Durable architecture decision recorded on 2026-07-18.`

## Outcome

The runtime HTTP configuration type is now owned by `utils/httpx.Config`.

- `config.HTTPConfig` is a compatibility alias for `httpx.Config`.
- `config.Config.HTTP` and `config.ScrapeConfig.HTTP` hold `*httpx.Config`.
- `cmd.CreateClient` consumes `cfg.GetHTTPConfig()` directly.
- `svc/scrape.Config` holds `HTTP *httpx.Config`.
- The old `cmd/client.go:httpConfigToHttpx` mapper and scrape HTTP field-by-field mappers are gone.
- The public YAML shape for root Yahoo HTTP settings remains assembled by `config.Config.assembleHTTPConfig`.

## Actual Data Flow

```tree
config/effective.yaml
├── yahoo + rate_limit + retry + circuit_breaker
│   └── config.assembleHTTPConfig()
│       └── config.Config.HTTP (*httpx.Config)
│           └── cmd.CreateClient()
│               └── facade.NewClientWithConfig()
│
└── scrape business switches only
    ├── enabled / robots_policy / cache_ttl_ms / endpoints
    └── config.assembleScrapeHTTPConfig()
        └── Go-owned scrape HTTP defaults
            └── config.Scrape.HTTP (*httpx.Config)
                └── facade.NewScrapeClientFromConfig()
                    └── svc/scrape.Config.HTTP
```

## Divergence from the Original Plan

The original plan said scrape HTTP YAML keys such as `user_agent`, `timeout_ms`, `qps`, `burst`, and `retry.*` would remain backward compatible. That did not happen.

The implemented contract is:

- Scrape HTTP knobs are no longer accepted from YAML.
- `config/effective.yaml` exposes only scrape business switches.
- Scrape HTTP customization is a Go API operation through `cfg.Scrape.HTTP`, `facade.NewScrapeClientFromConfig`, or a caller-supplied `httpx.Client`.
- Existing deployments that relied on scrape-specific HTTP YAML overrides must migrate to Go-side configuration; the loader does not preserve those removed keys.

This was a deliberate simplification in commit `bcbb839`, not an unimplemented YAML adapter.

## Source-of-truth Boundary

Type ownership is centralized, but default values are not stored in one file.

- Canonical type: `utils/httpx.Config`.
- Yahoo/root HTTP assembly: `config.assembleHTTPConfig`.
- Scrape defaults at service boundary: `svc/scrape.DefaultConfig`.
- Mirrored scrape defaults at config leaf: `config.scrapeHTTPDefaults`.

The two scrape default builders intentionally duplicate values because `config` is a leaf and cannot import `svc/scrape`. Tests must keep the following values aligned:

| Field | Value |
| --- | --- |
| Base URL | `https://finance.yahoo.com` |
| Timeout | `10s` |
| QPS / burst | `0.7 / 1` |
| Attempts | `4` |
| Backoff | `300ms` base, `150ms` jitter, `4000ms` max |
| Body cap | `8 MiB` |
| Circuit rate threshold | `0.30` after `10` requests |

This is a documented cross-layer synchronization obligation, not a claim of zero duplicated values.

## Mapper Boundary

The completed goal was removal of HTTP-field mappers.

`facade.NewScrapeClientFromConfig` still maps scrape-only endpoint switches from `config.ScrapeEndpointConfig` to `scrape.EndpointConfig`. That mapping is outside HTTP configuration ownership and remains necessary because `config` cannot expose a `svc/scrape` type.

## Verified Implementation Files

- `utils/httpx/client.go`
- `config/http.go`
- `config/config.go`
- `config/scrape.go`
- `config/adapters.go`
- `config/loader.go`
- `config/loader_test.go`
- `svc/scrape/types.go`
- `svc/scrape/client.go`
- `facade/scrape.go`
- `cmd/client.go`

## Validation Contract

```bash
go test ./config ./svc/scrape ./facade ./cmd -count=1
go test -race ./... -count=1
go build ./...
go vet ./...
git diff --check
```

The configuration tests must assert:

- root YAML units map into `Config.HTTP` correctly;
- `Config.Scrape.HTTP` is non-nil after load;
- scrape body cap remains `8 MiB`;
- both post-load fields have the exact `*httpx.Config` type;
- removed scrape YAML HTTP keys are not documented as supported.

## Future Change Boundary

If scrape-specific HTTP YAML customization is required again, add a YAML-only DTO in `config` and assemble it into `*httpx.Config`. Do not embed `httpx.Config` into YAML structs, do not import `svc/scrape` from `config`, and do not reintroduce field-by-field mappers in `cmd` or `facade`.
