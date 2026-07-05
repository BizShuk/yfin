// fetch_test.go — `FetchJSON` happy-path decode + `ErrNoData` on no-data stat
// + embedded `Response.GetStat` exposure + a stubCaller verifying the
// injected transport contract. Capacity: ~5 test cases via stub or httptest.
package twse

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestHttpxClient points the test client at a local httptest server.
// It builds a real `*httpx.Client` whose BaseURL is empty (so the
// caller's `c.baseURL+path` is the absolute URL) and wraps it in a
// `*twse.Client` via NewClientWithURL.
func newTestHttpxClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	hc := httpx.NewClient(&httpx.Config{
		BaseURL:          "",
		Timeout:          5 * time.Second,
		MaxAttempts:      1,
		BackoffBaseMs:    1,
		BackoffJitterMs:  0,
		MaxDelayMs:       10,
		QPS:              1000,
		Burst:            1000,
		CircuitWindow:    time.Second,
		FailureThreshold: 1000,
		ResetTimeout:     time.Second,
	})
	return NewClientWithURL(hc, srv.URL)
}

func TestFetchJSON_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"OK","title":"MI_INDEX","fields":["a","b"],"data":[["1","2"],["3","4"]]}`))
	}))
	defer srv.Close()

	client := newTestHttpxClient(t, srv)
	got, err := FetchJSON[TestResponse](context.Background(), client, "/test/endpoint", nil)
	require.NoError(t, err)
	require.Equal(t, "OK", got.Stat)
	require.Equal(t, "MI_INDEX", got.Title)
	require.Len(t, got.Data, 2)
}

func TestFetchJSON_NoDataReturnsErrNoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stat":"沒有符合條件的資料","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestHttpxClient(t, srv)
	_, err := FetchJSON[TestResponse](context.Background(), client, "/test/endpoint", nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoData))
}

func TestFetchJSON_StatAtTopLevel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stat":"OK","data":[]}`))
	}))
	defer srv.Close()

	client := newTestHttpxClient(t, srv)
	got, err := FetchJSON[TestResponse](context.Background(), client, "/test/endpoint", nil)
	require.NoError(t, err)
	require.Equal(t, "OK", got.Stat)
}

type EmbeddedResponse struct {
	Response
	Date string `json:"date"`
}

func (r *EmbeddedResponse) GetStat() string { return r.Response.Stat }

func TestFetchJSON_EmbeddedStructReportsStat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stat":"OK","date":"20221230","fields":["a"],"data":[["x"]]}`))
	}))
	defer srv.Close()

	client := newTestHttpxClient(t, srv)
	got, err := FetchJSON[EmbeddedResponse](context.Background(), client, "/test/endpoint", nil)
	require.NoError(t, err)
	require.Equal(t, "OK", got.GetStat())
	require.Equal(t, "20221230", got.Date)
}

// stubCaller captures the last (path, query) pair fed to it and returns
// a canned body, bypassing the network. It satisfies httpx.Caller.
type stubCaller struct {
	lastPath  string
	lastQuery url.Values
	body      []byte
	meta      *httpx.Meta
	err       error
}

func (s *stubCaller) Get(ctx context.Context, path string, q url.Values) ([]byte, *httpx.Meta, error) {
	s.lastPath = path
	s.lastQuery = q
	if s.meta == nil {
		s.meta = &httpx.Meta{Status: 200, Host: "www.twse.com.tw"}
	}
	return s.body, s.meta, s.err
}

// TestFetchJSON_UsesInjectedCaller verifies the new Client-based
// FetchJSON plumbs through the injected httpx.Caller without going near
// the network and writes the right (path, query) into it.
func TestFetchJSON_UsesInjectedCaller(t *testing.T) {
	stub := &stubCaller{
		body: []byte(`{"stat":"OK","data":[],"title":"MI_INDEX"}`),
	}
	client := NewClientWithURL(stub, "https://www.twse.com.tw/rwd/zh")
	_, err := FetchJSON[MI_INDEXResponse](
		context.Background(),
		client,
		"/afterTrading/STOCK_DAY",
		url.Values{"stockNo": {"2330"}},
	)
	require.NoError(t, err)
	assert.Equal(t, "https://www.twse.com.tw/rwd/zh/afterTrading/STOCK_DAY", stub.lastPath)
	assert.Equal(t, "2330", stub.lastQuery.Get("stockNo"))
	assert.Equal(t, "json", stub.lastQuery.Get("response"))
}

// TestFetchJSON_Non2xxReturnsError covers the unhappy path: a 5xx response
// from the injected caller must surface as a wrapped error. The real
// httpx.Caller.Get returns (nil, meta, error) on non-2xx, so the stub
// mirrors that contract.
func TestFetchJSON_Non2xxReturnsError(t *testing.T) {
	stub := &stubCaller{
		err: fmt.Errorf("httpx: status 503: <html>oops</html>"),
	}
	client := NewClientWithURL(stub, "https://www.twse.com.tw/rwd/zh")
	_, err := FetchJSON[TestResponse](context.Background(), client, "/x", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
	assert.Contains(t, err.Error(), "status 503")
}

// TestClient_NewClientCapturesPackageBaseURL ensures the package-level
// BaseURL var is the default when NewClient is called (no test override
// in scope).
func TestClient_NewClientCapturesPackageBaseURL(t *testing.T) {
	stub := &stubCaller{}
	client := NewClient(stub)
	require.Equal(t, BaseURL, client.BaseURL())
	require.Same(t, httpx.Caller(stub), client.Caller())
}

// TestResponse is a sample struct matching the TWSE JSON envelope.
type TestResponse struct {
	Stat   string     `json:"stat"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

func (r *TestResponse) GetStat() string { return r.Stat }
