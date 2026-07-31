// news.go — `NewsItem` plain data struct for the SDK surface.

package model

import "time"

// NewsItem is one news article as exposed by the SDK facade. PublishedAt is
// UTC; an absent provider timestamp is represented by the zero time.Time.
type NewsItem struct {
	Title       string    `json:"title,omitempty"`
	URL         string    `json:"url,omitempty"`
	Source      string    `json:"source,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	Symbols     []string  `json:"symbols,omitempty"`
}
