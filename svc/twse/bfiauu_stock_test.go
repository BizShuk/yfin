// bfiauu_stock_test.go — `FetchBFIAUUSTOCK` stockNo-required validation + decode verifying stockNo in URL + ParseBFIAUUSTOCKRow row too-short. Capacity: 3 test cases.
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

func TestFetchBFIAUUSTOCK_RequiresStockNo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when stockNo is missing")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchBFIAUUSTOCK(context.Background(), client, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error when stockNo is missing, got nil")
	}
	if !strings.Contains(err.Error(), "stockNo") {
		t.Fatalf("expected error to mention stockNo, got: %v", err)
	}
}

func TestFetchBFIAUUSTOCK_Decode(t *testing.T) {
	var seenURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "單一證券鉅額交易",
			"fields": []string{"序號", "證券代號", "證券名稱", "買進證券商", "賣出證券商", "成交數量", "成交金額", "成交價格", "成交時間", "買進成交價"},
			"data": [][]string{
				{"1", "2330", "台積電", "元大", "凱基", "100,000", "60,000,000", "600.00", "13:30:00", "599.50"},
			},
			"date":    "20260620",
			"stockNo": "2330",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	opts := url.Values{}
	opts.Set("stockNo", "2330")
	raw, err := FetchBFIAUUSTOCK(context.Background(), client, "20260620", opts)
	if err != nil {
		t.Fatalf("FetchBFIAUUSTOCK returned error: %v", err)
	}
	resp, ok := raw.(BlockBFIAUUResponse)
	if !ok {
		t.Fatalf("expected BlockBFIAUUResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if !strings.Contains(seenURL, "stockNo=2330") {
		t.Errorf("expected URL to contain stockNo=2330, got %q", seenURL)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseBFIAUUSTOCKRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseBFIAUUSTOCKRow returned error: %v", err)
	}
	if row.StockCode != "2330" {
		t.Errorf("StockCode = %q, want 2330", row.StockCode)
	}
	if row.StockName != "台積電" {
		t.Errorf("StockName = %q", row.StockName)
	}
	if row.TradeVolume != 100000 {
		t.Errorf("TradeVolume = %v, want 100000", row.TradeVolume)
	}
	if row.TradeAmount != 60000000 {
		t.Errorf("TradeAmount = %v, want 60000000", row.TradeAmount)
	}
}

func TestParseBFIAUUSTOCKRow_TooShort(t *testing.T) {
	_, err := ParseBFIAUUSTOCKRow([]string{"1"})
	if err == nil {
		t.Fatal("expected error for short row, got nil")
	}
}
