// fmtqik_test.go — `FetchFMTQIK` date-required validation + decode verifying date in URL + ParseFMTQIKRow row too-short. Capacity: 3 test cases.
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

func TestFetchFMTQIK_RequiresDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when date is missing")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchFMTQIK(context.Background(), client, "", url.Values{})
	if err == nil {
		t.Fatal("expected error when date is missing, got nil")
	}
	if !strings.Contains(err.Error(), "date") {
		t.Fatalf("expected error to mention date, got: %v", err)
	}
}

func TestFetchFMTQIK_Decode(t *testing.T) {
	var seenURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "臺股指數及交易量表",
			"fields": []string{"日期", "成交股數", "成交金額", "成交筆數", "發行量加權股價指數"},
			"data": [][]string{
				{"114/06/20", "8,000,000,000", "300,000,000,000", "2,000,000", "18,000.50"},
			},
			"date": "20250620",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchFMTQIK(context.Background(), client, "20250620", url.Values{})
	if err != nil {
		t.Fatalf("FetchFMTQIK returned error: %v", err)
	}
	resp, ok := raw.(model.FMTQIKResponse)
	if !ok {
		t.Fatalf("expected *model.FMTQIKResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if !strings.Contains(seenURL, "date=20250620") {
		t.Errorf("expected URL to contain date=20250620, got %q", seenURL)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseFMTQIKRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseFMTQIKRow returned error: %v", err)
	}
	if row.Date != "114/06/20" {
		t.Errorf("Date = %q, want %q", row.Date, "114/06/20")
	}
	if row.Volume != 8000000000 {
		t.Errorf("Volume = %d, want 8000000000", row.Volume)
	}
	if row.Amount != 300000000000 {
		t.Errorf("Amount = %d, want 300000000000", row.Amount)
	}
	if row.Transactions != 2000000 {
		t.Errorf("Transactions = %d, want 2000000", row.Transactions)
	}
	if row.Index != 18000.50 {
		t.Errorf("Index = %v, want 18000.50", row.Index)
	}
}

func TestParseFMTQIKRow_TooShort(t *testing.T) {
	_, err := ParseFMTQIKRow([]string{"114/06/20", "8,000,000,000"})
	if err == nil {
		t.Fatal("expected error for short row, got nil")
	}
}
