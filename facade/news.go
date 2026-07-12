// news.go — `NewsItem` type alias + unexported `fromProtoNews` proto → SDK
// converter (drops proto-only fields like Id / SentimentScoreBp / IngestTime /
// Meta to keep the public surface minimal). Struct lives in model/news.go;
// facade.NewsItem is a back-compat alias.
package facade

import (
	newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"

	"github.com/bizshuk/yfin/model"
)

// NewsItem is one news article as exposed by the SDK facade. Aliased from
// model.NewsItem — new code should use model.NewsItem directly.
type NewsItem = model.NewsItem

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
