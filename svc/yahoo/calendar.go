// Fetches and decodes Yahoo earnings-calendar data.

package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type CalendarDTO struct {
	EarningsDates   []int64
	EarningsAverage RawValue
	RevenueAverage  RawValue
	ExDividendDate  RawInt
	DividendDate    RawInt
}

type calendarResult struct {
	QuoteSummary struct {
		Result []struct {
			CalendarEvents struct {
				Earnings struct {
					EarningsDate    []RawInt `json:"earningsDate"`
					EarningsAverage RawValue `json:"earningsAverage"`
					RevenueAverage  RawValue `json:"revenueAverage"`
				} `json:"earnings"`
				ExDividendDate RawInt `json:"exDividendDate"`
				DividendDate   RawInt `json:"dividendDate"`
			} `json:"calendarEvents"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeCalendar(data []byte) (*CalendarDTO, error) {
	var r calendarResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("calendar: empty result")
	}
	res := r.QuoteSummary.Result[0].CalendarEvents
	dto := &CalendarDTO{
		EarningsAverage: res.Earnings.EarningsAverage,
		RevenueAverage:  res.Earnings.RevenueAverage,
		ExDividendDate:  res.ExDividendDate,
		DividendDate:    res.DividendDate,
	}
	for _, e := range res.Earnings.EarningsDate {
		if e.Raw != nil {
			dto.EarningsDates = append(dto.EarningsDates, *e.Raw)
		}
	}
	return dto, nil
}

func (c *Client) FetchCalendar(ctx context.Context, symbol string) (*CalendarDTO, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"calendarEvents"})
	if err != nil {
		return nil, err
	}
	return DecodeCalendar(raw)
}
