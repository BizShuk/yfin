// mi_qfiis_test.go — `FetchMI_QFIIS` decode verifying selectType=ALL + 4-column foreign-holding row + no-data. Capacity: 2 test cases.
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

func TestFetchMI_QFIIS_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/fund/MI_QFIIS") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("selectType") != "ALL" {
			t.Errorf("expected selectType=ALL, got %q", r.URL.Query().Get("selectType"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "外資及陸資投資持股統計",
			"fields": []string{"證券代號", "證券名稱", "持有股數", "佔發行股數%"},
			"data": [][]string{
				{"2330", "台積電", "12,345,678,901", "75.55%"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchMI_QFIIS(context.Background(), client, "20260620", url.Values{})
	if err != nil {
		t.Fatalf("FetchMI_QFIIS returned error: %v", err)
	}
	resp, ok := raw.(MI_QFIISResponse)
	if !ok {
		t.Fatalf("expected MI_QFIISResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseMI_QFIISRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseMI_QFIISRow returned error: %v", err)
	}
	if row.Code != "2330" {
		t.Errorf("Code = %q, want %q", row.Code, "2330")
	}
	if row.Name != "台積電" {
		t.Errorf("Name = %q, want %q", row.Name, "台積電")
	}
	if row.SharesHeld != 12345678901 {
		t.Errorf("SharesHeld = %d, want %d", row.SharesHeld, 12345678901)
	}
	if row.IssuePct != 75.55 {
		t.Errorf("IssuePct = %v, want 75.55", row.IssuePct)
	}
}

func TestFetchMI_QFIIS_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := FetchMI_QFIIS(context.Background(), client, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
