// fmsrfk_test.go — `FetchFMSRFK` decode verifying path/stockNo=2330/date=year + multi-row monthly aggregate + no-data. Capacity: 2 test cases.
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

func TestFetchFMSRFK_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/exchangeReport/FMSRFK") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("stockNo") != "2330" {
			t.Errorf("expected stockNo=2330, got %q", r.URL.Query().Get("stockNo"))
		}
		if r.URL.Query().Get("date") != "2022" {
			t.Errorf("expected date=2022, got %q", r.URL.Query().Get("date"))
		}
		if r.URL.Query().Get("response") != "json" {
			t.Errorf("expected response=json, got %q", r.URL.Query().Get("response"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "個股月成交資訊",
			"fields": []string{"年度", "月份", "最高", "最低", "加權平均價", "成交股數", "成交金額", "週轉率%"},
			"data": [][]string{
				{"2022", "01", "688.00", "600.00", "643.21", "1,234,567,890", "789,012,345,678", "12.34"},
				{"2022", "02", "680.00", "610.00", "650.10", "987,654,321", "642,135,792,468", "10.20"},
			},
			"stockNo": "2330",
			"date":    "2022",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchFMSRFK(context.Background(), client, "2330", "2022", url.Values{})
	if err != nil {
		t.Fatalf("FetchFMSRFK returned error: %v", err)
	}
	resp, ok := raw.(FMSRFKResponse)
	if !ok {
		t.Fatalf("expected FMSRFKResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if resp.StockNo != "2330" {
		t.Errorf("StockNo = %q, want 2330", resp.StockNo)
	}
	if resp.Date != "2022" {
		t.Errorf("Date = %q, want 2022", resp.Date)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 data rows, got %d", len(resp.Data))
	}

	row, err := ParseFMSRFKRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseFMSRFKRow returned error: %v", err)
	}
	if row.Year != "2022" {
		t.Errorf("Year = %q, want 2022", row.Year)
	}
	if row.Month != "01" {
		t.Errorf("Month = %q, want 01", row.Month)
	}
	if row.High != 688.00 {
		t.Errorf("High = %v, want 688.00", row.High)
	}
	if row.Low != 600.00 {
		t.Errorf("Low = %v, want 600.00", row.Low)
	}
	if row.WAvgPrice != 643.21 {
		t.Errorf("WAvgPrice = %v, want 643.21", row.WAvgPrice)
	}
	if row.TradeVolume != 1234567890 {
		t.Errorf("TradeVolume = %v, want 1234567890", row.TradeVolume)
	}
	if row.TradeValue != 789012345678 {
		t.Errorf("TradeValue = %v, want 789012345678", row.TradeValue)
	}
	if row.TurnoverPct != 12.34 {
		t.Errorf("TurnoverPct = %v, want 12.34", row.TurnoverPct)
	}
}

func TestFetchFMSRFK_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchFMSRFK(context.Background(), client, "9999", "1900", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
