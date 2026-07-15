// http.go — `HTTPConfig` is now a type alias for `httpx.Config` (the
// canonical HTTP-layer config). `(*Config).GetHTTPConfig` returns a
// populated `*httpx.Config` post-load; no extra adapter struct is
// maintained. Callers reading `cfg.GetHTTPConfig().QPS`,
// `.MaxAttempts`, etc. work unchanged.
//
// Historical note: the original `config.HTTPConfig` was a flat struct
// with yaml-shaped fields (Timeout time.Duration, FailureThreshold
// float64). With the `*httpx.Config` consolidation it became an alias
// so existing imports of `config.HTTPConfig` continue to compile, and
// fields read against `httpx.Config` rather than a near-duplicate.
//
// Capacity: 1 type alias.
package config

import "github.com/bizshuk/yfin/utils/httpx"

// HTTPConfig is the canonical post-load HTTP config. This alias is
// kept so existing callers (`cmd/client.go` etc.) and tests compile
// without modification; new code should reference `httpx.Config`
// directly. The shape is owned by `utils/httpx`.
type HTTPConfig = httpx.Config
