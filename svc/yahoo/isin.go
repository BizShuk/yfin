// Fetches and parses a security's ISIN.

package yahoo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const isinSearchURL = "https://markets.businessinsider.com/ajax/SearchController_Suggest"

func parseISIN(body, ticker string) (string, error) {
	up := strings.ToUpper(ticker)
	for _, line := range strings.Split(body, "\n") {
		// fields are pipe-separated: SYMBOL|ISIN|NAME (possibly preceded by an index + tab)
		parts := strings.Split(line, "|")
		if len(parts) >= 2 && strings.Contains(strings.ToUpper(parts[0]), up) {
			return strings.TrimSpace(parts[1]), nil
		}
	}
	return "", fmt.Errorf("isin not found for %s", ticker)
}

// FetchISIN looks up the ISIN for a symbol via business-insider.
func (c *Client) FetchISIN(ctx context.Context, symbol string) (string, error) {
	// strip exchange suffix (2330.TW -> 2330)
	q := symbol
	if i := strings.Index(symbol, "."); i > 0 {
		q = symbol[:i]
	}
	u := isinSearchURL + "?max_results=25&query=" + url.QueryEscape(q)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return parseISIN(string(body), q)
}
