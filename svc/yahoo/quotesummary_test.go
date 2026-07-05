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
