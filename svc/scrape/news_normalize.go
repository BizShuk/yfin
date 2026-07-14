// news_normalize.go — value normalization and deduplication for scraped
// news: URL canonicalization + tracking-param stripping, publishing-info
// and relative-time parsing, related-ticker extraction, and article dedup.
// Split out of extract_news.go: these are pure value transforms, invoked by
// both the HTML and the embedded-JSON extraction strategies.
package scrape

import (
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"github.com/bizshuk/yfin/model"
)

// normalizeURL converts relative URLs to absolute and cleans tracking parameters
func normalizeURL(articleURL, baseURL string) string {
	// Make URL absolute
	if !strings.HasPrefix(articleURL, "http") {
		if u, err := url.Parse(baseURL); err == nil {
			if parsed, err := url.Parse(articleURL); err == nil {
				articleURL = u.ResolveReference(parsed).String()
			}
		}
	}

	// Clean tracking parameters
	articleURL = cleanTrackingParams(articleURL)

	return articleURL
}

// cleanTrackingParams removes UTM and other tracking parameters
func cleanTrackingParams(urlStr string) string {
	patterns := []string{
		newsRegexConfig.URLCleanup.UTMParams,
		newsRegexConfig.URLCleanup.TrackingParams,
	}

	for _, pattern := range patterns {
		if pattern != "" {
			re := regexp.MustCompile(pattern)
			urlStr = re.ReplaceAllString(urlStr, "")
		}
	}

	// Clean up any remaining & at the end or beginning of query string
	urlStr = regexp.MustCompile(`[?&]+$`).ReplaceAllString(urlStr, "")
	urlStr = regexp.MustCompile(`\?&`).ReplaceAllString(urlStr, "?")

	return urlStr
}

// parsePublishingInfo extracts source and published time from publishing info
func parsePublishingInfo(info string, now time.Time) (string, *time.Time) {
	// Split on bullet point or similar separators
	parts := regexp.MustCompile(`\s*[•·|]\s*`).Split(info, -1)

	var source string
	var publishedAt *time.Time

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Try to parse as relative time
		if parsedTime := parseRelativeTime(part, now); parsedTime != nil {
			publishedAt = parsedTime
		} else {
			// Assume it's the source
			source = part
		}
	}

	return source, publishedAt
}

// parseRelativeTime converts relative time strings to absolute time
func parseRelativeTime(timeStr string, now time.Time) *time.Time {
	timeStr = strings.ToLower(strings.TrimSpace(timeStr))

	// Minutes ago
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Minutes); re != nil {
		if matches := re.FindStringSubmatch(timeStr); len(matches) > 1 {
			if minutes, err := strconv.Atoi(matches[1]); err == nil {
				result := now.Add(-time.Duration(minutes) * time.Minute).UTC()
				return &result
			}
		}
	}

	// Hours ago
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Hours); re != nil {
		if matches := re.FindStringSubmatch(timeStr); len(matches) > 1 {
			if hours, err := strconv.Atoi(matches[1]); err == nil {
				result := now.Add(-time.Duration(hours) * time.Hour).UTC()
				return &result
			}
		}
	}

	// Days ago
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Days); re != nil {
		if matches := re.FindStringSubmatch(timeStr); len(matches) > 1 {
			if days, err := strconv.Atoi(matches[1]); err == nil {
				result := now.Add(-time.Duration(days) * 24 * time.Hour).UTC()
				return &result
			}
		}
	}

	// Weeks ago
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Weeks); re != nil {
		if matches := re.FindStringSubmatch(timeStr); len(matches) > 1 {
			if weeks, err := strconv.Atoi(matches[1]); err == nil {
				result := now.Add(-time.Duration(weeks) * 7 * 24 * time.Hour).UTC()
				return &result
			}
		}
	}

	// Yesterday
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Yesterday); re != nil {
		if re.MatchString(timeStr) {
			// Set to start of yesterday (conservative approach)
			yesterday := now.Add(-24 * time.Hour)
			result := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
			return &result
		}
	}

	// Ensure no future times
	return nil
}

// extractRelatedTickers finds ticker symbols in the container
func extractRelatedTickers(container string) []string {
	if newsRegexConfig.RelatedTickers == "" {
		return nil
	}

	re := regexp.MustCompile(newsRegexConfig.RelatedTickers)
	matches := re.FindAllStringSubmatch(container, -1)

	var tickers []string
	tickerSet := make(map[string]bool) // For deduplication

	for _, match := range matches {
		if len(match) > 1 {
			ticker := strings.ToUpper(strings.TrimSpace(match[1]))
			// Validate ticker format (A-Z, 0-9, ., -)
			if isValidTicker(ticker) && !tickerSet[ticker] {
				tickers = append(tickers, ticker)
				tickerSet[ticker] = true
			}
		}
	}

	return tickers
}

// isValidTicker validates ticker symbol format
func isValidTicker(ticker string) bool {
	if len(ticker) == 0 || len(ticker) > 10 {
		return false
	}

	for _, char := range ticker {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '.' || char == '-') {
			return false
		}
	}

	return true
}

// deduplicateArticles removes duplicate articles using URL and content heuristics
func deduplicateArticles(articles []model.ScrapeNewsItem) []model.ScrapeNewsItem {
	seen := make(map[string]bool)
	var result []model.ScrapeNewsItem

	for _, article := range articles {
		// Primary dedup key: normalized URL
		normalizedURL := normalizeURLForDedup(article.URL)
		if seen[normalizedURL] {
			continue
		}

		// Secondary dedup: check for similar articles by title, source, and time
		isDuplicate := false
		titleNorm := strings.ToLower(strings.TrimSpace(article.Title))
		sourceNorm := strings.ToLower(strings.TrimSpace(article.Source))

		for _, existing := range result {
			existingTitleNorm := strings.ToLower(strings.TrimSpace(existing.Title))
			existingSourceNorm := strings.ToLower(strings.TrimSpace(existing.Source))

			// Check if title and source match
			if titleNorm == existingTitleNorm && sourceNorm == existingSourceNorm {
				// Check if times are within 2 minutes of each other
				if article.PublishedAt != nil && existing.PublishedAt != nil {
					timeDiff := article.PublishedAt.Sub(*existing.PublishedAt)
					if timeDiff < 0 {
						timeDiff = -timeDiff
					}
					if timeDiff <= 2*time.Minute {
						isDuplicate = true
						break
					}
				} else if article.PublishedAt == nil && existing.PublishedAt == nil {
					// Both have no timestamp, consider duplicate
					isDuplicate = true
					break
				}
			}
		}

		if isDuplicate {
			continue
		}

		seen[normalizedURL] = true
		result = append(result, article)
	}

	// Sort by published time (newest first)
	sort.Slice(result, func(i, j int) bool {
		if result[i].PublishedAt == nil && result[j].PublishedAt == nil {
			return false
		}
		if result[i].PublishedAt == nil {
			return false
		}
		if result[j].PublishedAt == nil {
			return true
		}
		return result[i].PublishedAt.After(*result[j].PublishedAt)
	})

	return result
}

// normalizeURLForDedup normalizes URL for deduplication
func normalizeURLForDedup(urlStr string) string {
	if u, err := url.Parse(urlStr); err == nil {
		// Lowercase host, remove query and fragment
		u.Host = strings.ToLower(u.Host)
		u.RawQuery = ""
		u.Fragment = ""
		return u.String()
	}
	return strings.ToLower(urlStr)
}
