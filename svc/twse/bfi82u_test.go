// bfi82u_test.go — `FetchBFI82U` decode verifying path + date + type=day + response=json + 3-institution rows + no-data. Capacity: 2 test cases.
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

func TestFetchBFI82U_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path and required params
		if !strings.Contains(r.URL.Path, "/fund/BFI82U") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20260620" {
			t.Errorf("missing/wrong date param: %s", r.URL.RawQuery)
		}
		if r.URL.Query().Get("type") != "day" {
			t.Errorf("missing type=day param: %s", r.URL.RawQuery)
		}
		if r.URL.Query().Get("response") != "json" {
			t.Errorf("missing response=json: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "三大法人買賣金額",
			"fields": []string{"單位名稱", "買進金額", "賣出金額", "買賣差額"},
			"data": [][]string{
				{"自營商(自行買賣)", "1,234,567,890", "987,654,321", "246,913,569"},
				{"投信", "555,555,555", "444,444,444", "111,111,111"},
				{"外資及陸資", "9,999,999,999", "8,888,888,888", "1,111,111,111"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchBFI82U(context.Background(), client, "20260620", nil)
	if err != nil {
		t.Fatalf("FetchBFI82U returned error: %v", err)
	}
	resp, ok := raw.(BFI82UResponse)
	if !ok {
		t.Fatalf("expected *BFI82UResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if resp.Date != "20260620" {
		t.Errorf("Date = %q, want 20260620", resp.Date)
	}
	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 data rows, got %d", len(resp.Data))
	}
	row, err := ParseBFI82URow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseBFI82URow returned error: %v", err)
	}
	if row.UnitName != "自營商(自行買賣)" {
		t.Errorf("UnitName = %q", row.UnitName)
	}
	if row.Buy != 1234567890 {
		t.Errorf("Buy = %v, want 1234567890", row.Buy)
	}
	if row.Sell != 987654321 {
		t.Errorf("Sell = %v, want 987654321", row.Sell)
	}
	if row.Net != 246913569 {
		t.Errorf("Net = %v, want 246913569", row.Net)
	}
}

func TestFetchBFI82U_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchBFI82U(context.Background(), client, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
