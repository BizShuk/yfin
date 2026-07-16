// Tests `CrumbManager.Crumb` cache hit and `Invalidate` forced refetch against an `httptest` server that mimics the cookie + `/v1/test/getcrumb` endpoints.
package yahoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bizshuk/yfin/utils/httpx"
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

func TestCrumbManagerAcceptsForbiddenCookieBootstrap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			http.SetCookie(w, &http.Cookie{Name: "A1", Value: "tok", Path: "/"})
			http.Error(w, "forbidden", http.StatusForbidden)
		case "/v1/test/getcrumb":
			_, _ = w.Write([]byte("CRUMB"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := httpx.DefaultConfig()
	cfg.MaxAttempts = 1
	manager := NewCrumbManager(httpx.NewClient(cfg), srv.URL, srv.URL)

	crumb, err := manager.Crumb(context.Background())
	require.NoError(t, err)
	require.Equal(t, "CRUMB", crumb)
}
