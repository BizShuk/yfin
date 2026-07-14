// mi_index_odd_test.go — `FetchMI_INDEX_ODD` + `ParseMIIndexOddRow` decode covering full 8-column odd-lot row + no-data. Capacity: 2 test cases.
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

func TestFetchMI_INDEX_ODD_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "零股交易行情單",
			"fields": []string{"證券代號", "證券名稱", "成交股數", "成交金額", "開盤", "最高", "最低", "收盤"},
			"data": [][]string{
				{"2330", "台積電", "12,345", "1,234,500", "100.00", "101.00", "99.50", "100.50"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchMI_INDEX_ODD(context.Background(), client, "20260620", url.Values{})
	if err != nil {
		t.Fatalf("FetchMI_INDEX_ODD returned error: %v", err)
	}
	resp, ok := raw.(model.MI_INDEX_ODDResponse)
	if !ok {
		t.Fatalf("expected model.MI_INDEX_ODDResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseMIIndexOddRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseMIIndexOddRow returned error: %v", err)
	}
	if row.Code != "2330" {
		t.Errorf("Code = %q, want %q", row.Code, "2330")
	}
	if row.Name != "台積電" {
		t.Errorf("Name = %q, want %q", row.Name, "台積電")
	}
	if row.Volume != 12345 {
		t.Errorf("Volume = %d, want 12345", row.Volume)
	}
	if row.Amount != 1234500 {
		t.Errorf("Amount = %d, want 1234500", row.Amount)
	}
	if row.Open != 100.00 {
		t.Errorf("Open = %v, want 100.00", row.Open)
	}
	if row.High != 101.00 {
		t.Errorf("High = %v, want 101.00", row.High)
	}
	if row.Low != 99.50 {
		t.Errorf("Low = %v, want 99.50", row.Low)
	}
	if row.Close != 100.50 {
		t.Errorf("Close = %v, want 100.50", row.Close)
	}
}

func TestFetchMI_INDEX_ODD_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchMI_INDEX_ODD(context.Background(), client, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
