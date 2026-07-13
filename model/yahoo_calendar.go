// yahoo_calendar.go — Yahoo earnings-calendar + dividend-dates DTOs.
// Originally split across svc/yahoo/calendar.go (`CalendarDTO`) and
// svc/yahoo/earnings_dates.go (`EarningsDateRow`); promoted to model/ so
// external consumers can depend on the shapes without importing the
// Decode / Fetch / HTML-parse behavior of svc/yahoo.

package model

// CalendarDTO is the decoded calendarEvents result for a single symbol.
type CalendarDTO struct {
	EarningsDates   []int64
	EarningsAverage RawValue
	RevenueAverage  RawValue
	ExDividendDate  RawInt
	DividendDate    RawInt
}

// EarningsDateRow mirrors yfinance's earnings_dates DataFrame columns.
type EarningsDateRow struct {
	Date        string   `json:"date"`
	EPSEstimate *float64 `json:"eps_estimate,omitempty"`
	ReportedEPS *float64 `json:"reported_eps,omitempty"`
	SurprisePct *float64 `json:"surprise_pct,omitempty"`
}