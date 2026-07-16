// — `ReadTickerList` loads a CSV ticker universe from an io.Reader. Capacity: 1 ticker per row, header row skipped, blank ticker values tolerated.

package cache

import (
	"encoding/csv"
	"io"
	"strings"
)

// ReadTickerList reads a CSV file where the ticker is the last comma-separated field
// of each non-header, non-empty line. Mirrors the yfinance script convention.
func ReadTickerList(src io.Reader) ([]string, error) {
	reader := csv.NewReader(src)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	var tickers []string
	for index, record := range records {
		if index == 0 {
			continue
		}
		if len(record) == 0 {
			continue
		}
		t := strings.TrimSpace(record[len(record)-1])
		if t == "" {
			continue
		}
		tickers = append(tickers, t)
	}
	return tickers, nil
}
