// HTTPClient returns the process-wide shared stdlib HTTP client used by service packages.

package config

import (
	"net/http"
	"sync"
	"time"
)

// HTTPTimeout is the per-request timeout for the shared service HTTP client.
const HTTPTimeout = 30 * time.Second

var (
	httpOnce   sync.Once
	httpClient *http.Client
)

// HTTPClient returns the process-wide shared HTTP client used by service
// packages (svc/*) to reach external REST endpoints. It is created once and
// reused so the underlying transport's connection pool is shared across calls.
//
// The client is host-agnostic: it carries no base URL. Service packages own
// the full request URL (scheme+host+path) and build the *http.Request
// themselves, so a single shared client can serve every upstream.
func HTTPClient() *http.Client {
	httpOnce.Do(func() {
		httpClient = &http.Client{Timeout: HTTPTimeout}
	})
	return httpClient
}
