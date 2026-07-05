// t86_test.go — `FetchT86` decode verifying selectType=ALL + 12-column three-institution row + no-data. Capacity: 2 test cases.
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

func TestFetchT86_Decode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/fund/T86") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("selectType") != "ALL" {
			t.Errorf("expected selectType=ALL, got %q", r.URL.Query().Get("selectType"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "三大法人買賣超日報",
			"fields": []string{"證券代號", "證券名稱", "外陸資買進股數", "外陸資賣出股數", "外陸資買賣超股數", "投信買進股數", "投信賣出股數", "投信買賣超股數", "自營商買進股數", "自營商賣出股數", "自營商買賣超股數", "三大法人買賣超股數"},
			"data": [][]string{
				{"2330", "台積電", "100,000", "50,000", "+50,000", "20,000", "10,000", "+10,000", "5,000", "3,000", "+2,000", "+62,000"},
			},
			"date": "20260620",
		})
	}))
	defer srv.Close()

	newTestClient(t, srv)
	raw, err := FetchT86(context.Background(), "20260620", url.Values{})
	if err != nil {
		t.Fatalf("FetchT86 returned error: %v", err)
	}
	resp, ok := raw.(T86Response)
	if !ok {
		t.Fatalf("expected T86Response, got %T", raw)
	}
	if resp.GetStat() != "OK" {
		t.Fatalf("expected stat OK, got %q", resp.GetStat())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(resp.Data))
	}
	row, err := ParseT86Row(resp.Data[0])
	if err != nil {
		t.Fatalf("ParseT86Row returned error: %v", err)
	}
	if row.Code != "2330" {
		t.Errorf("Code = %q, want %q", row.Code, "2330")
	}
	if row.Name != "台積電" {
		t.Errorf("Name = %q, want %q", row.Name, "台積電")
	}
	if row.ForeignBuy != 100000 {
		t.Errorf("ForeignBuy = %d, want %d", row.ForeignBuy, 100000)
	}
	if row.ForeignSell != 50000 {
		t.Errorf("ForeignSell = %d, want %d", row.ForeignSell, 50000)
	}
	if row.ForeignNet != 50000 {
		t.Errorf("ForeignNet = %d, want %d", row.ForeignNet, 50000)
	}
	if row.TrustBuy != 20000 {
		t.Errorf("TrustBuy = %d, want %d", row.TrustBuy, 20000)
	}
	if row.TrustSell != 10000 {
		t.Errorf("TrustSell = %d, want %d", row.TrustSell, 10000)
	}
	if row.TrustNet != 10000 {
		t.Errorf("TrustNet = %d, want %d", row.TrustNet, 10000)
	}
	if row.DealerBuy != 5000 {
		t.Errorf("DealerBuy = %d, want %d", row.DealerBuy, 5000)
	}
	if row.DealerSell != 3000 {
		t.Errorf("DealerSell = %d, want %d", row.DealerSell, 3000)
	}
	if row.DealerNet != 2000 {
		t.Errorf("DealerNet = %d, want %d", row.DealerNet, 2000)
	}
	if row.TotalNet != 62000 {
		t.Errorf("TotalNet = %d, want %d", row.TotalNet, 62000)
	}
}

func TestFetchT86_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	newTestClient(t, srv)
	_, err := FetchT86(context.Background(), "20260620", url.Values{})
	if err == nil {
		t.Fatal("expected error for no-data response, got nil")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected ErrNoData, got: %v", err)
	}
}
