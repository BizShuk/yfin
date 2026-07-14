// twtb4u_test.go — `FetchTWTB4U` decode (path/date/query verified) + no-data + empty-date validation. Capacity: 3 test cases.
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

func TestFetchTWTB4U_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/afterTrading/TWTB4U") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20260620" {
			t.Errorf("unexpected date param: %s", r.URL.Query().Get("date"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "當日沖銷交易標的",
			"fields": []string{"證券代號", "證券名稱", "當日沖銷交易成交股數", "當日沖銷交易成交金額", "當日沖銷交易買進成交金額", "當日沖銷交易賣出成交金額"},
			"data": [][]string{
				{"2330", "台積電", "12,345,678", "98,765,432", "49,000,000", "49,765,432"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchTWTB4U(context.Background(), client, "20260620", url.Values{})
	if err != nil {
		t.Fatalf("FetchTWTB4U returned error: %v", err)
	}
	resp, ok := raw.(model.TWTB4UResponse)
	if !ok {
		t.Fatalf("expected model.TWTB4UResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if resp.Date != "20260620" {
		t.Errorf("Date = %q, want %q", resp.Date, "20260620")
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseTWTB4URow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseTWTB4URow returned error: %v", err)
	}
	if row.Code != "2330" {
		t.Errorf("Code = %q, want %q", row.Code, "2330")
	}
	if row.Name != "台積電" {
		t.Errorf("Name = %q, want %q", row.Name, "台積電")
	}
	if row.TradeShares != 12345678 {
		t.Errorf("TradeShares = %d, want %d", row.TradeShares, 12345678)
	}
	if row.TradeAmount != 98765432 {
		t.Errorf("TradeAmount = %d, want %d", row.TradeAmount, 98765432)
	}
	if row.BuyAmount != 49000000 {
		t.Errorf("BuyAmount = %d, want %d", row.BuyAmount, 49000000)
	}
	if row.SellAmount != 49765432 {
		t.Errorf("SellAmount = %d, want %d", row.SellAmount, 49765432)
	}
}

func TestFetchTWTB4U_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchTWTB4U(context.Background(), client, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}

func TestFetchTWTB4U_EmptyDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	client := newTestClient(t, srv)
	_, err := FetchTWTB4U(context.Background(), client, "", url.Values{})
	if err == nil {
		t.Fatal("expected error for empty date, got nil")
	}
}
