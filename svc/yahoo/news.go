// Fetches Yahoo Finance's ticker news XHR and converts its content envelope to
// the stable model.NewsItem surface.
package yahoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bizshuk/yfin/model"
)

type newsRequest struct {
	ServiceConfig struct {
		SnippetCount int      `json:"snippetCount"`
		Symbols      []string `json:"s"`
	} `json:"serviceConfig"`
}

type newsEnvelope struct {
	Data struct {
		TickerStream struct {
			Stream []newsArticle `json:"stream"`
		} `json:"tickerStream"`
	} `json:"data"`
}

type newsArticle struct {
	Ad      json.RawMessage `json:"ad"`
	Content struct {
		Title    string `json:"title"`
		Summary  string `json:"summary"`
		PubDate  string `json:"pubDate"`
		Provider struct {
			DisplayName string `json:"displayName"`
		} `json:"provider"`
		CanonicalURL struct {
			URL string `json:"url"`
		} `json:"canonicalUrl"`
		ClickThroughURL struct {
			URL string `json:"url"`
		} `json:"clickThroughUrl"`
	} `json:"content"`
}

// FetchNews fetches the latest ten ticker news items from Yahoo Finance.
func (c *Client) FetchNews(ctx context.Context, symbol string) ([]model.NewsItem, error) {
	payload := newsRequest{}
	payload.ServiceConfig.SnippetCount = 10
	payload.ServiceConfig.Symbols = []string{symbol}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode news request: %w", err)
	}

	endpoint, err := url.Parse(strings.TrimRight(c.newsBaseURL, "/") + "/xhr/ncp")
	if err != nil {
		return nil, fmt.Errorf("build news URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("queryRef", "latestNews")
	query.Set("serviceKey", "ncp_fin")
	endpoint.RawQuery = query.Encode()

	ctx = circuitContext(ctx, circuitGroupNews)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create news request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch news: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read news response: %w", err)
	}
	return decodeNews(responseBody, symbol)
}

func decodeNews(data []byte, symbol string) ([]model.NewsItem, error) {
	var envelope newsEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode news response: %w", err)
	}

	items := make([]model.NewsItem, 0, len(envelope.Data.TickerStream.Stream))
	for _, article := range envelope.Data.TickerStream.Stream {
		if newsArticleHasAd(article.Ad) {
			continue
		}
		title := strings.TrimSpace(article.Content.Title)
		articleURL := strings.TrimSpace(article.Content.CanonicalURL.URL)
		if articleURL == "" {
			articleURL = strings.TrimSpace(article.Content.ClickThroughURL.URL)
		}
		if title == "" || articleURL == "" {
			continue
		}
		publishedAt, _ := time.Parse(time.RFC3339, article.Content.PubDate)
		items = append(items, model.NewsItem{
			Title:       title,
			URL:         articleURL,
			Source:      strings.TrimSpace(article.Content.Provider.DisplayName),
			Summary:     strings.TrimSpace(article.Content.Summary),
			PublishedAt: publishedAt,
			Symbols:     []string{symbol},
		})
	}
	return items, nil
}

func newsArticleHasAd(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed != "" && trimmed != "null" && trimmed != "[]" && trimmed != "{}"
}
