// twse_test.go — twse CLI tests using httptest-backed fake TWSE: `MI_INDEX`
// happy path, empty-data fallback, unknown endpoint, missing `--stock`,
// `FMSRFK` stock-no dispatch, registry ↔ fetcher-map coverage, and
// `facade.NewTwseClient` composition verification. Like the production
// file, this test reaches TWSE only through `facade` — no svc/twse import.
// Capacity: 6 test functions + 3 test helpers
// (`setTwseClientForTest`/`captureStdout`/`resetTwseCfg`).
package twse

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/utils/httpx"
)

// twseBaseURL is the production TWSE origin facade.NewTwseClient must target.
const twseBaseURL = "https://www.twse.com.tw/rwd/zh"

// newTWSEClientForTest builds a *facade.TwseClient pointed at srv.URL via a
// real (latency-free) httpx.Client. This exercises the full transport path
// that production uses — just pointed at a local httptest server.
func newTWSEClientForTest(t *testing.T, srv *httptest.Server) *facade.TwseClient {
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
	return facade.NewTwseClientWithHTTP(hc, srv.URL)
}

func resetTwseCfg(t *testing.T) {
	t.Helper()
	twseCfg = twseConfig{}
}

// setTwseClientForTest swaps the twseClientProvider for the lifetime of one test.
func setTwseClientForTest(t *testing.T, client *facade.TwseClient) func() {
	t.Helper()
	prev := twseClientProvider
	twseClientProvider = func() *facade.TwseClient { return client }
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

	c := twseCmd
	var stdout, stderr string
	stdout, stderr = captureStdout(t, func() {
		if err := runTwseEndpoint(c, nil); err != nil {
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

	c := twseCmd
	var stdout, stderr string
	stdout, stderr = captureStdout(t, func() {
		if err := runTwseEndpoint(c, nil); err != nil {
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

	c := twseCmd
	_, stderr := captureStdout(t, func() {
		if err := runTwseEndpoint(c, nil); err == nil {
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

	c := twseCmd
	_, stderr := captureStdout(t, func() {
		if err := runTwseEndpoint(c, nil); err == nil {
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

	c := twseCmd
	stdout, _ := captureStdout(t, func() {
		if err := runTwseEndpoint(c, nil); err != nil {
			t.Errorf("runTwseEndpoint returned error: %v", err)
		}
	})
	if !strings.Contains(stdout, "2330") {
		t.Errorf("expected stockNo in output, got: %s", stdout)
	}
}

// TestNameToFetcher_CoversAllRegistryEntries asserts every endpoint the
// registry advertises actually has a fetcher wired up, so a new registry
// entry cannot ship without its dispatch arm.
func TestNameToFetcher_CoversAllRegistryEntries(t *testing.T) {
	for name := range facade.TwseRegistry {
		if !facade.TwseHasFetcher(name) {
			t.Errorf("registry entry %q has no fetcher wired in facade/twse_dispatch.go", name)
		}
	}
}

// TestNewTwseClient verifies facade.NewTwseClient yields a non-nil handle
// pointed at the production TWSE origin with a non-nil injected Caller.
func TestNewTwseClient(t *testing.T) {
	client := facade.NewTwseClient()
	if client == nil {
		t.Fatal("facade.NewTwseClient returned nil client")
	}
	if got, want := client.BaseURL(), twseBaseURL; got != want {
		t.Errorf("client.BaseURL() = %q, want %q", got, want)
	}
	if client.Caller() == nil {
		t.Error("client.Caller() returned nil; facade.NewTwseClient must inject a non-nil httpx.Caller")
	}
}
