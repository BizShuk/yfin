// news.go — `NewsItem` compatibility alias re-exported from model.
package facade

import (
	"github.com/bizshuk/yfin/model"
)

// NewsItem is one news article as exposed by the SDK facade. Aliased from
// model.NewsItem — new code should use model.NewsItem directly.
type NewsItem = model.NewsItem
