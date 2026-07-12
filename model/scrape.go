// scrape.go — `FetchMeta`, `ScrapeNewsItem` (raw scrape news article),
// `NewsStats` + scrape value types (`Scaled` alias to `ScaledDecimal`,
// `Currency`, `YahooNum`, `YahooInt`, `YahooString`) + scrape infrastructure
// (`RobotsRule`, `RobotsCache`, `BackoffPolicyConfig`, `RateLimitConfig`).
//
// Originally lived in svc/scrape/types.go + svc/scrape/types_json.go (value
// types). Promoted to model/ so cmd and external consumers can reference
// scrape DTOs without pulling in svc/scrape.
//
// Naming note: `ScrapeNewsItem` (raw Yahoo parse shape) is distinct from
// `model.NewsItem` (the cleaned-up SDK DTO from facade/news.go). The former
// carries `ImageURL` + `RelatedTickers` (raw scrape fields); the latter
// surfaces only `Title/URL/Source/Summary/PublishedAt/Symbols` for downstream
// consumers.

package model

import "time"

// FetchMeta contains metadata about a fetch operation
type FetchMeta struct {
	URL          string        `json:"url"`
	Host         string        `json:"host"`
	Status       int           `json:"status"`
	Attempt      int           `json:"attempt"`
	Bytes        int           `json:"bytes"`
	Gzip         bool          `json:"gzip"`
	Redirects    int           `json:"redirects"`
	Duration     time.Duration `json:"duration"`
	FromCache    bool          `json:"from_cache"`
	RobotsPolicy string        `json:"robots_policy"`
}

// ScrapeNewsItem represents a single news article extracted from Yahoo Finance
// (raw parse shape; consumers should normally use model.NewsItem instead).
type ScrapeNewsItem struct {
	Title          string     `json:"title"`
	URL            string     `json:"url"`
	Source         string     `json:"source"`
	PublishedAt    *time.Time `json:"published_at"`
	ImageURL       string     `json:"image_url"`
	RelatedTickers []string   `json:"related_tickers"`
}

// NewsStats represents statistics about news extraction
type NewsStats struct {
	TotalFound    int       `json:"total_found"`
	TotalReturned int       `json:"total_returned"`
	Deduped       int       `json:"deduped"`
	NextPageHint  string    `json:"next_page_hint"`
	AsOf          time.Time `json:"as_of"`
}

// Scaled alias to model.ScaledDecimal — kept for back-compat with scrape code
// that imported the scrape-local Scaled struct. New code should use
// model.ScaledDecimal directly.
type Scaled = ScaledDecimal

// Currency represents an ISO-4217 currency code
type Currency = string

// YahooNum represents Yahoo's numeric format with raw, fmt, and longFmt
type YahooNum struct {
	Raw     *float64 `json:"raw,omitempty"`
	Fmt     string   `json:"fmt,omitempty"`
	LongFmt string   `json:"longFmt,omitempty"`
}

// YahooInt represents Yahoo's integer format with raw, fmt, and longFmt
type YahooInt struct {
	Raw     *int64  `json:"raw,omitempty"`
	Fmt     string `json:"fmt,omitempty"`
	LongFmt string `json:"longFmt,omitempty"`
}

// YahooString represents Yahoo's string format that might contain numbers
type YahooString struct {
	Raw     *string `json:"raw,omitempty"`
	Fmt     string  `json:"fmt,omitempty"`
	LongFmt string  `json:"longFmt,omitempty"`
}

// RobotsRule represents a robots.txt rule
type RobotsRule struct {
	UserAgent string
	Allow     []string
	Disallow  []string
}

// RobotsCache represents cached robots.txt data
type RobotsCache struct {
	Host      string
	Rules     []RobotsRule
	FetchedAt time.Time
	TTL       time.Duration
}

// BackoffPolicyConfig represents backoff configuration
type BackoffPolicyConfig struct {
	BaseDelay    time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	JitterFactor float64
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	QPS            float64
	Burst          int
	PerHostWorkers int
}

// IsExpired checks if the robots cache is expired
func (rc *RobotsCache) IsExpired() bool {
	return time.Since(rc.FetchedAt) > rc.TTL
}

// Scaled.Format / Float64 / String — methods previously defined on scrape.Scaled
// now resolve to ScaledDecimal via the type alias. They are defined on
// ScaledDecimal above; nothing extra needed here.

// ToYahooNum converts a raw struct to YahooNum
func ToYahooNum(raw *float64, fmt, longFmt string) YahooNum {
	return YahooNum{
		Raw:     raw,
		Fmt:     fmt,
		LongFmt: longFmt,
	}
}

// ToYahooInt converts a raw struct to YahooInt
func ToYahooInt(raw *int64, fmt, longFmt string) YahooInt {
	return YahooInt{
		Raw:     raw,
		Fmt:     fmt,
		LongFmt: longFmt,
	}
}