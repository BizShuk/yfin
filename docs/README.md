# yfin Documentation Index

Business definition / quick start 由 repository [README](../README.md) 擁有；architecture / ownership 由 [CLAUDE](../CLAUDE.md) 擁有；domain terms 由 [terminology.md](terminology.md) 擁有。

## Getting Started

- [Installation](getting-started/install.md)
- [Onboarding](getting-started/onboarding.md)
- [Package Navigation](getting-started/packages.md)

## CLI

- [Usage Workflows](cli/usage.md)
- [Commands Reference](cli/commands.md)
- [Standalone Soak](cli/soak-testing.md)

Exact runtime flags 仍以 `yfin --help` / `<command> --help` 為準。

## Go API

- [API Surface](api/reference.md)
- [Data Structures](api/data-structures.md)
- [Examples](api/examples.md)

Exact exported declarations 以 `go doc ./facade` / `go doc ./model` 為準。

## Operations

- [Configuration](operations/configuration.md)
- [Observability](operations/observability.md)
- [Performance](operations/performance.md)
- [Error Handling](operations/error-handling.md)
- [Data Quality](operations/data-quality.md)

## Scrape

- [Overview](scrape/overview.md)
- [CLI Guide](scrape/cli.md)

## Integration and Selection

- [Python yfinance Migration](integrations/migration-guide.md)
- [Versioning](integrations/versioning.md)
- [Data Surface Comparison](comparisons/method-comparison.md)

## Project Records

- [Specifications](specs/)
- [Decision Memory](memory/)
- [CHANGELOG](CHANGELOG.md)
- [Historical Archive](history/README.md)

`docs/history/` 是 snapshot archive，不代表 current runtime contract。
