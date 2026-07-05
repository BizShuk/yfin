// news.go — `NewsItem` plain SDK struct (Title / URL / Source / Summary /
// PublishedAt / Symbols) + unexported `fromProtoNews` proto → SDK converter
// (drops proto-only fields like Id / SentimentScoreBp / IngestTime / Meta to
// keep the public surface minimal). Capacity: 1 struct + 1 internal converter.
package facade

import (
	"time"

	newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"
)

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

// fromProtoNews converts a slice of ampy-proto NewsItem pointers into
// plain SDK NewsItem values. Nil entries in the input slice are skipped so
// the output slice stays dense and free of zero-value artifacts.
//
// Unexported because Step 6 only routes scrape → emit → proto → SDK output
// through facade.Client; external consumers should use the SDK structs
// without seeing the proto shape.
func fromProtoNews(items []*newsv1.NewsItem) []NewsItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]NewsItem, 0, len(items))
	for _, ni := range items {
		if ni == nil {
			continue
		}
		row := NewsItem{
			Title:   ni.GetHeadline(),
			URL:     ni.GetUrl(),
			Source:  ni.GetSource(),
			Summary: ni.GetBody(),
			Symbols: ni.GetTickers(),
		}
		if ts := ni.GetPublishedAt(); ts != nil {
			row.PublishedAt = ts.AsTime().UTC()
		}
		out = append(out, row)
	}
	return out
}