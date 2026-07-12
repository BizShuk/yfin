// company_info.go — `CompanyInfo` plain data struct for the SDK surface.
// Originally lived in facade/company_info.go; promoted to model/ for cross-layer reuse.

package model

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
