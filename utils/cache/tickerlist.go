// ReadTickerList loads a newline-delimited ticker list from disk.

package cache

import (
	"bufio"
	"os"
	"strings"
)

// ReadTickerList reads a CSV file where the ticker is the last comma-separated field
// of each non-header, non-empty line. Mirrors the yfinance script convention.
func ReadTickerList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tickers []string
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		if first { // skip header
			first = false
			continue
		}
		parts := strings.Split(strings.TrimSpace(sc.Text()), ",")
		if len(parts) == 0 {
			continue
		}
		t := strings.TrimSpace(parts[len(parts)-1])
		if t == "" {
			continue
		}
		tickers = append(tickers, t)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return tickers, nil
}
