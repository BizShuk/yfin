// Fetches and decodes Yahoo ESG scores.

package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type ESGDTO struct {
	TotalEsg           RawValue `json:"totalEsg"`
	EnvironmentScore   RawValue `json:"environmentScore"`
	SocialScore        RawValue `json:"socialScore"`
	GovernanceScore    RawValue `json:"governanceScore"`
	RatingYear         int      `json:"ratingYear"`
	HighestControversy RawValue `json:"highestControversy"`
}

type esgResult struct {
	QuoteSummary struct {
		Result []struct {
			ESGScores ESGDTO `json:"esgScores"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeESG(data []byte) (*ESGDTO, error) {
	var r esgResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("esg: empty result")
	}
	d := r.QuoteSummary.Result[0].ESGScores
	return &d, nil
}

func (c *Client) FetchESG(ctx context.Context, symbol string) (*ESGDTO, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"esgScores"})
	if err != nil {
		return nil, err
	}
	return DecodeESG(raw)
}
