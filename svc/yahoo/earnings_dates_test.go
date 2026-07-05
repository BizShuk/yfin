package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEarningsDatesHTML(t *testing.T) {
	// Minimal HTML mimicking the calendar/earnings table.
	html := []byte(`
<html><body><table>
  <tr><th>Symbol</th><th>Company</th><th>Call Time</th><th>EPS Estimate</th><th>Reported EPS</th><th>Surprise(%)</th><th>Date</th></tr>
  <tr><td>INTC</td><td>Intel</td><td>AMC</td><td>2.97</td><td>-</td><td>-</td><td>2025-10-30</td></tr>
  <tr><td>INTC</td><td>Intel</td><td>AMC</td><td>1.73</td><td>1.54</td><td>-10.88</td><td>2025-07-22</td></tr>
</table></body></html>`)

	rows, err := ParseEarningsDatesHTML(html, "INTC")
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "2025-10-30", rows[0].Date)
	require.InDelta(t, 2.97, *rows[0].EPSEstimate, 0.001)
	require.Nil(t, rows[0].ReportedEPS) // "-" parses to nil
	require.Nil(t, rows[0].SurprisePct)
	require.InDelta(t, 1.54, *rows[1].ReportedEPS, 0.001)
	require.InDelta(t, -10.88, *rows[1].SurprisePct, 0.001)
}
