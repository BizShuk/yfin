# Downstream Cache Ownership

## Decision

Downstream applications may select an app-owned raw root, but yfin exclusively
owns Yahoo/TWSE access, cache request identity, refresh policy, path validation,
and atomic raw persistence.

Supported boundaries are:

- `facade.NewCachedClient` for Yahoo daily bars
- `facade.NewCachedTwseClient` for TWSE dispatches
- `facade.NewCachedFoundationClient` for the ordered foundation surface
- `facade.NewRawCache` for read-only legacy artifact access

The yfin batch CLI and downstream consumers share
`facade.FoundationCommands` and `(*facade.Client).FetchFoundation`; command
mapping must not be copied into another package or repository.

## Rationale

A prior Stock bridge launched Python `yfinance` and wrote raw JSON directly.
That created two provider implementations and two cache owners. A single yfin
boundary gives every consumer the same rate limiting, endpoint mapping,
freshness semantics, safe paths, and atomic writes.

## Constraints

- A downstream app may coordinate ticker/command workers and shape normalized
  application data.
- It must not import `yfin/svc/*`, implement provider HTTP protocols, launch a
  provider script, or write provider raw artifacts.
- Cache-safe filename encoding must never alter the symbol sent to a provider.

## Revisit

Revisit only if yfin is intentionally split into separately versioned provider
and cache SDKs with an equally explicit ownership boundary.
