package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type InsiderDTO struct {
	Transactions     []InsiderTransaction
	PurchaseActivity NetSharePurchaseActivity
	Roster           []InsiderHolder
}

type InsiderTransaction struct {
	FilerName       string `json:"filerName"`
	TransactionText string `json:"transactionText"`
	Shares          RawInt `json:"shares"`
	Value           RawInt `json:"value"`
	StartDate       RawInt `json:"startDate"`
	OwnershipType   string `json:"ownership"`
}

type NetSharePurchaseActivity struct {
	Period                   string   `json:"period"`
	BuyInfoShares            RawInt   `json:"buyInfoShares"`
	SellInfoShares           RawInt   `json:"sellInfoShares"`
	NetInfoShares            RawInt   `json:"netInfoShares"`
	TotalInsiderShares       RawInt   `json:"totalInsiderShares"`
	NetPercentInsiderShares  RawValue `json:"netPercentInsiderShares"`
	BuyPercentInsiderShares  RawValue `json:"buyPercentInsiderShares"`
	SellPercentInsiderShares RawValue `json:"sellPercentInsiderShares"`
	BuyInfoCount             RawInt   `json:"buyInfoCount"`
	SellInfoCount            RawInt   `json:"sellInfoCount"`
	NetInfoCount             RawInt   `json:"netInfoCount"`
}

// InsiderPurchaseTable is one row in the yfinance-style label/value table.
type InsiderPurchaseTable struct {
	LabelColumn string
	Labels      []string
	Shares      []any
	Trans       []any
}

func InsiderPurchaseSummaryTable(a *NetSharePurchaseActivity) InsiderPurchaseTable {
	label := "Insider Purchases Last " + a.Period
	labels := []string{
		"Purchases",
		"Sales",
		"Net Shares Purchased (Sold)",
		"Total Insider Shares Held",
		"% Net Shares Purchased (Sold)",
		"% Buy Shares",
		"% Sell Shares",
	}
	shares := []any{
		rawIntValue(a.BuyInfoShares),
		rawIntValue(a.SellInfoShares),
		rawIntValue(a.NetInfoShares),
		rawIntValue(a.TotalInsiderShares),
		rawFloatValue(a.NetPercentInsiderShares),
		rawFloatValue(a.BuyPercentInsiderShares),
		rawFloatValue(a.SellPercentInsiderShares),
	}
	trans := []any{
		rawIntValue(a.BuyInfoCount),
		rawIntValue(a.SellInfoCount),
		rawIntValue(a.NetInfoCount),
		nil, nil, nil, nil,
	}
	return InsiderPurchaseTable{LabelColumn: label, Labels: labels, Shares: shares, Trans: trans}
}

func rawIntValue(r RawInt) any {
	if r.Raw == nil {
		return nil
	}
	return *r.Raw
}

func rawFloatValue(r RawValue) any {
	if r.Raw == nil {
		return nil
	}
	return *r.Raw
}

type InsiderHolder struct {
	Name            string `json:"name"`
	Relation        string `json:"relation"`
	PositionDirect  RawInt `json:"positionDirect"`
	LatestTransDate RawInt `json:"latestTransDate"`
}

type insiderResult struct {
	QuoteSummary struct {
		Result []struct {
			InsiderTransactions struct {
				Transactions []InsiderTransaction `json:"transactions"`
			} `json:"insiderTransactions"`
			NetSharePurchaseActivity NetSharePurchaseActivity `json:"netSharePurchaseActivity"`
			InsiderHolders           struct {
				Holders []InsiderHolder `json:"holders"`
			} `json:"insiderHolders"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeInsider(data []byte) (*InsiderDTO, error) {
	var r insiderResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("insider: empty result")
	}
	res := r.QuoteSummary.Result[0]
	return &InsiderDTO{
		Transactions:     res.InsiderTransactions.Transactions,
		PurchaseActivity: res.NetSharePurchaseActivity,
		Roster:           res.InsiderHolders.Holders,
	}, nil
}

func (c *Client) FetchInsider(ctx context.Context, symbol string) (*InsiderDTO, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol,
		[]string{"insiderTransactions", "netSharePurchaseActivity", "insiderHolders"})
	if err != nil {
		return nil, err
	}
	return DecodeInsider(raw)
}
