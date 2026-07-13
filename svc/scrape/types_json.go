// types_json.go — back-compat type aliases for `model.*` scrape value types
// (`Scaled`→`model.ScaledDecimal`, `Currency`, `YahooNum/YahooInt/YahooString`,
// `ToYahooNum`/`ToYahooInt` helpers) and DTOs (`KeyStatisticsDTO`,
// `FinancialsDTO`, `ProfileDTO`, `AnalysisDTO`, `PeriodLine`, `Recommendation`,
// `QuarterlyEPS`, `Officer`). All of these now live in model/scrape.go and
// model/scrape_dtos.go. This file retains the coercion helpers
// (`CoerceCurrency`, `ParseYahooDate`, `ParseYahooPeriod`, `StringToInt64`,
// `StringToFloat64`) which carry parsing logic and don't belong in model/.

package scrape

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bizshuk/yfin/model"
)

// Value type aliases — defined in model/scrape.go.
type (
	Scaled      = model.Scaled // alias to model.ScaledDecimal
	Currency    = model.Currency
	YahooNum    = model.YahooNum
	YahooInt    = model.YahooInt
	YahooString = model.YahooString
)

// DTO aliases — defined in model/scrape_dtos.go.
type (
	KeyStatisticsDTO = model.KeyStatisticsDTO
	PeriodLine       = model.PeriodLine
	FinancialsDTO    = model.FinancialsDTO
	Recommendation   = model.Recommendation
	QuarterlyEPS     = model.QuarterlyEPS
	AnalysisDTO      = model.AnalysisDTO
	Officer          = model.Officer
	ProfileDTO       = model.ProfileDTO
)

// ToYahooNum converts a raw struct to YahooNum.
func ToYahooNum(raw *float64, fmtStr, longFmt string) model.YahooNum {
	return model.ToYahooNum(raw, fmtStr, longFmt)
}

// ToYahooInt converts a raw struct to YahooInt.
func ToYahooInt(raw *int64, fmtStr, longFmt string) model.YahooInt {
	return model.ToYahooInt(raw, fmtStr, longFmt)
}

// Numeric coercion helpers — scraping-side logic, retained here.

// CoerceCurrency extracts currency from various Yahoo formats
func CoerceCurrency(v any) (Currency, bool) {
	switch val := v.(type) {
	case string:
		currency := strings.TrimSpace(strings.ToUpper(val))
		if len(currency) == 3 {
			return currency, true
		}
		return "", false
	case map[string]any:
		if curr, ok := val["currency"].(string); ok {
			return CoerceCurrency(curr)
		}
		return "", false
	default:
		return "", false
	}
}

// ParseYahooDate parses various Yahoo date formats
func ParseYahooDate(ts any) (time.Time, bool) {
	switch val := ts.(type) {
	case float64:
		return time.Unix(int64(val), 0).UTC(), true
	case int64:
		return time.Unix(val, 0).UTC(), true
	case string:
		formats := []string{
			"2006-01-02",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
			"2006-01-02 15:04:05",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t.UTC(), true
			}
		}
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}

// ParseYahooPeriod parses Yahoo's period format (e.g., "2023-12-31")
func ParseYahooPeriod(periodStr string) (time.Time, time.Time, bool) {
	if t, ok := ParseYahooDate(periodStr); ok {
		year := t.Year()
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		return start, t, true
	}
	return time.Time{}, time.Time{}, false
}

// StringToInt64 safely converts a string to int64
func StringToInt64(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	val, err := strconv.ParseInt(s, 10, 64)
	return val, err == nil
}

// StringToFloat64 safely converts a string to float64
func StringToFloat64(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	val, err := strconv.ParseFloat(s, 64)
	return val, err == nil
}

// (fmt import retained for downstream parse functions that may use it.)
var _ = fmt.Sprintf
