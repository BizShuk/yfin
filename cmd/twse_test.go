// twse_test.go — twse CLI tests using httptest-backed fake TWSE: `MI_INDEX`
// happy path, empty-data fallback, unknown endpoint, missing `--stock`,
// `FMSRFK` stock-no dispatch, and a registry-coverage check over
// `twseNameToFetcher`. Capacity: 6 test functions + 3 test helpers
// (`setTwseClientForTest`/`captureStdout`/`resetTwseCfg`).
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bizshuk/yfin/svc/twse"
	"github.com/bizshuk/yfin/utils/httpx"
)

// newTWSEClientForTest builds a *twse.Client pointed at srv.URL via a
// real (latency-free) httpx.Client. Per Task 4, every TWSE fetch is
// routed through the injected caller, so this exercises the full
// transport path that production uses — just pointed at a local
// httptest server instead of www.twse.com.tw.
func newTWSEClientForTest(t *testing.T, srv *httptest.Server) *twse.Client {
	t.Helper()
	hc := httpx.NewClient(&httpx.Config{
		BaseURL:          "",
		Timeout:          5 * time.Second,
		MaxAttempts:      1,
		BackoffBaseMs:    1,
		BackoffJitterMs:  0,
		MaxDelayMs:       10,
		QPS:              1000,
		Burst:            1000,
		CircuitWindow:    time.Second,
		FailureThreshold: 1000,
		ResetTimeout:     time.Second,
	})
	return twse.NewClientWithURL(hc, srv.URL)
}

func resetTwseCfg(t *testing.T) {
	t.Helper()
	twseCfg = twseConfig{}
}

// setTwseClientForTest swaps the twseClientProvider for the lifetime of
// one test. Returns a restore function suitable for t.Cleanup.
func setTwseClientForTest(t *testing.T, client *twse.Client) func() {
	t.Helper()
	prev := twseClientProvider
	twseClientProvider = func() *twse.Client { return client }
	return func() { twseClientProvider = prev }
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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer srv.Close()

	twseCfg.endpoint = "MI_INDEX"
	twseCfg.date = "20221230"

	restore := setTwseClientForTest(t, newTWSEClientForTest(t, srv))
	defer restore()

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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	twseCfg.endpoint = "MI_INDEX"
	twseCfg.date = "19000101"

	restore := setTwseClientForTest(t, newTWSEClientForTest(t, srv))
	defer restore()

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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer srv.Close()

	twseCfg.endpoint = "FMSRFK"
	twseCfg.date = "2022"
	twseCfg.stockNo = "2330"

	restore := setTwseClientForTest(t, newTWSEClientForTest(t, srv))
	defer restore()

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

func TestDispatcher_Call_DispatchesViaClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":  "OK",
			"title": "twse:MI_INDEX",
			"data":  [][]string{{"1", "2"}},
			"date":  "20260620",
		})
	}))
	defer srv.Close()

	client := newTWSEClientForTest(t, srv)
	d := twse.NewDispatcher(client)
	raw, err := d.Call(context.Background(), "MI_INDEX", "20260620", url.Values{})
	if err != nil {
		t.Fatalf("Dispatcher.Call: %v", err)
	}
	if raw == nil {
		t.Fatal("expected non-nil response")
	}
}
