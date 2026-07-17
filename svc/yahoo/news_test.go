// Tests the yfinance-compatible Yahoo news XHR request and model conversion.
package yahoo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

func TestFetchNewsRequestAndDecode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/xhr/ncp", r.URL.Path)
		require.Equal(t, "latestNews", r.URL.Query().Get("queryRef"))
		require.Equal(t, "ncp_fin", r.URL.Query().Get("serviceKey"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload struct {
			ServiceConfig struct {
				SnippetCount int      `json:"snippetCount"`
				Symbols      []string `json:"s"`
			} `json:"serviceConfig"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, 10, payload.ServiceConfig.SnippetCount)
		require.Equal(t, []string{"AAPL"}, payload.ServiceConfig.Symbols)

		_, _ = w.Write([]byte(`{
          "data": {"tickerStream": {"stream": [
            {"id":"one","content":{
              "title":"Canonical article","summary":"Summary one","pubDate":"2026-07-16T10:00:00Z",
              "provider":{"displayName":"Yahoo Finance"},
              "canonicalUrl":{"url":"https://finance.yahoo.com/news/one"},
              "clickThroughUrl":{"url":"https://example.com/fallback-one"}
            }},
            {"id":"two","content":{
              "title":"Fallback article","summary":"Summary two","pubDate":"not-a-date",
              "provider":{"displayName":"Reuters"},
              "clickThroughUrl":{"url":"https://example.com/two"}
            }},
            {"id":"ad","ad":[{"campaign":"paid"}],"content":{
              "title":"Advertisement","canonicalUrl":{"url":"https://example.com/ad"}
            }},
            {"id":"invalid","content":{"title":"Missing URL"}}
          ]}}
        }`))
	}))
	defer server.Close()

	client := NewClient(httpx.NewClient(httpx.DefaultConfig()), "")
	client.newsBaseURL = server.URL
	items, err := client.FetchNews(context.Background(), "AAPL")
	require.NoError(t, err)
	require.Len(t, items, 2)

	require.Equal(t, "Canonical article", items[0].Title)
	require.Equal(t, "https://finance.yahoo.com/news/one", items[0].URL)
	require.Equal(t, "Yahoo Finance", items[0].Source)
	require.Equal(t, "Summary one", items[0].Summary)
	require.Equal(t, time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC), items[0].PublishedAt)
	require.Equal(t, []string{"AAPL"}, items[0].Symbols)

	require.Equal(t, "https://example.com/two", items[1].URL)
	require.Equal(t, "Reuters", items[1].Source)
	require.True(t, items[1].PublishedAt.IsZero())
}

func TestDecodeNewsEmptyStream(t *testing.T) {
	items, err := decodeNews([]byte(`{"data":{"tickerStream":{"stream":[]}}}`), "AAPL")
	require.NoError(t, err)
	require.Empty(t, items)
}
