package httpx

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_PersistsCookiesAcrossRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/set" {
			http.SetCookie(w, &http.Cookie{Name: "A1", Value: "token123", Path: "/"})
			return
		}
		if c, err := r.Cookie("A1"); err == nil {
			_, _ = w.Write([]byte(c.Value))
		}
	}))
	defer srv.Close()

	c := NewClient(DefaultConfig())
	u, _ := url.Parse(srv.URL)

	req1, _ := http.NewRequest("GET", srv.URL+"/set", nil)
	resp1, err := c.Do(req1.Context(), req1)
	require.NoError(t, err)
	resp1.Body.Close()

	require.NotEmpty(t, c.Jar().Cookies(u))
}
