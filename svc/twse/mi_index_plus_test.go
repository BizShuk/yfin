// mi_index_plus_test.go — `FetchMI_INDEX_PLUS` decode + no-data `ErrNoData`. Capacity: 2 test cases.
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

func TestFetchMI_INDEX_PLUS_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "盤後定價交易",
			"fields": []string{"指數", "收盤指數", "漲跌點數", "漲跌百分比"},
			"data": [][]string{
				{"發行量加權股價指數", "17,512.45", "+12.33", "+0.07%"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchMI_INDEX_PLUS(context.Background(), client, "20260620", url.Values{})
	if err != nil {
		t.Fatalf("FetchMI_INDEX_PLUS returned error: %v", err)
	}
	resp, ok := raw.(MI_INDEX_PLUSResponse)
	if !ok {
		t.Fatalf("expected MI_INDEX_PLUSResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseMIIndexPlusRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseMIIndexPlusRow returned error: %v", err)
	}
	if row.IndexName != "發行量加權股價指數" {
		t.Errorf("IndexName = %q, want %q", row.IndexName, "發行量加權股價指數")
	}
	if row.Close != 17512.45 {
		t.Errorf("Close = %v, want 17512.45", row.Close)
	}
	if row.Change != 12.33 {
		t.Errorf("Change = %v, want 12.33", row.Change)
	}
	if row.ChangePct != 0.07 {
		t.Errorf("ChangePct = %v, want 0.07", row.ChangePct)
	}
}

func TestFetchMI_INDEX_PLUS_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchMI_INDEX_PLUS(context.Background(), client, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
