// mi_5mins_test.go — `FetchMI_5MINS` decode verifying request path + 7-column cumulative row + no-data. Capacity: 2 test cases.
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

func TestFetchMI_5MINS_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/afterTrading/MI_5MINS") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "每5秒委託成交統計",
			"fields": []string{"時間", "累積委買筆數", "累積委買張數", "累積委賣筆數", "累積委賣張數", "累計成交筆數", "累計成交張數"},
			"data": [][]string{
				{"13:30:00", "12,345", "67,890", "11,111", "65,432", "9,876", "54,321"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	newTestClient(t, srv)
	raw, err := FetchMI_5MINS(context.Background(), "20260620", url.Values{})
	if err != nil {
		t.Fatalf("FetchMI_5MINS returned error: %v", err)
	}
	resp, ok := raw.(MI_5MINSResponse)
	if !ok {
		t.Fatalf("expected MI_5MINSResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseMI_5MINSRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseMI_5MINSRow returned error: %v", err)
	}
	if row.Time != "13:30:00" {
		t.Errorf("Time = %q, want %q", row.Time, "13:30:00")
	}
	if row.CumBuyOrders != 12345 {
		t.Errorf("CumBuyOrders = %d, want %d", row.CumBuyOrders, 12345)
	}
	if row.CumBuyLots != 67890 {
		t.Errorf("CumBuyLots = %d, want %d", row.CumBuyLots, 67890)
	}
	if row.CumSellOrders != 11111 {
		t.Errorf("CumSellOrders = %d, want %d", row.CumSellOrders, 11111)
	}
	if row.CumSellLots != 65432 {
		t.Errorf("CumSellLots = %d, want %d", row.CumSellLots, 65432)
	}
	if row.CumTradeOrders != 9876 {
		t.Errorf("CumTradeOrders = %d, want %d", row.CumTradeOrders, 9876)
	}
	if row.CumTradeLots != 54321 {
		t.Errorf("CumTradeLots = %d, want %d", row.CumTradeLots, 54321)
	}
}

func TestFetchMI_5MINS_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	newTestClient(t, srv)
	_, err := FetchMI_5MINS(context.Background(), "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
