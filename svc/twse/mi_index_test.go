package twse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AmpyFin/yfinance-go/internal/httpx"
)

// newTestClient returns a Caller backed by an httpx client wired to the
// httptest server. It also overrides BaseURL so FetchJSON's internal
// logic and the Caller's URL builder both point at the test server.
func newTestClient(t *testing.T, srv *httptest.Server) Caller {
	t.Helper()
	cfg := httpx.DefaultConfig()
	cfg.Timeout = 5_000_000_000 // 5s in ns
	cfg.MaxAttempts = 1
	cfg.BaseURL = srv.URL
	c := httpx.NewClient(cfg)
	old := BaseURL
	BaseURL = srv.URL
	t.Cleanup(func() { BaseURL = old })
	return NewHttpxCaller(c)
}

func TestFetchMI_INDEX_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":  "OK",
			"title": "每日收盤行情",
			"fields": []string{"指數", "收盤指數", "漲跌點數", "漲跌百分比"},
			"data": [][]string{
				{"發行量加權股價指數", "17,500.12", "+120.34", "+0.69%"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	opts := url.Values{}
	opts.Set("type", "ALL")
	raw, err := FetchMI_INDEX(context.Background(), c, "20260620", opts)
	if err != nil {
		t.Fatalf("FetchMI_INDEX returned error: %v", err)
	}
	resp, ok := raw.(MI_INDEXResponse)
	if !ok {
		t.Fatalf("expected MI_INDEXResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseMIIndexRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseMIIndexRow returned error: %v", err)
	}
	if row.IndexName != "發行量加權股價指數" {
		t.Errorf("IndexName = %q, want %q", row.IndexName, "發行量加權股價指數")
	}
	if row.Close != 17500.12 {
		t.Errorf("Close = %v, want 17500.12", row.Close)
	}
	if row.Change != 120.34 {
		t.Errorf("Change = %v, want 120.34", row.Change)
	}
	if row.ChangePct != 0.69 {
		t.Errorf("ChangePct = %v, want 0.69", row.ChangePct)
	}
}

func TestFetchMI_INDEX_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := FetchMI_INDEX(context.Background(), c, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
