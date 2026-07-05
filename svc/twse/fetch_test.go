// fetch_test.go — `FetchJSON` happy-path decode + `ErrNoData` on no-data stat + embedded `Response.GetStat` exposure. Capacity: ~4 test cases via httptest server.
package twse

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// newTestHttpxClient points the package BaseURL at the httptest server so
// FetchJSON resolves the path against it, restoring it on cleanup. No client
// is injected: FetchJSON uses the shared stdlib client from internal/config.
func newTestHttpxClient(t *testing.T, srv *httptest.Server) {
	t.Helper()
	old := BaseURL
	BaseURL = srv.URL
	t.Cleanup(func() { BaseURL = old })
}

func TestFetchJSON_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"OK","title":"MI_INDEX","fields":["a","b"],"data":[["1","2"],["3","4"]]}`))
	}))
	defer srv.Close()

	newTestHttpxClient(t, srv)
	got, err := FetchJSON[TestResponse](context.Background(), "/test/endpoint", nil)
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

	newTestHttpxClient(t, srv)
	_, err := FetchJSON[TestResponse](context.Background(), "/test/endpoint", nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoData))
}

func TestFetchJSON_StatAtTopLevel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stat":"OK","data":[]}`))
	}))
	defer srv.Close()

	newTestHttpxClient(t, srv)
	got, err := FetchJSON[TestResponse](context.Background(), "/test/endpoint", nil)
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

	newTestHttpxClient(t, srv)
	got, err := FetchJSON[EmbeddedResponse](context.Background(), "/test/endpoint", nil)
	require.NoError(t, err)
	require.Equal(t, "OK", got.GetStat())
	require.Equal(t, "20221230", got.Date)
}

// TestResponse is a sample struct matching the TWSE JSON envelope.
type TestResponse struct {
	Stat   string     `json:"stat"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

func (r *TestResponse) GetStat() string { return r.Stat }
