// tests/unit/scrape/coverage_test.go — Smoke tests for the parts of svc/scrape
// that have no end-to-end coverage: regex config loaders, error helpers, the
// pure helpers in types_json.go, and the high-level HTML/JSON parsers
// (financials / statistics / analysis / analyst-insights / profile / news-json)
// exercised against minimal synthetic input that matches each parser's regex
// patterns. These are not real-data fixtures — they verify the parse functions
// execute without panic and produce non-nil DTOs with at least one populated
// field. Real Yahoo HTML fixture work belongs in tests/testdata/fixtures/yahoo.
package scrape_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bizshuk/yfin/svc/scrape"
)

// --- Regex config loaders ---------------------------------------------------

func TestLoadAllRegexConfigs(t *testing.T) {
	t.Parallel()
	loaders := map[string]func() error{
		"LoadFinancialsRegexConfig":      scrape.LoadFinancialsRegexConfig,
		"LoadRegexConfig (statistics)":   scrape.LoadRegexConfig,
		"LoadAnalysisRegexConfig":        scrape.LoadAnalysisRegexConfig,
		"LoadAnalystInsightsRegexConfig": scrape.LoadAnalystInsightsRegexConfig,
		"LoadNewsRegexConfig":            scrape.LoadNewsRegexConfig,
	}
	for name, load := range loaders {
		if err := load(); err != nil {
			t.Errorf("%s failed: %v", name, err)
		}
	}
}

// --- Error helpers ----------------------------------------------------------

func TestScrapeError_Format(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  *scrape.ScrapeError
		want []string
	}{
		{
			name: "without status",
			err:  &scrape.ScrapeError{Type: "invalid_url", Message: "bad url", URL: "https://x"},
			want: []string{"invalid_url", "bad url", "https://x"},
		},
		{
			name: "with status",
			err:  &scrape.ScrapeError{Type: "http_error", Message: "Not Found", URL: "https://x", Status: 404},
			want: []string{"http_error", "Not Found", "https://x", "404"},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.err.Error()
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("Error() = %q; missing %q", got, want)
				}
			}
		})
	}
}

func TestErrHTTP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status       int
		wantType     string
		wantMsgSub   string
	}{
		{404, "http_error", "Not Found"},
		{500, "http_error", "Internal Server Error"},
		{429, "http_error", "Too Many Requests"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run("", func(t *testing.T) {
			t.Parallel()
			err := scrape.ErrHTTP(tc.status, "https://example.com")
			if err.Type != tc.wantType {
				t.Errorf("Type = %q, want %q", err.Type, tc.wantType)
			}
			if err.Status != tc.status {
				t.Errorf("Status = %d, want %d", err.Status, tc.status)
			}
			if !strings.Contains(err.Message, tc.wantMsgSub) {
				t.Errorf("Message = %q, want substring %q", err.Message, tc.wantMsgSub)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"timeout", scrape.ErrTimeout, true},
		{"rate_limited", scrape.ErrRateLimited, true},
		{"http 429", scrape.ErrHTTP(429, "https://x"), true},
		{"http 500", scrape.ErrHTTP(500, "https://x"), true},
		{"http 502", scrape.ErrHTTP(502, "https://x"), true},
		{"http 404", scrape.ErrHTTP(404, "https://x"), false},
		{"http 401", scrape.ErrHTTP(401, "https://x"), false},
		{"robots_denied", scrape.ErrRobotsDenied, false},
		{"invalid_url", scrape.ErrInvalidURL, false},
		{"network connection refused", errors.New("dial tcp: connection refused"), true},
		{"network no such host", errors.New("dial: no such host"), true},
		{"unrelated text", errors.New("something else went wrong"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := scrape.IsRetryableError(tc.err); got != tc.want {
				t.Errorf("IsRetryableError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// --- types_json.go pure helpers --------------------------------------------

func TestCoerceCurrency(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		in    any
		want  string
		wantOk bool
	}{
		{"uppercase USD", "USD", "USD", true},
		{"lowercase twd", "twd", "TWD", true},
		{"with spaces", " JPY ", "JPY", true},
		{"too short", "US", "", false},
		{"too long", "USDX", "", false},
		{"map with currency", map[string]any{"currency": "eur"}, "EUR", true},
		{"map without currency", map[string]any{"other": "x"}, "", false},
		{"unsupported type", 123, "", false},
		{"nil", nil, "", false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := scrape.CoerceCurrency(tc.in)
			if ok != tc.wantOk || got != tc.want {
				t.Errorf("CoerceCurrency(%v) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.wantOk)
			}
		})
	}
}

func TestParseYahooDate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      any
		wantOk  bool
	}{
		{"unix float", float64(1700000000), true},
		{"unix int", int64(1700000000), true},
		{"iso date", "2026-01-15", true},
		{"iso datetime", "2026-01-15T10:00:00Z", true},
		{"iso datetime ms", "2026-01-15T10:00:00.000Z", true},
		{"space separated", "2026-01-15 10:00:00", true},
		{"garbage", "not a date", false},
		{"nil", nil, false},
		{"int", 42, false}, // int (not int64) is not handled
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, ok := scrape.ParseYahooDate(tc.in)
			if ok != tc.wantOk {
				t.Errorf("ParseYahooDate(%v) ok = %v, want %v", tc.in, ok, tc.wantOk)
			}
		})
	}
}

func TestParseYahooPeriod(t *testing.T) {
	t.Parallel()
	// Valid period
	start, end, ok := scrape.ParseYahooPeriod("2026-06-30")
	if !ok {
		t.Fatal("ParseYahooPeriod valid input returned ok=false")
	}
	if start.Year() != 2026 || start.Month() != time.January || start.Day() != 1 {
		t.Errorf("start = %v, want 2026-01-01", start)
	}
	if end.Year() != 2026 || end.Month() != time.June || end.Day() != 30 {
		t.Errorf("end = %v, want 2026-06-30", end)
	}
	// Invalid
	if _, _, ok := scrape.ParseYahooPeriod("garbage"); ok {
		t.Error("ParseYahooPeriod garbage returned ok=true")
	}
}

func TestStringToInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in     string
		want   int64
		wantOk bool
	}{
		{"123", 123, true},
		{"1,234,567", 1234567, true},
		{"  42  ", 42, true},
		{"-5", -5, true},
		{"", 0, false},
		{"abc", 0, false},
		{"3.14", 0, false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got, ok := scrape.StringToInt64(tc.in)
			if ok != tc.wantOk || got != tc.want {
				t.Errorf("StringToInt64(%q) = (%d, %v), want (%d, %v)", tc.in, got, ok, tc.want, tc.wantOk)
			}
		})
	}
}

func TestStringToFloat64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in     string
		want   float64
		wantOk bool
	}{
		{"3.14", 3.14, true},
		{"1,234.56", 1234.56, true},
		{"  42  ", 42, true},
		{"", 0, false},
		{"abc", 0, false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got, ok := scrape.StringToFloat64(tc.in)
			if ok != tc.wantOk || got != tc.want {
				t.Errorf("StringToFloat64(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.wantOk)
			}
		})
	}
}

// --- Synthetic fixtures for the high-level HTML/JSON parsers ---------------
//
// Each fixture is the *minimum* HTML/JSON needed to satisfy the regex
// patterns in svc/scrape/regex/*.yaml. These are smoke tests, not real-data
// tests — they confirm the parser pipeline (regex load → match → DTO fill)
// executes without panic and returns a non-nil DTO with at least one populated
// field.

const financialsHTML = `
<html><body>
<div>Currency in USD</div>
<div>Total Revenue</div></div> <div class="column yf-t22klz alt">394,328</div><div class="column yf-t22klz">383,285</div>
<div>Operating Income</div></div> <div class="column yf-t22klz alt">112,995</div><div class="column yf-t22klz">114,301</div>
<div>Net Income Common Stockholders</div></div> <div class="column yf-t22klz alt">93,736</div><div class="column yf-t22klz">96,995</div>
<div>Basic EPS</div></div> <div class="column yf-t22klz alt">6.11</div><div class="column yf-t22klz">6.16</div>
</body></html>
`

const statisticsHTML = `
<html><body><table>
<tr><td>Market Cap</td><td>2.85T</td></tr>
<tr><td>Trailing P/E</td><td>30.45</td></tr>
<tr><td>Forward P/E</td><td>28.10</td></tr>
<tr><td>Beta (5Y Monthly)</td><td>1.24</td></tr>
<tr><td>Shares Outstanding</td><td>15.41B</td></tr>
</table></body></html>
`

const analysisHTML = `
<html><body>
<section data-testid="earningsEstimate">
<table>
<tr class="yf-kl4vme"><td class="yf-kl4vme">No. of Analysts</td> <td class="yf-kl4vme">15</td><td class="yf-kl4vme">12</td><td class="yf-kl4vme">35</td><td class="yf-kl4vme">30</td> </tr>
<tr class="yf-kl4vme"><td class="yf-kl4vme">Avg. Estimate</td> <td class="yf-kl4vme">1.65</td><td class="yf-kl4vme">1.85</td><td class="yf-kl4vme">7.20</td><td class="yf-kl4vme">8.10</td> </tr>
<tr class="yf-kl4vme"><td class="yf-kl4vme">Low Estimate</td> <td class="yf-kl4vme">1.50</td><td class="yf-kl4vme">1.70</td><td class="yf-kl4vme">6.80</td><td class="yf-kl4vme">7.60</td> </tr>
<tr class="yf-kl4vme"><td class="yf-kl4vme">High Estimate</td> <td class="yf-kl4vme">1.80</td><td class="yf-kl4vme">2.00</td><td class="yf-kl4vme">7.60</td><td class="yf-kl4vme">8.50</td> </tr>
<tr class="yf-kl4vme"><td class="yf-kl4vme">Year Ago EPS</td> <td class="yf-kl4vme">1.46</td><td class="yf-kl4vme">1.64</td><td class="yf-kl4vme">6.16</td><td class="yf-kl4vme">6.11</td> </tr>
</table>
<div>Currency in USD</div>
</section>
</body></html>
`

// analystInsightsHTML embeds the financialData JSON block matched by the
// regex in regex/analyst_insights.yaml. The 8 captures are: currentPrice,
// targetMeanPrice, targetMedianPrice, targetHighPrice, targetLowPrice,
// recommendationMean, recommendationKey, numberOfAnalystOpinions.
const analystInsightsHTML = `
<html><body>
<script type="application/json">
{"root":{"financialData":{"currentPrice":{"raw":178.50,"fmt":"178.50"},"targetMeanPrice":{"raw":200.00,"fmt":"200.00"},"targetMedianPrice":{"raw":195.00,"fmt":"195.00"},"targetHighPrice":{"raw":250.00,"fmt":"250.00"},"targetLowPrice":{"raw":150.00,"fmt":"150.00"},"recommendationMean":{"raw":2.10,"fmt":"2.10"},"recommendationKey":"buy","numberOfAnalystOpinions":{"raw":35,"fmt":"35"}}}}
</script>
</body></html>
`

// profileHTML embeds the assetProfile JSON that ParseComprehensiveProfile
// extracts via the embedded `<script type="application/json">` payload.
const profileHTML = `
<html><body>
<script type="application/json" data-test="assetProfile">
{"body":"{\"quoteSummary\":{\"result\":[{\"assetProfile\":{\"address1\":\"One Apple Park Way\",\"city\":\"Cupertino\",\"state\":\"CA\",\"zip\":\"95014\",\"country\":\"USA\",\"phone\":\"(408) 996-1010\",\"website\":\"https://www.apple.com\",\"industry\":\"Consumer Electronics\",\"sector\":\"Technology\",\"fullTimeEmployees\":164000,\"longBusinessSummary\":\"Apple Inc. designs and markets consumer electronics.\",\"maxAge\":86400,\"auditRisk\":1,\"boardRisk\":1,\"compensationRisk\":3,\"shareHolderRightsRisk\":1,\"overallRisk\":1,\"companyOfficers\":[{\"name\":\"Tim Cook\",\"title\":\"CEO\",\"yearBorn\":1960,\"totalPay\":{\"raw\":98732700},\"exercisedValue\":{\"raw\":0},\"unexercisedValue\":{\"raw\":100000000}}]}}]}}"}
</script>
</body></html>
`

// newsJSONHTML contains a `<script>` with `tickerStream` and an embedded body
// carrying a STORY block — exercises the extract_news_json.go happy path.
const newsJSONHTML = `
<html><body>
<script>tickerStream("body",{"body":"{\"items\":[{\"id\":\"abc-123\",\"content\":{\"contentType\":\"STORY\",\"title\":\"Apple announces new product\",\"canonicalUrl\":{\"url\":\"https://finance.yahoo.com/news/apple-product-123.html\"},\"provider\":{\"displayName\":\"Yahoo Finance\"},\"pubDate\":\"2026-07-15T10:00:00Z\",\"originalUrl\":\"https://example.com/img.webp\",\"stockTickers\":[{\"symbol\":\"AAPL\"}]}}]}"});</script>
</body></html>
`

// --- High-level parser smoke tests -----------------------------------------

func TestParseComprehensiveFinancials_Smoke(t *testing.T) {
	t.Parallel()
	dto, err := scrape.ParseComprehensiveFinancials([]byte(financialsHTML), "AAPL", "NMS")
	if err != nil {
		t.Fatalf("ParseComprehensiveFinancials failed: %v", err)
	}
	if dto == nil {
		t.Fatal("dto is nil")
	}
	if dto.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", dto.Symbol)
	}
	if dto.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", dto.Currency)
	}
	// At least one income statement value should be populated from our fixture.
	if dto.Current.TotalRevenue == nil {
		t.Error("Current.TotalRevenue is nil; expected non-nil Scaled")
	} else if dto.Current.TotalRevenue.Scaled == 0 {
		t.Error("Current.TotalRevenue.Scaled = 0; expected non-zero from fixture")
	}
}

func TestParseComprehensiveFinancialsWithCurrency_Smoke(t *testing.T) {
	t.Parallel()
	dto, err := scrape.ParseComprehensiveFinancialsWithCurrency([]byte(financialsHTML), []byte(financialsHTML), "AAPL", "NMS")
	if err != nil {
		t.Fatalf("ParseComprehensiveFinancialsWithCurrency failed: %v", err)
	}
	if dto == nil {
		t.Fatal("dto is nil")
	}
	if dto.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", dto.Currency)
	}
}

func TestParseComprehensiveKeyStatistics_Smoke(t *testing.T) {
	t.Parallel()
	dto, err := scrape.ParseComprehensiveKeyStatistics([]byte(statisticsHTML), "AAPL", "NMS")
	if err != nil {
		t.Fatalf("ParseComprehensiveKeyStatistics failed: %v", err)
	}
	if dto == nil {
		t.Fatal("dto is nil")
	}
	if dto.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", dto.Symbol)
	}
	// Either current values or additional values should be populated.
	if dto.Current.MarketCap == nil && dto.Additional.Beta == nil {
		t.Error("expected at least one populated field in Current or Additional")
	}
}

func TestParseAnalysis_Smoke(t *testing.T) {
	t.Parallel()
	dto, err := scrape.ParseAnalysis([]byte(analysisHTML), "AAPL", "NMS")
	if err != nil {
		t.Fatalf("ParseAnalysis failed: %v", err)
	}
	if dto == nil {
		t.Fatal("dto is nil")
	}
	if dto.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", dto.Symbol)
	}
	if dto.EarningsEstimate.CurrentQtr.NoOfAnalysts == nil {
		t.Error("EarningsEstimate.CurrentQtr.NoOfAnalysts is nil; expected 15 from fixture")
	} else if *dto.EarningsEstimate.CurrentQtr.NoOfAnalysts != 15 {
		t.Errorf("EarningsEstimate.CurrentQtr.NoOfAnalysts = %d, want 15", *dto.EarningsEstimate.CurrentQtr.NoOfAnalysts)
	}
	if dto.EarningsEstimate.Currency != "USD" {
		t.Errorf("EarningsEstimate.Currency = %q, want USD", dto.EarningsEstimate.Currency)
	}
}

func TestParseAnalystInsights_Smoke(t *testing.T) {
	t.Parallel()
	dto, err := scrape.ParseAnalystInsights([]byte(analystInsightsHTML), "AAPL", "NMS")
	if err != nil {
		t.Fatalf("ParseAnalystInsights failed: %v", err)
	}
	if dto == nil {
		t.Fatal("dto is nil")
	}
	if dto.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", dto.Symbol)
	}
	if dto.CurrentPrice == nil {
		t.Error("CurrentPrice is nil; expected 178.5 from fixture")
	} else if *dto.CurrentPrice != 178.50 {
		t.Errorf("CurrentPrice = %v, want 178.50", *dto.CurrentPrice)
	}
	if dto.RecommendationKey == nil || *dto.RecommendationKey != "buy" {
		t.Errorf("RecommendationKey = %v, want 'buy'", dto.RecommendationKey)
	}
	if dto.NumberOfAnalysts == nil || *dto.NumberOfAnalysts != 35 {
		t.Errorf("NumberOfAnalysts = %v, want 35", dto.NumberOfAnalysts)
	}
}

func TestParseComprehensiveProfile_Smoke(t *testing.T) {
	t.Parallel()
	dto, err := scrape.ParseComprehensiveProfile([]byte(profileHTML), "AAPL", "NMS")
	if err != nil {
		t.Fatalf("ParseComprehensiveProfile failed: %v", err)
	}
	if dto == nil {
		t.Fatal("dto is nil")
	}
	if dto.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", dto.Symbol)
	}
	if dto.City != "Cupertino" {
		t.Errorf("City = %q, want Cupertino", dto.City)
	}
	if dto.Sector != "Technology" {
		t.Errorf("Sector = %q, want Technology", dto.Sector)
	}
	if dto.FullTimeEmployees == nil || *dto.FullTimeEmployees != 164000 {
		t.Errorf("FullTimeEmployees = %v, want 164000", dto.FullTimeEmployees)
	}
	if len(dto.Executives) != 1 {
		t.Fatalf("Executives count = %d, want 1", len(dto.Executives))
	}
	if dto.Executives[0].Name != "Tim Cook" {
		t.Errorf("Executive name = %q, want Tim Cook", dto.Executives[0].Name)
	}
}

func TestParseComprehensiveProfile_NoAssetProfile(t *testing.T) {
	t.Parallel()
	_, err := scrape.ParseComprehensiveProfile([]byte("<html><body>no script tag</body></html>"), "AAPL", "NMS")
	if err == nil {
		t.Error("expected error when HTML has no assetProfile script tag")
	}
}

func TestParseNews_JSONPath_Smoke(t *testing.T) {
	t.Parallel()
	articles, stats, err := scrape.ParseNews([]byte(newsJSONHTML), "https://finance.yahoo.com", time.Now())
	if err != nil {
		t.Fatalf("ParseNews failed: %v", err)
	}
	if len(articles) == 0 {
		t.Fatal("expected at least 1 article from JSON path, got 0")
	}
	if stats == nil {
		t.Fatal("stats is nil")
	}
	if stats.TotalReturned == 0 {
		t.Error("TotalReturned = 0; expected >0")
	}
	// First article should have title and URL extracted from the embedded
	// tickerStream JSON. The story-block regex stops at the first nested
	// `}` after `"STORY"`, so we only assert fields reachable in that span
	// (title + canonicalUrl), not stockTickers/provider which sit later in
	// the block.
	if articles[0].Title == "" {
		t.Error("first article Title is empty")
	}
	if articles[0].URL != "https://finance.yahoo.com/news/apple-product-123.html" {
		t.Errorf("URL = %q, want fixture URL", articles[0].URL)
	}
}