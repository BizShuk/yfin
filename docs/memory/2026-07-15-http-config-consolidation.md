# HTTP Config Consolidation Hard Decision

`Status: Durable architecture decision recorded on 2026-07-18；default-value ownership updated on 2026-07-31.`

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
└── scrape business switches only
    ├── enabled / robots_policy / cache_ttl_ms / endpoints
    └── config.assembleScrapeHTTPConfig()
        └── utils/httpx.DefaultScrapeConfig()
            └── config.Scrape.HTTP (*httpx.Config)
                └── cmd.CreateClientFromConfig()
                    └── (*facade.Client).ConfigureScrape()
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

Type ownership and scrape default values are centralized:

- Canonical type: `utils/httpx.Config`.
- Yahoo/root HTTP assembly: `config.assembleHTTPConfig`.
- Scrape defaults: `utils/httpx.DefaultScrapeConfig`.
- `config.assembleScrapeHTTPConfig` 與 `svc/scrape.DefaultConfig` 都 consume
  the shared default builder。

`config` 與 `svc/scrape` 都可向下依賴 leaf `utils/httpx`，因此不需跨 layer
複製 default literals，也不會形成 import cycle。修改 scrape transport policy
時只更新 shared builder 及其 focused test。

## Mapper Boundary

The completed goal was removal of HTTP-field mappers.

`facade.NewScrapeClientFromConfig` still maps scrape-only endpoint switches from `config.ScrapeEndpointConfig` to `scrape.EndpointConfig`. That mapping is outside HTTP configuration ownership and remains necessary because `config` cannot expose a `svc/scrape` type.

## Verified Implementation Files

- `utils/httpx/client.go`
- `utils/httpx/defaults.go`
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

Executable tests assert:

- root YAML units map into `Config.HTTP` correctly;
- `Config.Scrape.HTTP` is non-nil after load;
- shared scrape defaults have focused `utils/httpx` coverage;
- both post-load fields have the exact `*httpx.Config` type;

## Future Change Boundary

If scrape-specific HTTP YAML customization is required again, add a YAML-only DTO in `config` and assemble it into `*httpx.Config`. Do not embed `httpx.Config` into YAML structs, do not import `svc/scrape` from `config`, and do not reintroduce field-by-field mappers in `cmd` or `facade`.
