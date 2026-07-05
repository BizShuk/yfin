// middleware_test.go — Tests `RequestMiddleware` chain ordering, error propagation, and `ResponseMiddleware` firing on the success path, plus the three service-specific middleware factories (TWSE / Yahoo / Scrape). Capacity: 7 test functions.
package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubCrumbProvider is a tiny CrumbProvider used to verify YahooMiddleware
// without depending on svc/yahoo.
type stubCrumbProvider struct {
	crumb string
	err   error
}

func (s *stubCrumbProvider) Crumb(ctx context.Context) (string, error) {
	return s.crumb, s.err
}

func TestMiddleware_OrderingAndError(t *testing.T) {
	var order []string
	c := NewClient(nil)
	c.Use(func(req *http.Request, meta *Meta) error {
		order = append(order, "first")
		req.Header.Set("X-Step", "1")
		return nil
	})
	c.Use(func(req *http.Request, meta *Meta) error {
		order = append(order, "second")
		return errors.New("blocked")
	})

	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := c.Do(ctx, req); err == nil {
		t.Fatalf("expected error from middleware chain, got nil")
	} else if got := err.Error(); !strings.Contains(got, "blocked") {
		t.Errorf("expected error to mention %q, got %q", "blocked", got)
	}

	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Errorf("expected middleware order [first, second], got %v", order)
	}
}

func TestMiddleware_ResponseRunsOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Step"); got != "1" {
			t.Errorf("expected X-Step=1 from request middleware, got %q", got)
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 1
	cfg.BackoffBaseMs = 1
	cfg.BackoffJitterMs = 0

	var respFired bool
	c := NewClient(cfg)
	c.Use(func(req *http.Request, meta *Meta) error {
		req.Header.Set("X-Step", "1")
		return nil
	})
	c.UseAfter(func(resp *http.Response, meta *Meta) {
		respFired = true
		if meta.Host == "" {
			t.Errorf("expected meta.Host populated, got empty")
		}
	})

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()

	if !respFired {
		t.Errorf("expected response middleware to run on success path")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestMeta_HostAndEndpointPopulated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 1
	cfg.BackoffBaseMs = 1
	cfg.BackoffJitterMs = 0

	var seen *Meta
	c := NewClient(cfg)
	c.Use(func(req *http.Request, meta *Meta) error {
		seen = meta
		return nil
	})

	req, err := http.NewRequest(http.MethodGet, server.URL+"/v8/finance/chart/AAPL", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if _, err := c.Do(context.Background(), req); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if seen == nil {
		t.Fatalf("request middleware did not run")
	}
	if seen.Host == "" {
		t.Errorf("expected meta.Host populated from req.URL.Host, got empty")
	}
	if seen.Endpoint == "" {
		t.Errorf("expected meta.Endpoint populated, got empty")
	}
}

func TestCommonEndpointFn(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"/v8/finance/chart/AAPL", "v8"},
		{"/STOCK_DAY", "STOCK_DAY"},
		{"root/path", "root"},
		{"/", "root"},
		{"", "root"},
	}
	for _, tc := range cases {
		if got := CommonEndpointFn(tc.in); got != tc.want {
			t.Errorf("CommonEndpointFn(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTWSEMiddleware_SetsUserAgent(t *testing.T) {
	const ua = "TwseAgent/1.0"
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := TWSEMiddleware(ua)(req, &Meta{}); err != nil {
		t.Fatalf("middleware: %v", err)
	}
	if got := req.Header.Get("User-Agent"); got != ua {
		t.Errorf("User-Agent = %q, want %q", got, ua)
	}
}

func TestTWSEMiddleware_EmptyUANoOp(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("User-Agent", "preset")
	if err := TWSEMiddleware("")(req, &Meta{}); err != nil {
		t.Fatalf("middleware: %v", err)
	}
	if got := req.Header.Get("User-Agent"); got != "preset" {
		t.Errorf("empty UA overwrote preset, got %q", got)
	}
}

func TestYahooMiddleware_InjectsCrumbQuery(t *testing.T) {
	cm := &stubCrumbProvider{crumb: "abc123"}
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/v11/finance/quoteSummary/AAPL?modules=foo", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := YahooMiddleware(cm)(req, &Meta{}); err != nil {
		t.Fatalf("middleware: %v", err)
	}
	if got := req.URL.Query().Get("crumb"); got != "abc123" {
		t.Errorf("crumb = %q, want %q", got, "abc123")
	}
	if got := req.URL.Query().Get("modules"); got != "foo" {
		t.Errorf("pre-existing modules param lost, got %q", got)
	}
}

func TestYahooMiddleware_NilProviderNoOp(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := YahooMiddleware(nil)(req, &Meta{}); err != nil {
		t.Fatalf("middleware: %v", err)
	}
	if got := req.URL.Query().Get("crumb"); got != "" {
		t.Errorf("nil provider should not set crumb, got %q", got)
	}
}

func TestYahooMiddleware_PropagatesCrumbError(t *testing.T) {
	cm := &stubCrumbProvider{err: errors.New("consent flow required")}
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	err = YahooMiddleware(cm)(req, &Meta{})
	if err == nil || !strings.Contains(err.Error(), "consent flow required") {
		t.Errorf("expected wrapped error containing 'consent flow required', got %v", err)
	}
}

func TestScrapeMiddleware_SetsBrowserHeaders(t *testing.T) {
	const ua = "ScrapeAgent/1.0"
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := ScrapeMiddleware(ua)(req, &Meta{}); err != nil {
		t.Fatalf("middleware: %v", err)
	}
	if got := req.Header.Get("User-Agent"); got != ua {
		t.Errorf("User-Agent = %q, want %q", got, ua)
	}
	if got := req.Header.Get("Accept"); got == "" {
		t.Error("expected Accept header set")
	}
	if got := req.Header.Get("Accept-Language"); got == "" {
		t.Error("expected Accept-Language header set")
	}
}
