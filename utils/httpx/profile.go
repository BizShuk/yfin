// profile.go — Per-host configuration (`HostProfile`) and a generic endpoint tagger. The profile carries QPS / Burst / MaxBody / UA / default headers and an `EndpointFn` for tagging requests; later tasks wire it into the limiter and body-cap path. Capacity: 1 type, 1 helper.
package httpx

import (
	"net/http"
	"strings"
)

// HostProfile bundles every per-host knob the unified client needs.
// Instances are typically registered once at process start (e.g.
// `YahooProfile`, `TWSEProfile`) and looked up by host. This task
// introduces the type only — limiter / body-cap wiring is deferred.
type HostProfile struct {
	Host       string
	EndpointFn func(path string) string
	QPS        float64
	Burst      int
	MaxBody    int64 // 0 means unlimited
	UserAgent  string
	Headers    http.Header
}

// CommonEndpointFn tags a URL path by its first non-empty segment after
// stripping leading slashes. Examples:
//
//	"/v8/finance/chart/AAPL" -> "v8"
//	"STOCK_DAY"              -> "STOCK_DAY"
//	"/"                     -> "root"
//	""                      -> "root"
//
// Use as a default `EndpointFn` for hosts whose paths do not need a
// bespoke mapping.
func CommonEndpointFn(path string) string {
	for len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if i := strings.IndexByte(path, '/'); i >= 0 {
		path = path[:i]
	}
	if path == "" {
		return "root"
	}
	return path
}
