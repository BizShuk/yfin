// profile.go — Per-host configuration (`HostProfile`), registration, request
// defaults, and a generic endpoint tagger.
package httpx

import (
	"net/http"
	"net/url"
	"strings"
)

// HostProfile bundles every per-host knob the unified client needs.
// Instances are registered once at process start and looked up by host.
type HostProfile struct {
	Host       string
	EndpointFn func(path string) string
	QPS        float64
	Burst      int
	MaxBody    int64 // 0 means unlimited
	UserAgent  string
	Headers    http.Header
}

// UseProfile registers a host-specific profile. Invalid or empty hosts are
// ignored. Register profiles during process setup, before concurrent requests.
func (c *Client) UseProfile(profile HostProfile) {
	host := normalizeProfileHost(profile.Host)
	if host == "" {
		return
	}

	profile.Host = host
	profile.Headers = profile.Headers.Clone()
	c.profileMu.Lock()
	c.profiles[host] = profile
	c.profileMu.Unlock()
}

func (c *Client) profileForHost(host string) (HostProfile, bool) {
	host = normalizeProfileHost(host)
	if host == "" {
		return HostProfile{}, false
	}

	c.profileMu.RLock()
	profile, ok := c.profiles[host]
	c.profileMu.RUnlock()
	if !ok {
		return HostProfile{}, false
	}
	profile.Headers = profile.Headers.Clone()
	return profile, true
}

func (c *Client) applyProfileDefaults(req *http.Request, meta *Meta) (HostProfile, bool) {
	profile, ok := c.profileForHost(req.URL.Hostname())
	if !ok {
		return HostProfile{}, false
	}

	if profile.EndpointFn != nil {
		meta.Endpoint = profile.EndpointFn(req.URL.Path)
	}
	if profile.UserAgent != "" {
		req.Header.Set("User-Agent", profile.UserAgent)
	}
	for name, values := range profile.Headers {
		req.Header.Del(name)
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	return profile, true
}

func normalizeProfileHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" || strings.Contains(host, "://") {
		return ""
	}
	u, err := url.Parse("//" + host)
	if err != nil {
		return ""
	}
	hostname := strings.TrimSuffix(u.Hostname(), ".")
	return strings.ToLower(hostname)
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
