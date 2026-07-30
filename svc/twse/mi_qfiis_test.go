// mi_qfiis_test.go — `FetchMI_QFIIS` decode verifying the live TWSE
// all-category selector + foreign-holding row + no-data. Capacity: 2 test cases.
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

func TestFetchMI_QFIIS_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/fund/MI_QFIIS") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("selectType") != "ALLBUT0999" {
			t.Errorf("expected selectType=ALLBUT0999, got %q", r.URL.Query().Get("selectType"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":  "OK",
			"title": "外資及陸資投資持股統計",
			"fields": []string{
				"證券代號",
				"證券名稱",
				"國際證券編碼",
				"發行股數",
				"外資及陸資尚可投資股數",
				"全體外資及陸資持有股數",
				"外資及陸資尚可投資比率",
				"全體外資及陸資持股比率",
				"外資及陸資共用法令投資上限比率",
				"陸資法令投資上限比率",
				"與前日異動原因(註)",
				"最近一次上市公司申報外資及陸資持股異動日期",
			},
			"data": [][]any{
				{
					"2330",
					"台積電",
					"TW0002330008",
					"25,932,071,458",
					"6,000,000,000",
					"19,932,071,458",
					23.14,
					76.86,
					"100.00",
					"100.00",
					"",
					"115/07/29",
				},
			},
			"notes": "一、與前日異動原因說明。",
			"date":  "20260620",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	raw, err := FetchMI_QFIIS(context.Background(), client, "20260620", url.Values{})
	if err != nil {
		t.Fatalf("FetchMI_QFIIS returned error: %v", err)
	}
	resp, ok := raw.(model.MI_QFIISResponse)
	if !ok {
		t.Fatalf("expected model.MI_QFIISResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	if len(resp.Notes) != 1 || resp.Notes[0] != "一、與前日異動原因說明。" {
		t.Fatalf("Notes = %#v, want one live note string", resp.Notes)
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
	if row.ISIN != "TW0002330008" {
		t.Errorf("ISIN = %q, want %q", row.ISIN, "TW0002330008")
	}
	if row.IssuedShares != 25932071458 {
		t.Errorf("IssuedShares = %d, want %d", row.IssuedShares, 25932071458)
	}
	if row.AvailableShares != 6000000000 {
		t.Errorf("AvailableShares = %d, want %d", row.AvailableShares, 6000000000)
	}
	if row.SharesHeld != 19932071458 {
		t.Errorf("SharesHeld = %d, want %d", row.SharesHeld, 19932071458)
	}
	if row.AvailablePct != 23.14 {
		t.Errorf("AvailablePct = %v, want 23.14", row.AvailablePct)
	}
	if row.IssuePct != 76.86 {
		t.Errorf("IssuePct = %v, want 76.86", row.IssuePct)
	}
	if row.ForeignMainlandLimitPct != 100 {
		t.Errorf("ForeignMainlandLimitPct = %v, want 100", row.ForeignMainlandLimitPct)
	}
	if row.MainlandLimitPct != 100 {
		t.Errorf("MainlandLimitPct = %v, want 100", row.MainlandLimitPct)
	}
	if row.ChangeReason != "" {
		t.Errorf("ChangeReason = %q, want empty", row.ChangeReason)
	}
	if row.LastReportedDate != "115/07/29" {
		t.Errorf("LastReportedDate = %q, want %q", row.LastReportedDate, "115/07/29")
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
