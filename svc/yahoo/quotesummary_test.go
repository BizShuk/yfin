// Tests `FetchQuoteSummary` against an `httptest` server: verifies the URL carries both `crumb=<value>` and a comma-joined `modules=` parameter.
package yahoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

func TestFetchQuoteSummary_InjectsCrumbAndModules(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			http.SetCookie(w, &http.Cookie{Name: "A1", Value: "tok", Path: "/"})
		case "/v1/test/getcrumb":
			_, _ = w.Write([]byte("CR"))
		default:
			gotURL = r.URL.String()
			_, _ = w.Write([]byte(`{"quoteSummary":{"result":[{}],"error":null}}`))
		}
	}))
	defer srv.Close()

	hc := httpx.NewClient(httpx.DefaultConfig())
	cm := NewCrumbManager(hc, srv.URL, srv.URL)
	c := NewClientWithAuth(hc, srv.URL, cm)

	raw, err := c.FetchQuoteSummary(context.Background(), "AAPL", []string{"esgScores", "secFilings"})
	require.NoError(t, err)
	require.Contains(t, string(raw), "quoteSummary")
	require.Contains(t, gotURL, "crumb=CR")
	require.True(t, strings.Contains(gotURL, "modules=esgScores%2CsecFilings") ||
		strings.Contains(gotURL, "modules=esgScores,secFilings"))
}

func TestFetchQuoteSummaryRotatesCrumbOnceOn401(t *testing.T) {
	var crumbCalls int
	var quoteCalls int
	var crumbs []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			http.SetCookie(w, &http.Cookie{Name: "A1", Value: "tok", Path: "/"})
		case "/v1/test/getcrumb":
			crumbCalls++
			_, _ = w.Write([]byte([]string{"CR1", "CR2"}[crumbCalls-1]))
		case "/v10/finance/quoteSummary/AAPL":
			quoteCalls++
			crumbs = append(crumbs, r.URL.Query().Get("crumb"))
			if quoteCalls == 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"quoteSummary":{"result":[{}],"error":null}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := httpx.DefaultConfig()
	cfg.MaxAttempts = 1
	hc := httpx.NewClient(cfg)
	client := NewClientWithAuth(hc, srv.URL, NewCrumbManager(hc, srv.URL, srv.URL))

	_, err := client.FetchQuoteSummary(context.Background(), "AAPL", []string{"assetProfile"})
	require.NoError(t, err)
	require.Equal(t, 2, crumbCalls)
	require.Equal(t, 2, quoteCalls)
	require.Equal(t, []string{"CR1", "CR2"}, crumbs)
}

func TestNewClientWiresDefaultCrumbManager(t *testing.T) {
	client := NewClient(httpx.NewClient(httpx.DefaultConfig()), "https://example.test")
	require.NotNil(t, client.crumb)
}
