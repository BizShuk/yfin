// Fetches and decodes Yahoo SEC filings.

package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type SecFiling struct {
	Date      string `json:"date"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	EdgarURL  string `json:"edgarUrl"`
	EpochDate int64  `json:"epochDate"`
}

type secFilingsResult struct {
	QuoteSummary struct {
		Result []struct {
			SecFilings struct {
				Filings []SecFiling `json:"filings"`
			} `json:"secFilings"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeSecFilings(data []byte) ([]SecFiling, error) {
	var r secFilingsResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("secFilings: empty result")
	}
	return r.QuoteSummary.Result[0].SecFilings.Filings, nil
}

func (c *Client) FetchSecFilings(ctx context.Context, symbol string) ([]SecFiling, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"secFilings"})
	if err != nil {
		return nil, err
	}
	return DecodeSecFilings(raw)
}
