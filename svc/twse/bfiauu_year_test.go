// bfiauu_year_test.go — `FetchBFIAUUYEAR` date-required validation + decode verifying date=YYYY0101 in URL + row too-short. Capacity: 3 test cases.
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

func TestFetchBFIAUUYEAR_RequiresDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when date is missing")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchBFIAUUYEAR(context.Background(), client, "", url.Values{})
	if err == nil {
		t.Fatal("expected error when date is missing, got nil")
	}
	if !strings.Contains(err.Error(), "date") {
		t.Fatalf("expected error to mention date, got: %v", err)
	}
}

func TestFetchBFIAUUYEAR_Decode(t *testing.T) {
	var seenURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "鉅額交易年成交資訊",
			"fields": []string{"年度", "成交筆數", "成交股數", "成交金額"},
			"data": [][]string{
				{"114", "12,345", "67,890,123", "1,234,567,890,123"},
			},
			"date": "20250101",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchBFIAUUYEAR(context.Background(), client, "20250101", url.Values{})
	if err != nil {
		t.Fatalf("FetchBFIAUUYEAR returned error: %v", err)
	}
	resp, ok := raw.(BFIAUUYEARResponse)
	if !ok {
		t.Fatalf("expected *BFIAUUYEARResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if !strings.Contains(seenURL, "date=20250101") {
		t.Errorf("expected URL to contain date=20250101, got %q", seenURL)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseBFIAUUYEARRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseBFIAUUYEARRow returned error: %v", err)
	}
	if row.Year != "114" {
		t.Errorf("Year = %q, want %q", row.Year, "114")
	}
	if row.Transactions != 12345 {
		t.Errorf("Transactions = %d, want 12345", row.Transactions)
	}
	if row.Volume != 67890123 {
		t.Errorf("Volume = %d, want 67890123", row.Volume)
	}
	if row.Amount != 1234567890123 {
		t.Errorf("Amount = %d, want 1234567890123", row.Amount)
	}
}

func TestParseBFIAUUYEARRow_TooShort(t *testing.T) {
	_, err := ParseBFIAUUYEARRow([]string{"114", "12,345"})
	if err == nil {
		t.Fatal("expected error for short row, got nil")
	}
}
