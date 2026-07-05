// Parses Yahoo earnings-dates HTML.

package yahoo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// EarningsDateRow mirrors yfinance's earnings_dates DataFrame columns.
type EarningsDateRow struct {
	Date        string   `json:"date"`
	EPSEstimate *float64 `json:"eps_estimate,omitempty"`
	ReportedEPS *float64 `json:"reported_eps,omitempty"`
	SurprisePct *float64 `json:"surprise_pct,omitempty"`
}

const earningsCalendarURL = "https://finance.yahoo.com/calendar/earnings"

// ParseEarningsDatesHTML extracts earnings rows for `symbol` from the Yahoo
// calendar/earnings HTML page. It is the table-row parser only (no network).
func ParseEarningsDatesHTML(body []byte, symbol string) ([]EarningsDateRow, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	rows := collectRows(doc)
	if len(rows) < 2 { // header + at least one data row
		return nil, fmt.Errorf("earnings-dates: no rows found")
	}
	header := normalizeHeader(rows[0])
	idx := map[string]int{}
	for i, h := range header {
		idx[h] = i
	}
	out := []EarningsDateRow{}
	for _, r := range rows[1:] {
		if len(r) < len(header) {
			continue
		}
		// Filter by symbol (case-insensitive)
		if sym := strings.ToUpper(strings.TrimSpace(getCell(r, idx, "symbol"))); sym != "" && sym != strings.ToUpper(symbol) {
			continue
		}
		row := EarningsDateRow{
			Date:        strings.TrimSpace(getCell(r, idx, "date")),
			EPSEstimate: parseCellFloat(getCell(r, idx, "eps estimate")),
			ReportedEPS: parseCellFloat(getCell(r, idx, "reported eps")),
			SurprisePct: parseCellFloat(getCell(r, idx, "surprise(%)")),
		}
		out = append(out, row)
	}
	return out, nil
}

func collectRows(n *html.Node) [][]string {
	var rows [][]string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "tr") {
			var cells []string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (strings.EqualFold(c.Data, "td") || strings.EqualFold(c.Data, "th")) {
					cells = append(cells, extractText(c))
				}
			}
			if len(cells) > 0 {
				rows = append(rows, cells)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return rows
}

func extractText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(m *html.Node) {
		if m.Type == html.TextNode {
			sb.WriteString(m.Data)
		}
		for c := m.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(sb.String())
}

func normalizeHeader(row []string) []string {
	out := make([]string, len(row))
	for i, h := range row {
		out[i] = strings.ToLower(strings.TrimSpace(h))
	}
	return out
}

func getCell(row []string, idx map[string]int, key string) string {
	i, ok := idx[key]
	if !ok || i >= len(row) {
		return ""
	}
	return row[i]
}

func parseCellFloat(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" || s == "N/A" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

// FetchEarningsDates scrapes the Yahoo calendar/earnings HTML page for `symbol`.
func (c *Client) FetchEarningsDates(ctx context.Context, symbol string) ([]EarningsDateRow, error) {
	u := earningsCalendarURL + "?symbol=" + symbol
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseEarningsDatesHTML(body, symbol)
}
