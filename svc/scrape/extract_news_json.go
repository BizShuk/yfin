// extract_news_json.go — the embedded-JSON extraction strategy: pulls the
// tickerStream <script> payload out of the page and decodes articles from
// it. This is the preferred path; extract_news.go falls back to HTML regex
// when it yields nothing.
package scrape

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	bodyPatternRegex   = `"body":"(\{.*?\})"`
	scriptPatternRegex = `<script[^>]*>([^<]*tickerStream[^<]*)</script>`
)

// extractNewsFromJSON extracts news from JSON data embedded in script tags
func extractNewsFromJSON(html, baseURL string, now time.Time) ([]NewsItem, error) {
	// Look for the script tag containing the tickerStream data
	scriptPattern := scriptPatternRegex
	scriptRe := regexp.MustCompile(scriptPattern)
	scriptMatches := scriptRe.FindStringSubmatch(html)

	if len(scriptMatches) < 2 {
		return nil, fmt.Errorf("no tickerStream script found")
	}

	jsonContent := scriptMatches[1]

	// The JSON is nested - extract the body content which contains the actual news data
	bodyPattern := bodyPatternRegex
	bodyRe := regexp.MustCompile(bodyPattern)
	bodyMatches := bodyRe.FindStringSubmatch(jsonContent)

	if len(bodyMatches) < 2 {
		return nil, fmt.Errorf("no body content found in script")
	}

	// The body content is escaped JSON, so we need to unescape it using JSON decoder
	raw := bodyMatches[1]
	var unescaped string
	if err := json.Unmarshal([]byte("\""+raw+"\""), &unescaped); err != nil {
		// Fallback simple unescape
		unescaped = strings.ReplaceAll(raw, `\\`, `\`)
		unescaped = strings.ReplaceAll(unescaped, `\"`, `"`)
	}

	// Now extract individual articles from the content arrays
	return extractArticlesFromNewsJSON(unescaped, baseURL, now)
}

// parseTickersFromJSON extracts ticker symbols from JSON ticker array
func parseTickersFromJSON(tickersJSON string) []string {
	tickerPattern := `"symbol":"([^"]*)"`
	re := regexp.MustCompile(tickerPattern)
	matches := re.FindAllStringSubmatch(tickersJSON, -1)

	var tickers []string
	tickerSet := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			ticker := strings.ToUpper(strings.TrimSpace(match[1]))
			if isValidTicker(ticker) && !tickerSet[ticker] {
				tickers = append(tickers, ticker)
				tickerSet[ticker] = true
			}
		}
	}

	return tickers
}

// extractArticlesFromNewsJSON extracts articles from the news JSON structure
func extractArticlesFromNewsJSON(bodyJSON, baseURL string, now time.Time) ([]NewsItem, error) {
	// Find blocks with contentType STORY directly to avoid brittle array parsing
	storyBlock := regexp.MustCompile(`\{"id":"[^"]*","content":\{[^}]*"contentType":"STORY"[^}]*\}`)
	blocks := storyBlock.FindAllString(bodyJSON, -1)

	var allArticles []NewsItem
	for _, blk := range blocks {
		// Extract core fields
		title := extractFirstGroup(blk, `"title":"([^"]*)"`)
		url := extractFirstGroup(blk, `"canonicalUrl":\{[^}]*"url":"([^"]*)"`)
		source := extractFirstGroup(blk, `"provider":\{[^}]*"displayName":"([^"]*)"`)
		pub := extractFirstGroup(blk, `"pubDate":"([^"]*)"`)
		img := extractFirstGroup(blk, `"originalUrl":"([^"]*)"`)
		tickRaw := extractFirstGroup(blk, `"stockTickers":\[([^\]]*)\]`)

		if title == "" || url == "" {
			continue
		}

		item := NewsItem{Title: strings.TrimSpace(title), URL: strings.TrimSpace(url), Source: strings.TrimSpace(source), ImageURL: strings.TrimSpace(img)}
		if pub != "" {
			if t, err := time.Parse(time.RFC3339, pub); err == nil {
				tt := t.UTC()
				item.PublishedAt = &tt
			}
		}
		if tickRaw != "" {
			item.RelatedTickers = parseTickersFromJSON(tickRaw)
		}
		allArticles = append(allArticles, item)
	}
	return allArticles, nil
}

// findObjectEnd finds the end of a JSON object starting from a position

// extractFirstGroup is a tiny helper to extract first capturing group
func extractFirstGroup(s, pattern string) string {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// enrichArticlesWithJSONMeta builds a title->(source,time) map once and enriches articles in place
func enrichArticlesWithJSONMeta(fullHTML string, articles []NewsItem) {
	if len(articles) == 0 {
		return
	}
	scriptPattern := `<script[^>]*>([^<]*tickerStream[^<]*)</script>`
	scriptRe := regexp.MustCompile(scriptPattern)
	scriptMatches := scriptRe.FindStringSubmatch(fullHTML)
	if len(scriptMatches) < 2 {
		return
	}
	// Unescape JSON body
	bodyPattern := `"body":"(\{.*?\})"`
	bodyRe := regexp.MustCompile(bodyPattern)
	bodyMatches := bodyRe.FindStringSubmatch(scriptMatches[1])
	if len(bodyMatches) < 2 {
		return
	}
	var jsonBody string
	if err := json.Unmarshal([]byte("\""+bodyMatches[1]+"\""), &jsonBody); err != nil {
		jsonBody = bodyMatches[1]
	}

	// Build a map from normalized title to (source, pubDate)
	meta := make(map[string]struct {
		src string
		t   *time.Time
	})

	storyBlock := regexp.MustCompile(`\{"id":"[^"]*","content":\{[^}]*"contentType":"STORY"[^}]*\}`)
	blocks := storyBlock.FindAllString(jsonBody, -1)
	for _, blk := range blocks {
		title := extractFirstGroup(blk, `"title":"([^"]*)"`)
		if title == "" {
			continue
		}
		src := extractFirstGroup(blk, `"provider":\{[^}]*"displayName":"([^"]*)"`)
		pub := extractFirstGroup(blk, `"pubDate":"([^"]*)"`)
		var pt *time.Time
		if pub != "" {
			if t, err := time.Parse(time.RFC3339, pub); err == nil {
				tt := t.UTC()
				pt = &tt
			}
		}
		key := strings.ToLower(strings.TrimSpace(title))
		meta[key] = struct {
			src string
			t   *time.Time
		}{src: strings.TrimSpace(src), t: pt}
	}

	// Enrich
	for i := range articles {
		key := strings.ToLower(strings.TrimSpace(articles[i].Title))
		if m, ok := meta[key]; ok {
			if articles[i].Source == "" && m.src != "" {
				articles[i].Source = m.src
			}
			if articles[i].PublishedAt == nil && m.t != nil {
				articles[i].PublishedAt = m.t
			}
		}
	}
}

// enhanceArticleWithJSON attempts to fill missing fields from JSON data in the full HTML
