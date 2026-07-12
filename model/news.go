// news.go — `NewsItem` plain data struct for the SDK surface.
// Only the fields most consumers actually need (Title / URL / Source /
// Summary / PublishedAt / Symbols) are surfaced; the proto's Id /
// SentimentScoreBp / IngestTime / Meta are deliberately dropped.
// Originally lived in facade/news.go; promoted to model/ for cross-layer reuse.

package model

import "time"

// NewsItem is one news article as exposed by the SDK facade. Only the
// fields most consumers actually need are surfaced (Title / URL / Source /
// Summary / PublishedAt / Symbols); the proto's Id, SentimentScoreBp and
// IngestTime / Meta are deliberately dropped — callers that need them
// should reach into the proto via the internal package.
//
// PublishedAt is UTC; a nil proto Timestamp yields the zero time.Time.
type NewsItem struct {
	Title       string    `json:"title,omitempty"`
	URL         string    `json:"url,omitempty"`
	Source      string    `json:"source,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	Symbols     []string  `json:"symbols,omitempty"`
}
