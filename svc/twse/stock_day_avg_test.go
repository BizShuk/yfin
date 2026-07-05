// stock_day_avg_test.go — `FetchStockDayAvg` date-required + stockNo-required validation + decode (date + stockNo in URL) + ParseStockDayAvgRow row too-short. Capacity: 4 test cases.
package twse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFetchStockDayAvg_RequiresDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when date is missing")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	newTestClient(t, srv)
	_, err := FetchStockDayAvg(context.Background(), "", url.Values{"stockNo": []string{"2330"}})
	if err == nil {
		t.Fatal("expected error when date is missing, got nil")
	}
	if !strings.Contains(err.Error(), "date") {
		t.Fatalf("expected error to mention date, got: %v", err)
	}
}

func TestFetchStockDayAvg_RequiresStockNo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when stockNo is missing")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	newTestClient(t, srv)
	_, err := FetchStockDayAvg(context.Background(), "20250601", url.Values{})
	if err == nil {
		t.Fatal("expected error when stockNo is missing, got nil")
	}
	if !strings.Contains(err.Error(), "stockNo") {
		t.Fatalf("expected error to mention stockNo, got: %v", err)
	}
}

func TestFetchStockDayAvg_Decode(t *testing.T) {
	var seenURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "個股月均價",
			"fields": []string{"年度", "月份", "最高", "最低", "加權平均價", "成交筆數", "成交股數", "成交金額"},
			"data": [][]string{
				{"114", "06", "1,000", "950", "975.50", "500,000", "100,000,000", "97,550,000,000"},
			},
			"date":    "20250601",
			"stockNo": "2330",
		})
	}))
	defer srv.Close()

	newTestClient(t, srv)
	opts := url.Values{}
	opts.Set("stockNo", "2330")
	raw, err := FetchStockDayAvg(context.Background(), "20250601", opts)
	if err != nil {
		t.Fatalf("FetchStockDayAvg returned error: %v", err)
	}
	resp, ok := raw.(StockDayAvgResponse)
	if !ok {
		t.Fatalf("expected *StockDayAvgResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if !strings.Contains(seenURL, "date=20250601") {
		t.Errorf("expected URL to contain date=20250601, got %q", seenURL)
	}
	if !strings.Contains(seenURL, "stockNo=2330") {
		t.Errorf("expected URL to contain stockNo=2330, got %q", seenURL)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseStockDayAvgRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseStockDayAvgRow returned error: %v", err)
	}
	if row.Year != "114" {
		t.Errorf("Year = %q, want %q", row.Year, "114")
	}
	if row.Month != "06" {
		t.Errorf("Month = %q, want %q", row.Month, "06")
	}
	if row.High != 1000 {
		t.Errorf("High = %v, want 1000", row.High)
	}
	if row.Low != 950 {
		t.Errorf("Low = %v, want 950", row.Low)
	}
	if row.WeightedAvg != 975.50 {
		t.Errorf("WeightedAvg = %v, want 975.50", row.WeightedAvg)
	}
	if row.Transactions != 500000 {
		t.Errorf("Transactions = %d, want 500000", row.Transactions)
	}
	if row.Volume != 100000000 {
		t.Errorf("Volume = %d, want 100000000", row.Volume)
	}
	if row.Amount != 97550000000 {
		t.Errorf("Amount = %d, want 97550000000", row.Amount)
	}
}

func TestParseStockDayAvgRow_TooShort(t *testing.T) {
	_, err := ParseStockDayAvgRow([]string{"114", "06"})
	if err == nil {
		t.Fatal("expected error for short row, got nil")
	}
}
