// bfimuu_test.go — `FetchBFIMUU` date-required validation + decode verifying date=YYYYMM01 in URL + ParseBFIMUURow row too-short. Capacity: 3 test cases.
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

func TestFetchBFIMUU_RequiresDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when date is missing")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchBFIMUU(context.Background(), client, "", url.Values{})
	if err == nil {
		t.Fatal("expected error when date is missing, got nil")
	}
	if !strings.Contains(err.Error(), "date") {
		t.Fatalf("expected error to mention date, got: %v", err)
	}
}

func TestFetchBFIMUU_Decode(t *testing.T) {
	var seenURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "鉅額交易月成交資訊",
			"fields": []string{"年月", "成交筆數", "成交股數", "成交金額"},
			"data": [][]string{
				{"114年06月", "1,234", "5,678,900", "123,456,789,000"},
			},
			"date": "20250601",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchBFIMUU(context.Background(), client, "20250601", url.Values{})
	if err != nil {
		t.Fatalf("FetchBFIMUU returned error: %v", err)
	}
	resp, ok := raw.(model.BFIMUResponse)
	if !ok {
		t.Fatalf("expected *model.BFIMUResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if !strings.Contains(seenURL, "date=20250601") {
		t.Errorf("expected URL to contain date=20250601, got %q", seenURL)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseBFIMUURow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseBFIMUURow returned error: %v", err)
	}
	if row.Period != "114年06月" {
		t.Errorf("Period = %q, want %q", row.Period, "114年06月")
	}
	if row.Transactions != 1234 {
		t.Errorf("Transactions = %d, want 1234", row.Transactions)
	}
	if row.Volume != 5678900 {
		t.Errorf("Volume = %d, want 5678900", row.Volume)
	}
	if row.Amount != 123456789000 {
		t.Errorf("Amount = %d, want 123456789000", row.Amount)
	}
}

func TestParseBFIMUURow_TooShort(t *testing.T) {
	_, err := ParseBFIMUURow([]string{"114年06月", "1,234"})
	if err == nil {
		t.Fatal("expected error for short row, got nil")
	}
}
