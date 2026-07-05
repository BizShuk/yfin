// bwibbu_d_test.go — `FetchBWIBBU_d` + `ParseBWIBBUdRow` happy-path + no-data `ErrNoData`. Capacity: 2 test cases (decode + no-data).
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

func TestFetchBWIBBU_d_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "個股日本益比、殖利率及股價淨值比",
			"fields": []string{"證券代號", "證券名稱", "本益比", "殖利率(%)", "股價淨值比"},
			"data": [][]string{
				{"2330", "台積電", "22.5", "1.85", "5.6"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchBWIBBU_d(context.Background(), client, "20260620", url.Values{})
	if err != nil {
		t.Fatalf("FetchBWIBBU_d returned error: %v", err)
	}
	resp, ok := raw.(BWIBBU_dResponse)
	if !ok {
		t.Fatalf("expected BWIBBU_dResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseBWIBBUdRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseBWIBBUdRow returned error: %v", err)
	}
	if row.Code != "2330" {
		t.Errorf("Code = %q, want %q", row.Code, "2330")
	}
	if row.Name != "台積電" {
		t.Errorf("Name = %q, want %q", row.Name, "台積電")
	}
	if row.PE != 22.5 {
		t.Errorf("PE = %v, want 22.5", row.PE)
	}
	if row.YieldPct != 1.85 {
		t.Errorf("YieldPct = %v, want 1.85", row.YieldPct)
	}
	if row.PBR != 5.6 {
		t.Errorf("PBR = %v, want 5.6", row.PBR)
	}
}

func TestFetchBWIBBU_d_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchBWIBBU_d(context.Background(), client, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
