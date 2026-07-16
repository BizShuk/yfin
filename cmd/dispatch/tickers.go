package dispatch

import (
	"bytes"
	_ "embed"

	"github.com/bizshuk/yfin/utils/cache"
)

//go:embed ticker_list.csv
var tickerListCSV []byte

func readEmbeddedTickerList() ([]string, error) {
	return cache.ReadTickerList(bytes.NewReader(tickerListCSV))
}
