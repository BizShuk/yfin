// company_info.go — `CompanyInfo` type alias + `FromCompanyInfo` string
// passthrough converter. Struct lives in model/company_info.go; facade.CompanyInfo
// is a back-compat alias.
package facade

import (
	"github.com/bizshuk/yfin/model"
)

// CompanyInfo is the public-shaped company information. Aliased from
// model.CompanyInfo — new code should use model.CompanyInfo directly.
type CompanyInfo = model.CompanyInfo

// FromCompanyInfo converts the internal model company info to a public struct.
func FromCompanyInfo(ci *model.NormalizedCompanyInfo) *CompanyInfo {
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
