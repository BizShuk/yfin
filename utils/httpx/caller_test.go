// caller_test.go — Tests `Caller.Get` returns *Meta with Status / Bytes / Duration / Attempts / Host / Endpoint / Gzip populated. Capacity: 5 test functions.
package httpx

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// stubCaller is a Caller implementation for tests that records invocation
// count and returns a canned body + Meta.
type stubCaller struct {
	calls   int
	lastCtx context.Context
	body    []byte
	meta    *Meta
	err     error
}

func (s *stubCaller) Get(_ context.Context, _ string, _ url.Values) ([]byte, *Meta, error) {
	s.calls++
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.body, s.meta, nil
}

// TestCaller_GetInterface — Sanity check: *Client satisfies Caller.
func TestCaller_GetInterface(t *testing.T) {
	c := NewClient(nil)
	var _ Caller = c // compile-time assertion
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// TestGet_PopulatesMetaOnSuccess — Get returns Meta with Status, Bytes, Duration, Host, Endpoint populated.
func TestGet_PopulatesMetaOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("hello"))
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 1
	cfg.BackoffBaseMs = 1
	cfg.BackoffJitterMs = 0

	c := NewClient(cfg)
	body, meta, err := c.Get(testCtx(t), "/some/path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(body, []byte("hello")) {
		t.Errorf("expected body %q, got %q", "hello", body)
	}
	if meta == nil {
		t.Fatal("expected meta, got nil")
	}
	if meta.Status != http.StatusOK {
		t.Errorf("expected Status 200, got %d", meta.Status)
	}
	if meta.Bytes != len("hello") {
		t.Errorf("expected Bytes %d, got %d", len("hello"), meta.Bytes)
	}
	if meta.Duration <= 0 {
		t.Errorf("expected Duration > 0, got %v", meta.Duration)
	}
	if meta.Attempts != 1 {
		t.Errorf("expected Attempts 1, got %d", meta.Attempts)
	}
	if meta.Host == "" {
		t.Errorf("expected Host populated, got empty")
	}
	if meta.Endpoint == "" {
		t.Errorf("expected Endpoint populated, got empty")
	}
	if meta.Gzip {
		t.Errorf("expected Gzip false for plain response, got true")
	}
}

// TestGet_AbsoluteURLOverridesBaseURL verifies that a caller can target a
// different host without the configured BaseURL being prepended.
func TestGet_AbsoluteURLOverridesBaseURL(t *testing.T) {
	var baseCalls int
	baseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseCalls++
		_, _ = w.Write([]byte("configured base"))
	}))
	defer baseServer.Close()

	var targetCalls int
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetCalls++
		if r.URL.Path != "/quote/AAPL/financials" {
			t.Errorf("expected target path %q, got %q", "/quote/AAPL/financials", r.URL.Path)
		}
		if r.URL.Query().Get("period") != "annual" {
			t.Errorf("expected period query %q, got %q", "annual", r.URL.Query().Get("period"))
		}
		_, _ = w.Write([]byte("target"))
	}))
	defer targetServer.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = baseServer.URL
	cfg.MaxAttempts = 1

	c := NewClient(cfg)
	body, meta, err := c.Get(
		testCtx(t),
		targetServer.URL+"/quote/AAPL/financials",
		url.Values{"period": {"annual"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if baseCalls != 0 {
		t.Errorf("expected configured base to receive 0 requests, got %d", baseCalls)
	}
	if targetCalls != 1 {
		t.Errorf("expected absolute target to receive 1 request, got %d", targetCalls)
	}
	if string(body) != "target" {
		t.Errorf("expected target body %q, got %q", "target", body)
	}
	wantHost := targetServer.Listener.Addr().String()
	if meta.Host != wantHost {
		t.Errorf("expected Meta.Host %q, got %q", wantHost, meta.Host)
	}
}

// TestGet_GzipFlagSet — Get sets Meta.Gzip=true for gzipped responses.
func TestGet_GzipFlagSet(t *testing.T) {
	want := []byte("gzipped payload")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write(want)
		_ = gz.Close()
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 1
	cfg.BackoffBaseMs = 1

	c := NewClient(cfg)
	body, meta, err := c.Get(testCtx(t), "/", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(body, want) {
		t.Errorf("expected body %q, got %q", want, body)
	}
	if meta == nil {
		t.Fatal("expected meta, got nil")
	}
	if !meta.Gzip {
		t.Errorf("expected Meta.Gzip true, got false")
	}
	if meta.Bytes != len(want) {
		t.Errorf("expected Bytes %d (decompressed), got %d", len(want), meta.Bytes)
	}
}

// TestGet_AttemptsOnRetry — After retry-then-success, Meta.Attempts > 1.
func TestGet_AttemptsOnRetry(t *testing.T) {
	var attempts int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 5
	cfg.BackoffBaseMs = 1
	cfg.BackoffJitterMs = 0
	cfg.MaxDelayMs = 5

	c := NewClient(cfg)
	body, meta, err := c.Get(testCtx(t), "/", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != "ok" {
		t.Errorf("expected body %q, got %q", "ok", body)
	}
	if meta == nil {
		t.Fatal("expected meta, got nil")
	}
	if meta.Attempts != 3 {
		t.Errorf("expected Attempts 3 (2 fails + 1 success), got %d", meta.Attempts)
	}
	if meta.Duration <= 0 {
		t.Errorf("expected Duration > 0, got %v", meta.Duration)
	}
}
