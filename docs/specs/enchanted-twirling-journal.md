# Plan: split `config/ampy_config.go` into single `config/types/` sub-package

## Context

`config/ampy_config.go` is a 681-line god file holding ~26 sub-config
structs + `Loader` (with `Load` / `GetEffectiveConfig` / env interpolation
/ redaction / validation) + flat `HTTPConfig` + root `Config` aggregator
+ adapter methods + `CreateEffectiveConfig`.

User direction (revised):
- Single sub-package (`config/types/`) — not 12 sub-packages.
- Each config in its own file inside that sub-package.
- Root `config/ampy_config.go` reduced to a thin alias re-exporting
  `Config` (and only `Config`) so external callers keep `config.Config` as
  a stable import.
- All other callers migrate to `github.com/bizshuk/yfin/config/types`.

## Target structure

```
config/
├── ampy_config.go              # thin alias: package doc + `type Config = types.Config`
├── effective.yaml              # (data, kept in place)
├── example.dev.yaml            # (data, kept in place)
├── example.prod.yaml           # (data, kept in place)
└── example.staging.yaml        # (data, kept in place)

config/types/                   # single sub-package, all implementation
├── config.go                   # Config root struct
├── adapters.go                 # (Config).GetHTTPConfig / GetBusConfig / GetFXConfig / GetScrapeConfig / ValidateInterval / ValidateAdjustmentPolicy
├── http.go                     # HTTPConfig (flat output) + FromConfig factory
├── loader.go                   # Loader struct + NewLoader + Load + interpolateEnvVars + interpolateString + mapToConfig + validate + GetEffectiveConfig + redactSecrets + redactSecretPatterns
├── defaults.go                 # CreateEffectiveConfig
├── app.go                      # AppConfig
├── yahoo.go                    # YahooConfig
├── concurrency.go              # ConcurrencyConfig
├── rate_limit.go               # RateLimitConfig
├── sessions.go                 # SessionsConfig
├── retry.go                    # RetryConfig + CircuitBreakerConfig (shared by yahoo + bus)
├── markets.go                  # MarketsConfig
├── fx.go                       # FXConfig + YahooWebConfig
├── bus.go                      # BusConfig + PublisherConfig + NATSConfig + KafkaConfig
├── scrape.go                   # ScrapeConfig + ScrapeRetryConfig + ScrapeEndpointConfig
├── observability.go            # ObservabilityConfig + LogsConfig + MetricsConfig + PrometheusConfig + TracingConfig + OTLPConfig
└── secrets.go                  # SecretConfig
```

## Dependency / cycle handling

- `config/types/` has **no imports** of `github.com/bizshuk/yfin/config`. This
  breaks the would-be cycle (root → types → root).
- `config/ampy_config.go` imports `config/types` for the alias only.
- All callers import `config/types` directly and use `types.X`.

## `config/ampy_config.go` target (~10 lines)

```go
// Package config is a thin alias kept so existing callers can continue
// to write `config.Config`. New code should import
// github.com/bizshuk/yfin/config/types directly — all structs, the
// Loader, the default-writer, and adapter methods live there now.
package config

import "github.com/bizshuk/yfin/config/types"

// Config is the canonical root config; see github.com/bizshuk/yfin/config/types.
type Config = types.Config
```

## Call-site migration

| File                                       | Change                                                   |
| ------------------------------------------ | -------------------------------------------------------- |
| `cmd/client.go:33`                         | `config.NewLoader` → `types.NewLoader`                   |
| `cmd/client.go:59`                         | `*config.HTTPConfig` → `*types.HTTPConfig`               |
| `cmd/admin/admin.go:67`                    | `config.NewLoader` → `types.NewLoader`                   |
| `cmd/scrape/scrape_run.go:78`              | `config.NewLoader` → `types.NewLoader`                   |
| `cmd/scrape/scrape_run.go:215`             | `*config.ScrapeConfig` → `*types.ScrapeConfig`           |
| `cmd/soak/main.go:75`                      | `config.NewLoader` → `types.NewLoader`                   |
| `cmd/soak/orchestrator_test.go:11`         | `config` import → `types` import; `*config.Config` → `*types.Config` |
| `cmd/market/pull.go:105`                   | `config.NewLoader` → `types.NewLoader`                   |
| `cmd/fundamentals/stats.go:55`             | `config.NewLoader` → `types.NewLoader`                   |
| `cmd/fundamentals/profile.go:54`           | `config.NewLoader` → `types.NewLoader`                   |
| `cmd/fundamentals/stats_helpers.go:15`     | `*config.ScrapeConfig` → `*types.ScrapeConfig`           |
| `facade/scrape.go:65`                      | `*config.ScrapeConfig` → `*types.ScrapeConfig`           |
| `tests/unit/config_test.go`                | `config.NewLoader` → `types.NewLoader`; `config.CreateEffectiveConfig` → `types.CreateEffectiveConfig` |
| `config/ampy_config_test.go`               | Move file to `config/types/types_test.go`; switch to package `types`; update all imports / calls. |

Test field access patterns (`cfg.Yahoo.BaseURL`, `cfg.App.Env`, ...) stay unchanged.

## Verification

1. `go build ./...`
2. `go vet ./...`
3. `go test ./config/...` (this exercises `types.Loader`)
4. `go test ./tests/unit/...`
5. `go test ./...`
6. `make build && ./yfin --help` — CLI launches; `yfin config --print-effective` works.
