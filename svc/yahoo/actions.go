package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

type Dividend struct {
	Date   int64   `json:"date"`
	Amount float64 `json:"amount"`
}

type Split struct {
	Date        int64  `json:"date"`
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	SplitRatio  string `json:"splitRatio"`
}

type ActionsDTO struct {
	Dividends []Dividend
	Splits    []Split
}

type actionsResult struct {
	Chart struct {
		Result []struct {
			Events struct {
				Dividends map[string]Dividend `json:"dividends"`
				Splits    map[string]Split    `json:"splits"`
			} `json:"events"`
		} `json:"result"`
	} `json:"chart"`
}

func ExtractActions(data []byte) (*ActionsDTO, error) {
	var r actionsResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.Chart.Result) == 0 {
		return nil, fmt.Errorf("actions: empty result")
	}
	ev := r.Chart.Result[0].Events
	out := &ActionsDTO{}
	for _, d := range ev.Dividends {
		out.Dividends = append(out.Dividends, d)
	}
	for _, s := range ev.Splits {
		out.Splits = append(out.Splits, s)
	}
	sort.Slice(out.Dividends, func(i, j int) bool { return out.Dividends[i].Date < out.Dividends[j].Date })
	sort.Slice(out.Splits, func(i, j int) bool { return out.Splits[i].Date < out.Splits[j].Date })
	return out, nil
}

func (c *Client) FetchActions(ctx context.Context, symbol string) (*ActionsDTO, error) {
	// Use a 1-year lookback so we capture dividends + splits history.
	raw, err := c.fetchChartRaw(ctx, symbol, 365)
	if err != nil {
		return nil, err
	}
	return ExtractActions(raw)
}
