// bfiamu_test.go — `FetchBFIAMU` decode verifying path/date/response=json + multi-row index close/change with negative change-pct coverage + no-data. Capacity: 2 test cases.
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

func TestFetchBFIAMU_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/afterTrading/BFIAMU") {
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
			"title":  "每日各類指數成交量值",
			"fields": []string{"指數", "收盤指數", "漲跌", "百分比"},
			"data": [][]string{
				{"發行量加權股價指數", "14,179.45", "+100.21", "+0.71%"},
				{"臺灣50指數", "12,345.67", "-50.00", "-0.40%"},
			},
			"date": "20221230",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchBFIAMU(context.Background(), client, "20221230", url.Values{})
	if err != nil {
		t.Fatalf("FetchBFIAMU returned error: %v", err)
	}
	resp, ok := raw.(BFIAMUResponse)
	if !ok {
		t.Fatalf("expected BFIAMUResponse, got %T", raw)
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

	row, err := ParseBFIAMURow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseBFIAMURow returned error: %v", err)
	}
	if row.IndexName != "發行量加權股價指數" {
		t.Errorf("IndexName = %q, want 發行量加權股價指數", row.IndexName)
	}
	if row.Close != 14179.45 {
		t.Errorf("Close = %v, want 14179.45", row.Close)
	}
	if row.Change != 100.21 {
		t.Errorf("Change = %v, want 100.21", row.Change)
	}
	if row.ChangePct != 0.71 {
		t.Errorf("ChangePct = %v, want 0.71", row.ChangePct)
	}

	row2, err := ParseBFIAMURow(resp.Data[1])
	if err != nil {
		t.Fatalf("ParseBFIAMURow[1] returned error: %v", err)
	}
	if row2.ChangePct != -0.40 {
		t.Errorf("ChangePct = %v, want -0.40", row2.ChangePct)
	}
}

func TestFetchBFIAMU_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchBFIAMU(context.Background(), client, "19000101", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
