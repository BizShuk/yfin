// news.go — `NewsItem` type alias re-exported from model. The former
// `fromProtoNews` proto→SDK converter was removed with the ampy-proto
// dependency — scrape news items now convert to model directly via
// model.ScrapeNewsToItems. Struct lives in model/news.go; facade.NewsItem
// is a back-compat alias.
package facade

import (
	"github.com/bizshuk/yfin/model"
)

// NewsItem is one news article as exposed by the SDK facade. Aliased from
// model.NewsItem — new code should use model.NewsItem directly.
type NewsItem = model.NewsItem
