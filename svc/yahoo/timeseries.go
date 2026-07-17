// Fetches the latest annual income, balance-sheet, and cash-flow values from
// Yahoo's fundamentals-timeseries endpoint and converts them directly to the
// stable model.FundamentalsSnapshot surface.
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bizshuk/yfin/model"
)

const fundamentalsPeriodStart int64 = 1483142400 // 2016-12-31T00:00:00Z

// StatementKind selects one stable financial-statement surface.
type StatementKind string

const (
	IncomeStatement StatementKind = "income"
	BalanceSheet    StatementKind = "balance"
	CashFlow        StatementKind = "cashflow"
)

type statementField struct {
	yahooType string
	modelKey  string
}

type statementDefinition struct {
	source string
	fields []statementField
}

var statementDefinitions = map[StatementKind]statementDefinition{
	IncomeStatement: {
		source: "yahoo/fundamentals-timeseries/income",
		fields: []statementField{
			{"TotalRevenue", "total_revenue"},
			{"CostOfRevenue", "cost_of_revenue"},
			{"GrossProfit", "gross_profit"},
			{"OperatingExpense", "operating_expense"},
			{"OperatingIncome", "operating_income"},
			{"NetNonOperatingInterestIncomeExpense", "net_non_operating_interest_income_expense"},
			{"OtherIncomeExpense", "other_income_expense"},
			{"PretaxIncome", "pretax_income"},
			{"TaxProvision", "tax_provision"},
			{"NetIncomeCommonStockholders", "net_income"},
			{"BasicEPS", "eps_basic"},
			{"DilutedEPS", "eps_diluted"},
			{"BasicAverageShares", "shares_outstanding_basic"},
			{"DilutedAverageShares", "shares_outstanding_diluted"},
			{"TotalExpenses", "total_expenses"},
			{"NormalizedIncome", "normalized_income"},
			{"EBIT", "ebit"},
			{"EBITDA", "ebitda"},
			{"ReconciledCostOfRevenue", "reconciled_cost_of_revenue"},
			{"ReconciledDepreciation", "reconciled_depreciation"},
			{"NormalizedEBITDA", "normalized_ebitda"},
		},
	},
	BalanceSheet: {
		source: "yahoo/fundamentals-timeseries/balance",
		fields: []statementField{
			{"TotalAssets", "total_assets"},
			{"TotalDebt", "total_debt"},
			{"CommonStockEquity", "shareholders_equity"},
			{"WorkingCapital", "working_capital"},
			{"TangibleBookValue", "tangible_book_value"},
		},
	},
	CashFlow: {
		source: "yahoo/fundamentals-timeseries/cashflow",
		fields: []statementField{
			{"OperatingCashFlow", "operating_cash_flow"},
			{"InvestingCashFlow", "investing_cash_flow"},
			{"FinancingCashFlow", "financing_cash_flow"},
			{"FreeCashFlow", "free_cash_flow"},
			{"CapitalExpenditure", "capital_expenditure"},
		},
	},
}

type timeseriesEnvelope struct {
	Timeseries struct {
		Result []json.RawMessage `json:"result"`
		Error  json.RawMessage   `json:"error"`
	} `json:"timeseries"`
}

type timeseriesMeta struct {
	Meta struct {
		Type []string `json:"type"`
	} `json:"meta"`
}

type timeseriesPoint struct {
	AsOfDate      string `json:"asOfDate"`
	CurrencyCode  string `json:"currencyCode"`
	ReportedValue struct {
		Raw *float64 `json:"raw"`
	} `json:"reportedValue"`
}

// FetchFinancialStatement fetches and decodes the latest annual values for a
// supported financial statement.
func (c *Client) FetchFinancialStatement(ctx context.Context, symbol string, kind StatementKind) (*model.FundamentalsSnapshot, error) {
	definition, ok := statementDefinitions[kind]
	if !ok {
		return nil, fmt.Errorf("unsupported statement kind %q", kind)
	}

	endpoint, err := url.Parse(strings.TrimRight(c.timeseriesBaseURL, "/") +
		"/ws/fundamentals-timeseries/v1/finance/timeseries/" + url.PathEscape(symbol))
	if err != nil {
		return nil, fmt.Errorf("build fundamentals-timeseries URL: %w", err)
	}
	types := make([]string, 0, len(definition.fields))
	for _, field := range definition.fields {
		types = append(types, "annual"+field.yahooType)
	}
	query := endpoint.Query()
	query.Set("symbol", symbol)
	query.Set("type", strings.Join(types, ","))
	query.Set("period1", fmt.Sprintf("%d", fundamentalsPeriodStart))
	query.Set("period2", fmt.Sprintf("%d", time.Now().UTC().Truncate(24*time.Hour).Add(24*time.Hour).Unix()))
	endpoint.RawQuery = query.Encode()

	ctx = circuitContext(ctx, circuitGroupTimeseries)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create fundamentals-timeseries request: %w", err)
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch annual %s statement: %w", kind, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read annual %s statement: %w", kind, err)
	}
	return decodeFinancialStatement(body, symbol, kind)
}

func decodeFinancialStatement(data []byte, symbol string, kind StatementKind) (*model.FundamentalsSnapshot, error) {
	definition, ok := statementDefinitions[kind]
	if !ok {
		return nil, fmt.Errorf("unsupported statement kind %q", kind)
	}

	var envelope timeseriesEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode annual %s statement: %w", kind, err)
	}
	if raw := strings.TrimSpace(string(envelope.Timeseries.Error)); raw != "" && raw != "null" {
		return nil, fmt.Errorf("annual %s statement returned Yahoo error: %s", kind, raw)
	}

	series := make(map[string][]timeseriesPoint, len(envelope.Timeseries.Result))
	for _, result := range envelope.Timeseries.Result {
		var meta timeseriesMeta
		if err := json.Unmarshal(result, &meta); err != nil || len(meta.Meta.Type) == 0 {
			continue
		}
		typeName := meta.Meta.Type[0]
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(result, &fields); err != nil {
			continue
		}
		var points []timeseriesPoint
		if err := json.Unmarshal(fields[typeName], &points); err != nil {
			continue
		}
		series[typeName] = points
	}

	lines := make([]model.FundamentalsLine, 0, len(definition.fields))
	var asOf time.Time
	for _, field := range definition.fields {
		point, end, found := latestAnnualPoint(series["annual"+field.yahooType])
		if !found {
			continue
		}
		currency := strings.ToUpper(strings.TrimSpace(point.CurrencyCode))
		if len(currency) != 3 {
			currency = ""
		}
		lines = append(lines, model.FundamentalsLine{
			Key:          field.modelKey,
			Value:        *point.ReportedValue.Raw,
			CurrencyCode: currency,
			PeriodStart:  end.AddDate(-1, 0, 1),
			PeriodEnd:    end,
		})
		if end.After(asOf) {
			asOf = end
		}
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no annual %s data for %s", kind, symbol)
	}

	return &model.FundamentalsSnapshot{
		Symbol: symbol,
		Source: definition.source,
		AsOf:   asOf,
		Lines:  lines,
	}, nil
}

func latestAnnualPoint(points []timeseriesPoint) (timeseriesPoint, time.Time, bool) {
	var latest timeseriesPoint
	var latestDate time.Time
	found := false
	for _, point := range points {
		if point.ReportedValue.Raw == nil {
			continue
		}
		date, err := time.Parse(time.DateOnly, point.AsOfDate)
		if err != nil || (found && !date.After(latestDate)) {
			continue
		}
		latest = point
		latestDate = date.UTC()
		found = true
	}
	return latest, latestDate, found
}
