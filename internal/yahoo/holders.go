package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type HoldersDTO struct {
	MajorBreakdown       MajorHoldersBreakdown
	InstitutionOwnership []HolderRow
	FundOwnership        []HolderRow
}

type MajorHoldersBreakdown struct {
	InsidersPercentHeld      RawValue `json:"insidersPercentHeld"`
	InstitutionsPercentHeld  RawValue `json:"institutionsPercentHeld"`
	InstitutionsFloatPctHeld RawValue `json:"institutionsFloatPercentHeld"`
	InstitutionsCount        RawInt   `json:"institutionsCount"`
}

type HolderRow struct {
	Organization string   `json:"organization"`
	PctHeld      RawValue `json:"pctHeld"`
	Position     RawInt   `json:"position"`
	Value        RawInt   `json:"value"`
}

type holdersResult struct {
	QuoteSummary struct {
		Result []struct {
			MajorHoldersBreakdown MajorHoldersBreakdown `json:"majorHoldersBreakdown"`
			InstitutionOwnership  struct {
				OwnershipList []HolderRow `json:"ownershipList"`
			} `json:"institutionOwnership"`
			FundOwnership struct {
				OwnershipList []HolderRow `json:"ownershipList"`
			} `json:"fundOwnership"`
		} `json:"result"`
		Error *struct {
			Description string `json:"description"`
		} `json:"error"`
	} `json:"quoteSummary"`
}

func DecodeHolders(data []byte) (*HoldersDTO, error) {
	var r holdersResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("holders: empty result")
	}
	res := r.QuoteSummary.Result[0]
	return &HoldersDTO{
		MajorBreakdown:       res.MajorHoldersBreakdown,
		InstitutionOwnership: res.InstitutionOwnership.OwnershipList,
		FundOwnership:        res.FundOwnership.OwnershipList,
	}, nil
}

func (c *Client) FetchHolders(ctx context.Context, symbol string) (*HoldersDTO, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol,
		[]string{"majorHoldersBreakdown", "institutionOwnership", "fundOwnership"})
	if err != nil {
		return nil, err
	}
	return DecodeHolders(raw)
}
