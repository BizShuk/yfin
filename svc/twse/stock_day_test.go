// stock_day_test.go — `FetchSTOCK_DAY` decode + missing-stockNo validation + no-data `ErrNoData`. Capacity: 3 test cases (decode, missing arg, no-data).
package twse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"github.com/bizshuk/yfin/model"
)

func TestFetchSTOCK_DAY_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "個股日成交資訊",
			"fields": []string{"日期", "成交股數", "成交金額", "開盤", "最高", "最低", "收盤", "漲跌價差", "成交筆數"},
			"data": [][]string{
				{"20260620", "1,234,567", "12,345,678", "100.00", "101.50", "99.50", "101.00", "+1.00", "5,678"},
			},
			"date":    "20260620",
			"stockNo": "2330",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	opts := url.Values{}
	opts.Set("stockNo", "2330")
	raw, err := FetchSTOCK_DAY(context.Background(), client, "20260620", opts)
	if err != nil {
		t.Fatalf("FetchSTOCK_DAY returned error: %v", err)
	}
	resp, ok := raw.(model.STOCK_DAYResponse)
	if !ok {
		t.Fatalf("expected model.STOCK_DAYResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseStockDayRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseStockDayRow returned error: %v", err)
	}
	if row.Date != "20260620" {
		t.Errorf("Date = %q, want %q", row.Date, "20260620")
	}
	if row.Volume != 1234567 {
		t.Errorf("Volume = %d, want 1234567", row.Volume)
	}
	if row.Open != 100.00 {
		t.Errorf("Open = %v, want 100.00", row.Open)
	}
	if row.Close != 101.00 {
		t.Errorf("Close = %v, want 101.00", row.Close)
	}
	if row.Transactions != 5678 {
		t.Errorf("Transactions = %d, want 5678", row.Transactions)
	}
}

func TestFetchSTOCK_DAY_MissingStockNo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when stockNo is missing")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchSTOCK_DAY(context.Background(), client, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for missing stockNo, got nil")
	}
	if !strings.Contains(err.Error(), "stockNo is required") {
		t.Fatalf("expected error containing 'stockNo is required', got: %v", err)
	}
}

func TestFetchSTOCK_DAY_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	opts := url.Values{}
	opts.Set("stockNo", "2330")
	_, err := FetchSTOCK_DAY(context.Background(), client, "20260620", opts)
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
