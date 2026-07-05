package facade

import "github.com/bizshuk/yfinance-go/svc/norm"

// CompanyInfo is the public-shaped company information.
type CompanyInfo struct {
	Symbol           string `json:"symbol"`
	LongName         string `json:"long_name,omitempty"`
	ShortName        string `json:"short_name,omitempty"`
	Exchange         string `json:"exchange,omitempty"`
	FullExchangeName string `json:"full_exchange_name,omitempty"`
	Currency         string `json:"currency,omitempty"`
	InstrumentType   string `json:"instrument_type,omitempty"`
	Timezone         string `json:"timezone,omitempty"`
}

// FromCompanyInfo converts the internal norm company info to a public struct.
func FromCompanyInfo(ci *norm.NormalizedCompanyInfo) *CompanyInfo {
	if ci == nil {
		return nil
	}
	return &CompanyInfo{
		Symbol:           ci.Security.Symbol,
		LongName:         ci.LongName,
		ShortName:        ci.ShortName,
		Exchange:         ci.Exchange,
		FullExchangeName: ci.FullExchangeName,
		Currency:         ci.Currency,
		InstrumentType:   ci.InstrumentType,
		Timezone:         ci.Timezone,
	}
}