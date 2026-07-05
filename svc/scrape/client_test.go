// client_test.go — Tests `Client.Fetch` delegates exactly once to `httpx.Caller.Get`, respects robots.txt policy, and propagates `*Meta` fields. Capacity: 4 test functions.
package scrape

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errStubCaller is the canned error used by stubCaller when its err
// field is non-nil. tests inject it to assert error-propagation paths.
var errStubCaller = errors.New("stub caller error")

// stubCaller is a minimal `httpx.Caller` for tests. It records the
// number of Get calls and the path/query arguments, returning a canned
// body + Meta.
type stubCaller struct {
	calls    int
	lastPath string
	lastQry  url.Values
	body     []byte
	meta     *httpx.Meta
	err      error
}

func (s *stubCaller) Get(_ context.Context, path string, q url.Values) ([]byte, *httpx.Meta, error) {
	s.calls++
	s.lastPath = path
	s.lastQry = q
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.body, s.meta, nil
}

// TestFetch_DelegatesToCallerOnce verifies the happy path: one robots
// check, one caller.Get, body + meta returned.
func TestFetch_DelegatesToCallerOnce(t *testing.T) {
	stub := &stubCaller{
		body: []byte("ok"),
		meta: &httpx.Meta{Status: 200, Host: "finance.yahoo.com", Endpoint: "quote"},
	}
	c, err := NewClientWithCaller(stub, DefaultConfig())
	require.NoError(t, err)

	body, fetchMeta, err := c.Fetch(context.Background(), "https://finance.yahoo.com/quote/AAPL")
	require.NoError(t, err)
	assert.Equal(t, []byte("ok"), body)
	assert.Equal(t, 1, stub.calls)
	assert.Equal(t, "/quote/AAPL", stub.lastPath)
	assert.NotNil(t, fetchMeta)
	assert.Equal(t, 200, fetchMeta.Status)
	assert.Equal(t, "finance.yahoo.com", fetchMeta.Host)
}

// TestFetch_PropagatesCallerError — Caller errors are surfaced.
func TestFetch_PropagatesCallerError(t *testing.T) {
	stub := &stubCaller{err: errStubCaller}
	c, err := NewClientWithCaller(stub, DefaultConfig())
	require.NoError(t, err)

	_, _, err = c.Fetch(context.Background(), "https://finance.yahoo.com/quote/AAPL")
	require.Error(t, err)
	assert.Equal(t, 1, stub.calls, "caller should still be invoked exactly once")
}

// TestFetch_RobotsDeniedDoesNotCallCaller — robots.txt deny short-circuits before caller.
func TestFetch_RobotsDeniedDoesNotCallCaller(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RobotsPolicy = string(RobotsEnforce)
	stub := &stubCaller{
		body: []byte("ok"),
		meta: &httpx.Meta{Status: 200, Host: "finance.yahoo.com"},
	}
	c, err := NewClientWithCaller(stub, cfg)
	require.NoError(t, err)

	// robots.txt enforcement + nonexistent host (robots fetch fails closed).
	_, _, err = c.Fetch(context.Background(), "https://this-host-does-not-exist-xyz.invalid/quote/AAPL")
	require.Error(t, err)
	assert.Equal(t, 0, stub.calls, "caller must not be invoked when robots.txt denies")
}

// TestFetch_RobotsIgnoreSkipsCaller — ignore policy still delegates.
func TestFetch_RobotsIgnoreSkipsCaller(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RobotsPolicy = string(RobotsIgnore)
	stub := &stubCaller{
		body: []byte("ok"),
		meta: &httpx.Meta{Status: 200, Host: "example.com"},
	}
	c, err := NewClientWithCaller(stub, cfg)
	require.NoError(t, err)

	// Path with disallowed chars but ignore policy => caller still called once.
	_, _, err = c.Fetch(context.Background(), "https://example.com/foo")
	// err may or may not be nil depending on caller setup; what we care about is calls.
	assert.Equal(t, 1, stub.calls, "caller should be invoked once when robots policy is ignore")
}

// TestNewClient_NilPoolFallbacksToFreshClient is a regression test for
// the typed-nil interface bug: when `pool` is nil, the deprecated
// NewClient must construct a fresh `*httpx.Client` rather than wrap the
// nil pointer in an `httpx.Caller` interface variable (which makes
// `caller == nil` falsely false and lets the nil escape into the
// returned Client, panicking on the first Fetch call when it
// dereferences `c.config.BaseURL` inside (*httpx.Client).Get).
func TestNewClient_NilPoolFallbacksToFreshClient(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RobotsPolicy = string(RobotsIgnore) // bypass robots.txt check
	c, err := NewClient(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, c)

	// Tight deadline so the test fails fast regardless of network.
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	var panicVal interface{}
	func() {
		defer func() { panicVal = recover() }()
		_, _, _ = c.Fetch(ctx, "https://finance.yahoo.com/quote/AAPL")
	}()
	assert.Nil(t, panicVal, "Fetch must not panic on a typed-nil caller; fresh-client fallback should run when pool is nil")
}