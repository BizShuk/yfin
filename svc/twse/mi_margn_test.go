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

func TestFetchMI_MARGN_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/marginTrading/MI_MARGN") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("selectType") != "ALL" {
			t.Errorf("expected selectType=ALL, got %q", r.URL.Query().Get("selectType"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":  "OK",
			"title": "融資融券餘額",
			"fields": []string{"股票代號", "股票名稱", "融資買進", "融資賣出", "融資現償", "融資餘額", "融券買進", "融券賣出", "融券現償", "融券餘額"},
			"data": [][]string{
				{"2330", "台積電", "1,000", "2,000", "300", "5,000", "400", "500", "100", "2,000"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	raw, err := FetchMI_MARGN(context.Background(), c, "20260620", url.Values{})
	if err != nil {
		t.Fatalf("FetchMI_MARGN returned error: %v", err)
	}
	resp, ok := raw.(*MI_MARGNResponse)
	if !ok {
		t.Fatalf("expected *MI_MARGNResponse, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseMI_MARGNRow(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseMI_MARGNRow returned error: %v", err)
	}
	if row.Code != "2330" {
		t.Errorf("Code = %q, want %q", row.Code, "2330")
	}
	if row.Name != "台積電" {
		t.Errorf("Name = %q, want %q", row.Name, "台積電")
	}
	if row.MarginBuy != 1000 {
		t.Errorf("MarginBuy = %d, want %d", row.MarginBuy, 1000)
	}
	if row.MarginSell != 2000 {
		t.Errorf("MarginSell = %d, want %d", row.MarginSell, 2000)
	}
	if row.MarginRepay != 300 {
		t.Errorf("MarginRepay = %d, want %d", row.MarginRepay, 300)
	}
	if row.MarginBalance != 5000 {
		t.Errorf("MarginBalance = %d, want %d", row.MarginBalance, 5000)
	}
	if row.ShortBuy != 400 {
		t.Errorf("ShortBuy = %d, want %d", row.ShortBuy, 400)
	}
	if row.ShortSell != 500 {
		t.Errorf("ShortSell = %d, want %d", row.ShortSell, 500)
	}
	if row.ShortRepay != 100 {
		t.Errorf("ShortRepay = %d, want %d", row.ShortRepay, 100)
	}
	if row.ShortBalance != 2000 {
		t.Errorf("ShortBalance = %d, want %d", row.ShortBalance, 2000)
	}
}

func TestFetchMI_MARGN_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := FetchMI_MARGN(context.Background(), c, "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}