# Versioning and Release Policy

yfin tags follow Semantic Versioning (`vMAJOR.MINOR.PATCH`). Repository tags are the source of truth：

```bash
git tag --list --sort=-version:refname
```

## Compatibility Surface

Version impact must consider：

- Exported Go API in `facade`, `model`, `config`, and supported utility packages.
- CLI command / flag behavior and exit codes.
- JSON field meanings and cache request identity.
- YAML configuration schema.
- Normalized metadata / legacy logical schema identifiers.

Internal refactoring is not breaking only when these externally observed contracts remain compatible。

## Version Selection

### MAJOR

Use for incompatible changes such as：

- removing or renaming exported APIs;
- changing an exported field type or meaning;
- removing a CLI command / flag without a compatibility window;
- incompatible config or persisted cache format changes.

### MINOR

Use for backward-compatible additions：

- new exported method / type;
- new optional CLI flag;
- new endpoint / command;
- additive config field.

### PATCH

Use for compatible fixes：

- parser / normalization bug fix;
- documentation correction;
- internal refactor;
- dependency or security update without public-contract break.

## Deprecation

Go API deprecation 使用 standard doc comment：

```go
// Deprecated: use NewSurface instead.
```

CLI compatibility flags should：

1. remain accepted;
2. explain whether they have runtime effect;
3. emit a warning when appropriate;
4. be removed only in a planned breaking release.

`--sessions` currently remains accepted but has no runtime effect。

## Build Metadata

Version command reads：

- `cmd.Version`
- `cmd.Commit`
- `cmd.Date`

Inject them at build time：

```bash
go build \
  -ldflags "\
-X github.com/bizshuk/yfin/cmd.Version=v1.2.3 \
-X github.com/bizshuk/yfin/cmd.Commit=$(git rev-parse --short HEAD) \
-X github.com/bizshuk/yfin/cmd.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o yfin .
```

Without ldflags, output intentionally uses development placeholders。

## Release Checklist

This repository currently has no checked-in automated release workflow，so publishing remains an explicit operator action：

1. Review public API / CLI / config compatibility.
2. Update [CHANGELOG](../CHANGELOG.md) with durable user-visible changes.
3. Run deterministic verification.
4. Compile integration-tagged packages without live requests.
5. Run live provider acceptance separately when required.
6. Build the release binary with metadata and smoke-test `version` / `--help`.
7. Create an annotated tag only after review.
8. Push tag / publish artifacts only with explicit release authorization.

Verification baseline：

```bash
go fmt ./...
go test ./...
go test -tags=integration ./... -run '^$'
go vet ./...
go build ./...
git diff --check
```

## Go Module Consumers

Consumers pin a concrete tag：

```bash
go get github.com/bizshuk/yfin@v1.2.3
```

Import paths do not contain a version suffix until a future major version requires Go semantic import versioning。

## Legacy Schema Identifier

Some normalized JSON retains `ampy.*` in `Meta.SchemaVersion` and golden manifest `schema_fqdn`。This is a compatibility identifier for the serialized shape, not an active protobuf package or publisher pipeline。Changing it requires an explicit compatibility decision and fixture migration。

## References

- [Semantic Versioning](https://semver.org/)
- [Go Modules Reference](https://go.dev/ref/mod)
- [CHANGELOG](../CHANGELOG.md)
