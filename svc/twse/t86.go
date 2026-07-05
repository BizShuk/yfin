// t86.go — `T86` (/fund/T86) three-institution (foreign+mainland / trust / dealer) daily buy/sell/net volume per stock. Capacity: ~1900 stocks × 3 institutions per response.
package twse

import (
	"context"
	"fmt"
	"net/url"
)

// T86Response embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type T86Response struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *T86Response) GetStat() string { return r.Response.Stat }

// T86Row is a typed representation of one T86 data row.
type T86Row struct {
	Code        string // 證券代號
	Name        string // 證券名稱
	ForeignBuy  int64  // 外陸資買進股數
	ForeignSell int64  // 外陸資賣出股數
	ForeignNet  int64  // 外陸資買賣超股數
	TrustBuy    int64  // 投信買進股數
	TrustSell   int64  // 投信賣出股數
	TrustNet    int64  // 投信買賣超股數
	DealerBuy   int64  // 自營商買進股數
	DealerSell  int64  // 自營商賣出股數
	DealerNet   int64  // 自營商買賣超股數
	TotalNet    int64  // 三大法人買賣超股數
}

// FetchT86 retrieves the three-institution daily buy/sell for `date`.
// selectType=ALL is always added by this fetcher.
func FetchT86(ctx context.Context, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/T86: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("selectType", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[T86Response](ctx, "/fund/T86", q)
}

// ParseT86Row converts one raw `data` row into a typed T86Row.
func ParseT86Row(row []string) (T86Row, error) {
	if len(row) < 12 {
		return T86Row{}, fmt.Errorf("T86: row too short: %d cols", len(row))
	}
	return T86Row{
		Code:        row[0],
		Name:        row[1],
		ForeignBuy:  ParseInt(row[2]),
		ForeignSell: ParseInt(row[3]),
		ForeignNet:  ParseInt(row[4]),
		TrustBuy:    ParseInt(row[5]),
		TrustSell:   ParseInt(row[6]),
		TrustNet:    ParseInt(row[7]),
		DealerBuy:   ParseInt(row[8]),
		DealerSell:  ParseInt(row[9]),
		DealerNet:   ParseInt(row[10]),
		TotalNet:    ParseInt(row[11]),
	}, nil
}
