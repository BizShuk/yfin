package twse

import (
	"context"
	"fmt"
	"net/url"

	"github.com/bizshuk/yfin/model"
)

// FetchMI_QFIIS retrieves the foreign+mainland investor holdings for `date`.
// selectType=ALLBUT0999 requests every category except warrants, matching
// the current TWSE UI's "全部(不含權證)" option.
func FetchMI_QFIIS(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_QFIIS: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("selectType", "ALLBUT0999")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	wire, err := FetchJSON[miQFIISWireResponse](ctx, client, "/fund/MI_QFIIS", q)
	if err != nil {
		return nil, err
	}
	return wire.toModel()
}

// ParseMI_QFIISRow converts one raw `data` row into a typed MI_QFIISRow.
func ParseMI_QFIISRow(row []string) (model.MI_QFIISRow, error) {
	if len(row) < 12 {
		return model.MI_QFIISRow{}, fmt.Errorf("MI_QFIIS: row too short: %d cols", len(row))
	}
	return model.MI_QFIISRow{
		Code:                    row[0],
		Name:                    row[1],
		ISIN:                    row[2],
		IssuedShares:            ParseInt(row[3]),
		AvailableShares:         ParseInt(row[4]),
		SharesHeld:              ParseInt(row[5]),
		AvailablePct:            ParsePercent(row[6]),
		IssuePct:                ParsePercent(row[7]),
		ForeignMainlandLimitPct: ParsePercent(row[8]),
		MainlandLimitPct:        ParsePercent(row[9]),
		ChangeReason:            row[10],
		LastReportedDate:        row[11],
	}, nil
}
