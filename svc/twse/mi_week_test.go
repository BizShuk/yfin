// mi_week_test.go — `FetchMI_WEEK` decode verifying path/date + multi-row stock market-cap + no-data. Capacity: 2 test cases.
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

func TestFetchMI_WEEK_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/statistics/MI_WEEK") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20221230" {
			t.Errorf("expected date=20221230, got %q", r.URL.Query().Get("date"))
		}
		if r.URL.Query().Get("response") != "json" {
			t.Errorf("expected response=json, got %q", r.URL.Query().Get("response"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "股票市值週報",
			"fields": []string{"股票代號", "股票名稱", "發行股數", "市值"},
			"data": [][]string{
				{"2330", "台積電", "25,930,380,000", "16,118,786,190,000"},
				{"2317", "鴻海", "14,000,000,000", "1,400,000,000,000"},
			},
			"date": "20221230",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchMI_WEEK(context.Background(), client, "20221230", url.Values{})
	if err != nil {
		t.Fatalf("FetchMI_WEEK returned error: %v", err)
	}
	resp, ok := raw.(MI_WEEKResponse)
	if !ok {
		t.Fatalf("expected MI_WEEKResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if resp.Date != "20221230" {
		t.Errorf("Date = %q, want 20221230", resp.Date)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 data rows, got %d", len(resp.Data))
	}

	row, err := ParseMIWeekRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseMIWeekRow returned error: %v", err)
	}
	if row.StockCode != "2330" {
		t.Errorf("StockCode = %q, want 2330", row.StockCode)
	}
	if row.StockName != "台積電" {
		t.Errorf("StockName = %q, want 台積電", row.StockName)
	}
	if row.SharesIssued != 25930380000 {
		t.Errorf("SharesIssued = %v, want 25930380000", row.SharesIssued)
	}
	if row.MarketCap != 16118786190000 {
		t.Errorf("MarketCap = %v, want 16118786190000", row.MarketCap)
	}

	row2, err := ParseMIWeekRow(resp.Data[1])
	if err != nil {
		t.Fatalf("ParseMIWeekRow[1] returned error: %v", err)
	}
	if row2.StockName != "鴻海" {
		t.Errorf("StockName = %q, want 鴻海", row2.StockName)
	}
}

func TestFetchMI_WEEK_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchMI_WEEK(context.Background(), client, "19000101", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
