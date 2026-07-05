package yahoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bizshuk/yfinance-go/utils/httpx"
	"github.com/stretchr/testify/require"
)

func TestCrumbManager_FetchesAndCachesCrumb(t *testing.T) {
	var crumbCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			http.SetCookie(w, &http.Cookie{Name: "A1", Value: "tok", Path: "/"})
		case "/v1/test/getcrumb":
			crumbCalls++
			_, _ = w.Write([]byte("abc123CRUMB"))
		}
	}))
	defer srv.Close()

	cm := NewCrumbManager(httpx.NewClient(httpx.DefaultConfig()), srv.URL, srv.URL)
	got, err := cm.Crumb(context.Background())
	require.NoError(t, err)
	require.Equal(t, "abc123CRUMB", got)

	// second call should be cached — no extra getcrumb hit
	_, _ = cm.Crumb(context.Background())
	require.Equal(t, 1, crumbCalls)
}

func TestCrumbManager_InvalidateForcesRefetch(t *testing.T) {
	var crumbCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			http.SetCookie(w, &http.Cookie{Name: "A1", Value: "tok", Path: "/"})
		case "/v1/test/getcrumb":
			crumbCalls++
			_, _ = w.Write([]byte("CRUMB"))
		}
	}))
	defer srv.Close()

	cm := NewCrumbManager(httpx.NewClient(httpx.DefaultConfig()), srv.URL, srv.URL)
	_, _ = cm.Crumb(context.Background())
	cm.Invalidate()
	_, _ = cm.Crumb(context.Background())
	require.Equal(t, 2, crumbCalls)
}
