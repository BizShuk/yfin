// client.go — TWSE HTTP transport factory + browser-like User-Agent. The
// `httpx.Client` is tuned for TWSE's public REST API (low QPS, conservative
// retry) and registers TWSEMiddleware so every request carries the canonical
// User-Agent TWSE would otherwise reject (TWSE rejects the default Go UA).
// Capacity: 1 const + 1 builder function.
package twse

import (
	"time"

	"github.com/bizshuk/yfin/svc/twse"
	"github.com/bizshuk/yfin/utils/httpx"
)

// twseUserAgent is the browser-like UA TWSE rejects the default Go
// `User-Agent` for.
const twseUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

// buildTWSEClient returns the process-wide `*twse.Client` used by the
// `yfin twse` subcommand. The client captures `twse.BaseURL` at construction
// time; tests pass `*twse.Client` built via `svc/twse.NewClientWithURL` to
// point at an httptest server.
//
// Wiring pattern for future services (illustrative — actual Yahoo +
// scrape migration is out of Task 4's scope):
//
//	yc := yahoo.NewCrumbManager(httpx.NewClient(...), cookieURL, apiURL)
//	hc.Use(httpx.YahooMiddleware(yc))   // crumb injected per request
//	sc := httpx.NewClient(&httpx.Config{...ScrapeProfile...})
//	sc.Use(httpx.ScrapeMiddleware(ua))  // browser headers
func buildTWSEClient() *twse.Client {
	hc := httpx.NewClient(&httpx.Config{
		// BaseURL is intentionally empty: twse.Client owns the full
		// TWSE host+path prefix and passes pre-built absolute URLs to
		// caller.Get. Setting it here would double-concatenate.
		BaseURL:          "",
		Timeout:          30 * time.Second,
		IdleTimeout:      90 * time.Second,
		MaxConnsPerHost:  10,
		MaxAttempts:      3,
		BackoffBaseMs:    500,
		BackoffJitterMs:  250,
		MaxDelayMs:       8000,
		QPS:              2.0,
		Burst:            4,
		CircuitWindow:    60 * time.Second,
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		UserAgent:        twseUserAgent,
		MaxBodyBytes:     0, // 0 = unlimited; TWSE responses are JSON envelopes, not HTML
	})
	// Pin User-Agent even if Config.UserAgent is overridden by a
	// future config-loader change. TWSEMiddleware with an empty UA
	// would be a no-op; we pass the canonical string explicitly.
	hc.Use(httpx.TWSEMiddleware(twseUserAgent))
	return twse.NewClient(hc)
}
