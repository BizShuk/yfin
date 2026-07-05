// parse.go — TWSE cell-level numeric helpers `ParseFloat`/`ParseInt`/`ParsePercent`/`StripNumberFmt` (thousands-separator aware). Capacity: 4 helpers consumed by all endpoint row parsers.
package twse

import (
	"strconv"
	"strings"
)

// ParseFloat parses a TWSE-formatted number into a float64. It strips
// thousands separators (commas) and surrounding whitespace/quotes, then
// delegates to strconv.ParseFloat. On failure it returns 0.
func ParseFloat(s string) float64 {
	s = StripNumberFmt(s)
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ParseInt parses a TWSE-formatted integer (e.g. trade volume) into an
// int64. It strips thousands separators before parsing. On failure
// returns 0.
func ParseInt(s string) int64 {
	s = StripNumberFmt(s)
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// ParsePercent parses a percentage string like "+1.23%" or "-0.45%"
// into a float64 (without dividing by 100). On failure returns 0.
func ParsePercent(s string) float64 {
	s = StripNumberFmt(s)
	s = strings.TrimSuffix(s, "%")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// StripNumberFmt removes thousands separators (commas) and trims
// surrounding whitespace and double-quote characters that TWSE
// sometimes includes in CSV-like cells.
func StripNumberFmt(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	for strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ",", "")
	}
	return s
}
