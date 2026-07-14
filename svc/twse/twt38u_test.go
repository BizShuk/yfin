// twt38u_test.go — `FetchTWT38U` decode verifying path/date/response + 4-column foreign-summary row + no-data. Capacity: 2 test cases.
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

func TestFetchTWT38U_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/fund/TWT38U") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("date") != "20260620" {
			t.Errorf("missing/wrong date param: %s", r.URL.RawQuery)
		}
		if r.URL.Query().Get("response") != "json" {
			t.Errorf("missing response=json: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "外資陸資買賣超彙總",
			"fields": []string{"單位名稱", "買進股數", "賣出股數", "買賣差額股數"},
			"data": [][]string{
				{"外資及陸資", "1,234,567,890", "987,654,321", "246,913,569"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchTWT38U(context.Background(), client, "20260620", nil)
	if err != nil {
		t.Fatalf("FetchTWT38U returned error: %v", err)
	}
	resp, ok := raw.(model.TWT38UResponse)
	if !ok {
		t.Fatalf("expected *model.TWT38UResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if resp.Date != "20260620" {
		t.Errorf("Date = %q, want 20260620", resp.Date)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseTWT38URow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseTWT38URow returned error: %v", err)
	}
	if row.UnitName != "外資及陸資" {
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

func TestFetchTWT38U_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchTWT38U(context.Background(), client, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
