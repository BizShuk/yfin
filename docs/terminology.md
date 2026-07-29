# yfin Terminology

| Term | Definition | Source |
| --- | --- | --- |
| Facade | Public SDK boundary used by the yfin CLI and downstream applications; it hides concrete `svc/*` types. | `facade/` |
| Service | Provider-specific HTTP, decode, validation, and scrape implementation below the facade. | `svc/` |
| Model | Lowest-layer data shapes and normalization logic with no internal package dependencies. | `model/` |
| Raw root | Caller-selected directory below which yfin exclusively persists provider cache artifacts. | `utils/cache.JSONStore` |
| Read-through cache | A facade that returns a fresh matching artifact or calls its provider once and atomically persists the result. | `facade.CachedClient`, `facade.CachedTwseClient`, `facade.CachedFoundationClient` |
| Foundation command | One of the ordered 30 Yahoo data surfaces exposed by `FoundationCommands` and `FetchFoundation`. | `facade/foundation.go` |
| Refresh tier | Daily, weekly, monthly, quarterly, or annual freshness policy assigned to a foundation command. | `utils/cache/refresh.go` |
| Request identity | Stable cache key derived from all provider inputs that can change a response. | `facade/cached_client.go`, `facade/cached_twse.go` |
| Run ID | Caller-provided correlation identifier propagated through provider and scrape operations. | `facade.Client` methods |
| MIC | Market Identifier Code inferred for a security listing. | `model.InferMIC` |
