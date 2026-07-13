// fundamentals.go — quarterly fundamentals HTTP fetch + JSON decode +
// per-statement validation. Type definitions (FundamentalsResponse,
// QuoteSummary, FundamentalsResult, IncomeStatementHistory, BalanceSheetHistory,
// CashflowStatementHistory, IncomeStatement, BalanceSheet, CashflowStatement,
// DateValue, Value, Fundamentals) live in `model/yahoo_raw.go`.
package yahoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bizshuk/yfin/model"
)

// Back-compat type aliases.
type (
	FundamentalsResponse        = model.FundamentalsResponse
	QuoteSummary                = model.QuoteSummary
	FundamentalsResult          = model.FundamentalsResult
	IncomeStatementHistory      = model.IncomeStatementHistory
	BalanceSheetHistory         = model.BalanceSheetHistory
	CashflowStatementHistory    = model.CashflowStatementHistory
	IncomeStatement             = model.IncomeStatement
	BalanceSheet                = model.BalanceSheet
	CashflowStatement           = model.CashflowStatement
	DateValue                   = model.DateValue
	Value                       = model.Value
	Fundamentals                = model.Fundamentals
)

// DecodeFundamentalsResponse decodes a Yahoo Finance fundamentals response
// (unknown fields allowed because the response has many fields we don't use).
func DecodeFundamentalsResponse(data []byte) (*model.FundamentalsResponse, error) {
	var response model.FundamentalsResponse
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode fundamentals response: %w", err)
	}
	if err := ValidateFundamentals(&response); err != nil {
		return nil, fmt.Errorf("invalid fundamentals response: %w", err)
	}
	return &response, nil
}

// DecodeFundamentalsResponseFromReader is the streaming variant.
func DecodeFundamentalsResponseFromReader(reader io.Reader) (*model.FundamentalsResponse, error) {
	var response model.FundamentalsResponse
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode fundamentals response: %w", err)
	}
	if err := ValidateFundamentals(&response); err != nil {
		return nil, fmt.Errorf("invalid fundamentals response: %w", err)
	}
	return &response, nil
}

// ValidateFundamentals runs structural + per-statement validation.
func ValidateFundamentals(r *model.FundamentalsResponse) error {
	if r.QuoteSummary.Error != nil {
		return fmt.Errorf("yahoo finance error: %s", *r.QuoteSummary.Error)
	}
	if len(r.QuoteSummary.Result) == 0 {
		return fmt.Errorf("no fundamentals results found")
	}
	for i, result := range r.QuoteSummary.Result {
		if err := validateFundamentalsResult(result); err != nil {
			return fmt.Errorf("result[%d]: %w", i, err)
		}
	}
	return nil
}

func validateFundamentalsResult(r model.FundamentalsResult) error {
	if r.IncomeStatementHistoryQuarterly == nil &&
		r.BalanceSheetHistoryQuarterly == nil &&
		r.CashflowStatementHistoryQuarterly == nil {
		return fmt.Errorf("no financial statements found")
	}
	if r.IncomeStatementHistoryQuarterly != nil {
		if err := validateIncomeStatementHistory(r.IncomeStatementHistoryQuarterly); err != nil {
			return fmt.Errorf("income statement: %w", err)
		}
	}
	if r.BalanceSheetHistoryQuarterly != nil {
		if err := validateBalanceSheetHistory(r.BalanceSheetHistoryQuarterly); err != nil {
			return fmt.Errorf("balance sheet: %w", err)
		}
	}
	if r.CashflowStatementHistoryQuarterly != nil {
		if err := validateCashflowStatementHistory(r.CashflowStatementHistoryQuarterly); err != nil {
			return fmt.Errorf("cashflow statement: %w", err)
		}
	}
	return nil
}

func validateIncomeStatementHistory(h *model.IncomeStatementHistory) error {
	if len(h.IncomeStatementHistory) == 0 {
		return fmt.Errorf("no income statements found")
	}
	for i, stmt := range h.IncomeStatementHistory {
		if err := validateIncomeStatement(stmt); err != nil {
			return fmt.Errorf("income statement[%d]: %w", i, err)
		}
	}
	return nil
}

func validateBalanceSheetHistory(h *model.BalanceSheetHistory) error {
	if len(h.BalanceSheetHistory) == 0 {
		return fmt.Errorf("no balance sheets found")
	}
	for i, sheet := range h.BalanceSheetHistory {
		if err := validateBalanceSheet(sheet); err != nil {
			return fmt.Errorf("balance sheet[%d]: %w", i, err)
		}
	}
	return nil
}

func validateCashflowStatementHistory(h *model.CashflowStatementHistory) error {
	if len(h.CashflowStatementHistory) == 0 {
		return fmt.Errorf("no cashflow statements found")
	}
	for i, stmt := range h.CashflowStatementHistory {
		if err := validateCashflowStatement(stmt); err != nil {
			return fmt.Errorf("cashflow statement[%d]: %w", i, err)
		}
	}
	return nil
}

func validateIncomeStatement(s model.IncomeStatement) error {
	if s.EndDate.Raw == 0 {
		return fmt.Errorf("missing end date")
	}
	return nil
}

func validateBalanceSheet(s model.BalanceSheet) error {
	if s.EndDate.Raw == 0 {
		return fmt.Errorf("missing end date")
	}
	return nil
}

func validateCashflowStatement(s model.CashflowStatement) error {
	if s.EndDate.Raw == 0 {
		return fmt.Errorf("missing end date")
	}
	return nil
}