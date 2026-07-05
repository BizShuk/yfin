// Fetches and decodes Yahoo recommendation-trend data.

package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type RecommendationTrendRow struct {
	Period     string `json:"period"`
	StrongBuy  int    `json:"strongBuy"`
	Buy        int    `json:"buy"`
	Hold       int    `json:"hold"`
	Sell       int    `json:"sell"`
	StrongSell int    `json:"strongSell"`
}

type recTrendResult struct {
	QuoteSummary struct {
		Result []struct {
			RecommendationTrend struct {
				Trend []RecommendationTrendRow `json:"trend"`
			} `json:"recommendationTrend"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeRecommendationTrend(data []byte) ([]RecommendationTrendRow, error) {
	var r recTrendResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("recommendationTrend: empty result")
	}
	return r.QuoteSummary.Result[0].RecommendationTrend.Trend, nil
}

func (c *Client) FetchRecommendationTrend(ctx context.Context, symbol string) ([]RecommendationTrendRow, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"recommendationTrend"})
	if err != nil {
		return nil, err
	}
	return DecodeRecommendationTrend(raw)
}
