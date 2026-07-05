// bwibbu_d.go — `BWIBBU_d` (/afterTrading/BWIBBU_d) per-stock P/E ratio, dividend yield %, and price-to-book ratio snapshot. Capacity: ~1900 stocks per response.
package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// BWIBBU_dResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type BWIBBU_dResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *BWIBBU_dResponse) GetStat() string { return r.Response.Stat }

// BWIBBUdRow is a typed representation of one BWIBBU_d data row.
// Fields: 證券代號, 證券名稱, 本益比, 殖利率(%), 股價淨值比.
type BWIBBUdRow struct {
	Code     string  // 證券代號
	Name     string  // 證券名稱
	PE       float64 // 本益比
	YieldPct float64 // 殖利率(%)
	PBR      float64 // 股價淨值比
}

// FetchBWIBBU_d retrieves the per-stock P/E, dividend yield, and P/B
// ratio snapshot for `date`. `opts` may include `selectType=ALL`
// (TWSE expects this).
func FetchBWIBBU_d(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BWIBBU_d: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("selectType", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[BWIBBU_dResponse](ctx, client, "/afterTrading/BWIBBU_d", q)
}

// ParseBWIBBUdRow converts one raw `data` row into a typed BWIBBUdRow.
func ParseBWIBBUdRow(row []string) (BWIBBUdRow, error) {
	if len(row) < 5 {
		return BWIBBUdRow{}, fmt.Errorf("BWIBBU_d: row too short: %d cols", len(row))
	}
	return BWIBBUdRow{
		Code:     strings.TrimSpace(row[0]),
		Name:     strings.TrimSpace(row[1]),
		PE:       ParseFloat(row[2]),
		YieldPct: ParsePercent(row[3]),
		PBR:      ParseFloat(row[4]),
	}, nil
}
