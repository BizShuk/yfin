// extract_news.go — ParseNews orchestration (JSON-first, HTML-regex
// fallback) plus the HTML-regex strategy itself. Regex config lives in
// news_regex_config.go, metrics in news_metrics.go, value normalization and
// dedup in news_normalize.go, and the embedded-JSON strategy in
// extract_news_json.go. Capacity: 25 articles max.
package scrape

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
	"github.com/bizshuk/yfin/model"
)

// ParseNews extracts news articles from HTML with robust error handling and deduplication
func ParseNews(html []byte, baseURL string, now time.Time) ([]model.ScrapeNewsItem, *model.NewsStats, error) {
	start := time.Now()

	// Initialize metrics
	metrics := newNewsMetrics()
	defer func() {
		metrics.recordNewsParseLatency(time.Since(start))
	}()

	htmlStr := string(html)

	// Try JSON-based extraction first (for real Yahoo Finance pages)
	articles, err := extractNewsFromJSON(htmlStr, baseURL, now)
	if err == nil && len(articles) > 0 {
		// JSON extraction successful
		originalCount := len(articles)
		articles = deduplicateArticles(articles)
		deduped := originalCount - len(articles)

		// Limit results (default 25 articles)
		const maxArticles = 25
		if len(articles) > maxArticles {
			articles = articles[:maxArticles]
		}

		// Extract pagination hint
		nextPageHint := extractNextPageHint(htmlStr)

		// Create statistics
		stats := &model.NewsStats{
			TotalFound:    originalCount,
			TotalReturned: len(articles),
			Deduped:       deduped,
			NextPageHint:  nextPageHint,
			AsOf:          now.UTC(),
		}

		metrics.recordNews("success")
		return articles, stats, nil
	}

	// Fall back to HTML-based extraction (for test fixtures or other formats)
	return parseNewsFromHTML(htmlStr, baseURL, now, metrics)
}

// extractArticleContainers finds all article containers in the HTML
func extractArticleContainers(html string) ([]string, error) {
	re, err := regexp.Compile(newsRegexConfig.ArticleContainer)
	if err != nil {
		return nil, fmt.Errorf("invalid article container regex: %w", err)
	}

	matches := re.FindAllStringSubmatch(html, -1)
	var containers []string

	for _, match := range matches {
		if len(match) > 1 {
			containers = append(containers, match[1])
		}
	}

	return containers, nil
}

// parseArticleFromContainer extracts article data from a single container
func parseArticleFromContainer(container, baseURL string, now time.Time) *model.ScrapeNewsItem {
	article := &model.ScrapeNewsItem{}

	// Extract title
	title := extractStringFromContainer(container, newsRegexConfig.Title)
	if title == "" {
		return nil // Skip articles without title
	}
	article.Title = html.UnescapeString(strings.TrimSpace(title))

	// Extract URL
	articleURL := extractStringFromContainer(container, newsRegexConfig.ArticleLink)
	if articleURL == "" {
		return nil // Skip articles without URL
	}
	article.URL = normalizeURL(articleURL, baseURL)

	// Extract publishing info (source and time)
	publishingInfo := extractStringFromContainer(container, newsRegexConfig.PublishingInfo)
	if publishingInfo != "" {
		source, publishedAt := parsePublishingInfo(publishingInfo, now)
		article.Source = source
		article.PublishedAt = publishedAt
	}

	// Extract image URL (optional)
	imageURL := extractStringFromContainer(container, newsRegexConfig.ImageURL)
	if imageURL != "" {
		article.ImageURL = imageURL
	}

	// Extract related tickers
	article.RelatedTickers = extractRelatedTickers(container)

	return article
}

// extractStringFromContainer extracts a string using regex from a container
func extractStringFromContainer(container, pattern string) string {
	if pattern == "" {
		return ""
	}

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(container)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// extractNextPageHint looks for pagination controls
func extractNextPageHint(html string) string {
	if newsRegexConfig.NextPageHint == "" {
		return ""
	}

	re := regexp.MustCompile(newsRegexConfig.NextPageHint)
	matches := re.FindStringSubmatch(html)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// parseNewsFromHTML falls back to HTML-based extraction for test fixtures
func parseNewsFromHTML(htmlStr, baseURL string, now time.Time, metrics *newsMetrics) ([]model.ScrapeNewsItem, *model.NewsStats, error) {
	// Load regex configuration
	if err := LoadNewsRegexConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to load news regex config: %w", err)
	}

	// Extract article containers
	containers, err := extractArticleContainers(htmlStr)
	if err != nil {
		metrics.recordNews("error")
		return nil, nil, fmt.Errorf("%w: %v", ErrNewsParse, err)
	}

	if len(containers) == 0 {
		metrics.recordNews("no_articles")
		return nil, nil, ErrNewsNoArticles
	}

	// Parse articles from containers
	var articles []model.ScrapeNewsItem
	for _, container := range containers {
		article := parseArticleFromContainer(container, baseURL, now)
		if article != nil {
			articles = append(articles, *article)
		}
	}

	// Enrich articles with source and published time from embedded JSON (if available)
	enrichArticlesWithJSONMeta(htmlStr, articles)

	if len(articles) == 0 {
		metrics.recordNews("no_valid_articles")
		return nil, nil, ErrNewsNoArticles
	}

	// Deduplicate articles
	originalCount := len(articles)
	articles = deduplicateArticles(articles)
	deduped := originalCount - len(articles)

	// Limit results (default 25 articles)
	const maxArticles = 25
	if len(articles) > maxArticles {
		articles = articles[:maxArticles]
	}

	// Extract pagination hint
	nextPageHint := extractNextPageHint(htmlStr)

	// Create statistics
	stats := &model.NewsStats{
		TotalFound:    originalCount,
		TotalReturned: len(articles),
		Deduped:       deduped,
		NextPageHint:  nextPageHint,
		AsOf:          now.UTC(),
	}

	metrics.recordNews("success")
	return articles, stats, nil
}
