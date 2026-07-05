// Fetches and decodes Yahoo analyst upgrades/downgrades.

package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type UpgradeRow struct {
	EpochGradeDate int64  `json:"epochGradeDate"`
	Firm           string `json:"firm"`
	ToGrade        string `json:"toGrade"`
	FromGrade      string `json:"fromGrade"`
	Action         string `json:"action"`
}

type upgradesResult struct {
	QuoteSummary struct {
		Result []struct {
			UpgradeDowngradeHistory struct {
				History []UpgradeRow `json:"history"`
			} `json:"upgradeDowngradeHistory"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeUpgrades(data []byte) ([]UpgradeRow, error) {
	var r upgradesResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("upgrades: empty result")
	}
	return r.QuoteSummary.Result[0].UpgradeDowngradeHistory.History, nil
}

func (c *Client) FetchUpgrades(ctx context.Context, symbol string) ([]UpgradeRow, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"upgradeDowngradeHistory"})
	if err != nil {
		return nil, err
	}
	return DecodeUpgrades(raw)
}
