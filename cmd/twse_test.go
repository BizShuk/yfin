// twse_test.go — twse CLI tests using httptest-backed fake TWSE: `MI_INDEX` happy path, empty-data fallback, unknown endpoint, missing `--stock`, `FMSRFK` stock-no dispatch, and a registry-coverage check over `twseNameToFetcher`. Capacity: 6 test functions + 2 test helpers (`withTestServer`/`captureStdout`/`resetTwseCfg`).
package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bizshuk/yfin/svc/twse"
)

// withTestServer redirects the twse.BaseURL to a local httptest server for
// the lifetime of the test, so runTwseEndpoint hits the fake TWSE.
func withTestServer(t *testing.T, handler http.HandlerFunc) (string, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	oldBase := twse.BaseURL
	twse.BaseURL = srv.URL
	return srv.URL, func() {
		srv.Close()
		twse.BaseURL = oldBase
	}
}

func resetTwseCfg(t *testing.T) {
	t.Helper()
	twseCfg = twseConfig{}
}

// captureStdout captures os.Stdout + os.Stderr for fn's duration.
func captureStdout(t *testing.T, fn func()) (string, string) {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr
	outDone, errDone := make(chan struct{}), make(chan struct{})
	var outBuf, errBuf bytes.Buffer
	go func() { _, _ = io.Copy(&outBuf, rOut); close(outDone) }()
	go func() { _, _ = io.Copy(&errBuf, rErr); close(errDone) }()
	fn()
	_ = wOut.Close()
	_ = wErr.Close()
	<-outDone
	<-errDone
	os.Stdout = oldOut
	os.Stderr = oldErr
	return outBuf.String(), errBuf.String()
}

func TestRunTwseEndpoint_MI_INDEX(t *testing.T) {
	defer resetTwseCfg(t)
	_, restore := withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/afterTrading/MI_INDEX") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":   "OK",
			"title":  "每日收盤行情",
			"fields": []string{"指數", "收盤指數", "漲跌點數", "漲跌百分比"},
			"data":   [][]string{{"發行量加權股價指數", "17,500.12", "+120.34", "+0.69%"}},
			"date":   "20221230",
		})
	})
	defer restore()

	twseCfg.endpoint = "MI_INDEX"
	twseCfg.date = "20221230"

	cmd := twseCmd
	var stdout, stderr string
	stdout, stderr = captureStdout(t, func() {
		if err := runTwseEndpoint(cmd, nil); err != nil {
			t.Errorf("runTwseEndpoint returned error: %v", err)
		}
	})
	_ = stderr
	if !strings.Contains(stdout, "發行量加權股價指數") {
		t.Errorf("expected index name in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "17,500.12") {
		t.Errorf("expected close price in output, got: %s", stdout)
	}
}

func TestRunTwseEndpoint_NoData(t *testing.T) {
	defer resetTwseCfg(t)
	_, restore := withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	})
	defer restore()

	twseCfg.endpoint = "MI_INDEX"
	twseCfg.date = "19000101"

	cmd := twseCmd
	var stdout, stderr string
	stdout, stderr = captureStdout(t, func() {
		if err := runTwseEndpoint(cmd, nil); err != nil {
			t.Errorf("expected nil error for no-data, got: %v", err)
		}
	})
	if !strings.Contains(stdout, "no data") && !strings.Contains(stderr, "no data") {
		t.Errorf("expected 'no data' info message, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestRunTwseEndpoint_UnknownEndpoint(t *testing.T) {
	defer resetTwseCfg(t)
	twseCfg.endpoint = "BOGUS_ENDPOINT"
	twseCfg.date = "20221230"

	cmd := twseCmd
	_, stderr := captureStdout(t, func() {
		if err := runTwseEndpoint(cmd, nil); err == nil {
			t.Error("expected error for unknown endpoint, got nil")
		}
	})
	if !strings.Contains(stderr, "unknown endpoint") {
		t.Errorf("expected 'unknown endpoint' in stderr, got: %s", stderr)
	}
}

func TestRunTwseEndpoint_STOCK_DAY_RequiresStock(t *testing.T) {
	defer resetTwseCfg(t)
	twseCfg.endpoint = "STOCK_DAY"
	twseCfg.date = "20221230"
	// --stock omitted on purpose; Registry says NeedsStock=true.

	cmd := twseCmd
	_, stderr := captureStdout(t, func() {
		if err := runTwseEndpoint(cmd, nil); err == nil {
			t.Error("expected error when --stock missing, got nil")
		}
	})
	if !strings.Contains(stderr, "--stock") {
		t.Errorf("expected --stock hint in stderr, got: %s", stderr)
	}
}

func TestRunTwseEndpoint_FMSRFK_DispatchesWithStock(t *testing.T) {
	defer resetTwseCfg(t)
	_, restore := withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/exchangeReport/FMSRFK") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("stockNo") != "2330" {
			t.Errorf("expected stockNo=2330, got %q", r.URL.Query().Get("stockNo"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":    "OK",
			"title":   "個股月成交資訊",
			"fields":  []string{"年度", "月份", "最高", "最低", "加權平均價", "成交股數", "成交金額", "週轉率%"},
			"data":    [][]string{{"2022", "01", "688", "600", "643", "1234567890", "789012345678", "12.34"}},
			"stockNo": "2330",
			"date":    "2022",
		})
	})
	defer restore()

	twseCfg.endpoint = "FMSRFK"
	twseCfg.date = "2022"
	twseCfg.stockNo = "2330"

	cmd := twseCmd
	stdout, _ := captureStdout(t, func() {
		if err := runTwseEndpoint(cmd, nil); err != nil {
			t.Errorf("runTwseEndpoint returned error: %v", err)
		}
	})
	if !strings.Contains(stdout, "2330") {
		t.Errorf("expected stockNo in output, got: %s", stdout)
	}
}

func TestNameToFetcher_CoversAllRegistryEntries(t *testing.T) {
	for name, ep := range twse.Registry {
		if _, ok := twseNameToFetcher[name]; !ok {
			t.Errorf("registry entry %q (%s) has no fetcher wired in cmd/twse.go", name, ep.Path)
		}
	}
}
