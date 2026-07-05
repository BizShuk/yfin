// types.go — shared `Response` envelope (stat/title/fields/data/notes/total + `GetStat` getter) embedded by every endpoint DTO. Capacity: 1 base struct per endpoint.
package twse

// statNoData is TWSE's traditional "no data" message (varies by endpoint, so we
// substring-check). The exact string from TWSE: "很抱歉，沒有符合條件的資料!" plus
// the Latin "No data" variant.
const statNoData = "沒有符合條件的資料"

// Response is the common TWSE JSON envelope; concrete endpoints embed this
// and add their own extra fields (e.g. "date", "stockNo").
type Response struct {
	Stat   string     `json:"stat"`
	Title  string     `json:"title,omitempty"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
	Notes  []string   `json:"notes,omitempty"`
	Total  int        `json:"total,omitempty"`
	// Catch-all for endpoint-specific fields (date, stockNo, etc.) - decoded separately if needed.
	Extra map[string]any `json:"-"`
}

// GetStat exposes the embedded Stat field so callers / FetchJSON can read it
// without importing reflection.
func (r *Response) GetStat() string { return r.Stat }
