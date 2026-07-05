// body_test.go — Tests gzip auto-decode + body size cap (MaxBodyBytes) on Caller.Call.
package httpx

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestCall_RejectsOversizedBody verifies that Call returns ErrBodyTooLarge
// when the response body exceeds Config.MaxBodyBytes.
func TestCall_RejectsOversizedBody(t *testing.T) {
	const max = int64(1 << 20) // 1 MiB
	const payload = 10 << 20   // 10 MiB

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.Repeat([]byte("x"), payload)
		_, _ = w.Write(buf)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 1
	cfg.BackoffBaseMs = 1
	cfg.MaxBodyBytes = max

	client := NewClient(cfg)

	_, err := client.Call(testCtx(t), "/", nil)
	if err == nil {
		t.Fatalf("expected ErrBodyTooLarge, got nil")
	}
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

// TestCall_AllowsBodyAtLimit verifies that a body exactly at MaxBodyBytes succeeds.
func TestCall_AllowsBodyAtLimit(t *testing.T) {
	const max = int64(1 << 20) // 1 MiB

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.Repeat([]byte("y"), int(max))
		_, _ = w.Write(buf)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 1
	cfg.BackoffBaseMs = 1
	cfg.MaxBodyBytes = max

	client := NewClient(cfg)

	body, err := client.Call(testCtx(t), "/", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int64(len(body)) != max {
		t.Fatalf("expected body length %d, got %d", max, len(body))
	}
}

// TestCall_DecodesGzipByDefault verifies that gzip-encoded responses are
// transparently decoded by Caller.Call.
func TestCall_DecodesGzipByDefault(t *testing.T) {
	want := []byte(`{"hello":"world"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/json")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write(want)
		_ = gz.Close()
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 1
	cfg.BackoffBaseMs = 1

	client := NewClient(cfg)

	body, err := client.Call(testCtx(t), "/", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(body, want) {
		t.Fatalf("expected %q, got %q", want, body)
	}
}

// TestCall_GzipBodyExceedingLimit verifies the size cap applies to the
// decompressed body, not the wire bytes.
func TestCall_GzipBodyExceedingLimit(t *testing.T) {
	const max = int64(1 << 10) // 1 KiB after decompress
	const payload = 16 << 10  // 16 KiB uncompressed

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write(bytes.Repeat([]byte("z"), payload))
		_ = gz.Close()
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 1
	cfg.BackoffBaseMs = 1
	cfg.MaxBodyBytes = max

	client := NewClient(cfg)

	_, err := client.Call(testCtx(t), "/", nil)
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge for gzipped oversized body, got %v", err)
	}
}

// TestCall_DefaultNoBodyCap verifies that MaxBodyBytes=0 (default) does not
// impose any limit, matching pre-existing behaviour.
func TestCall_DefaultNoBodyCap(t *testing.T) {
	const payload = 4 << 20 // 4 MiB

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("a"), payload))
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MaxAttempts = 1
	cfg.BackoffBaseMs = 1
	// cfg.MaxBodyBytes left at 0 = unlimited

	client := NewClient(cfg)

	body, err := client.Call(testCtx(t), "/", url.Values{"q": {"1"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) != payload {
		t.Fatalf("expected %d bytes, got %d", payload, len(body))
	}
}

// TestReadBody_DirectGzipEncoding covers readBody without going through Caller.
func TestReadBody_DirectGzipEncoding(t *testing.T) {
	want := []byte("plain-text-payload")
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write(want)
	_ = gz.Close()

	resp := &http.Response{
		Body:   io.NopCloser(&buf),
		Header: http.Header{"Content-Encoding": []string{"gzip"}},
	}
	body, err := readBody(resp, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(body, want) {
		t.Fatalf("expected %q, got %q", want, body)
	}
}

// TestReadBody_DirectTooLarge covers readBody's size-cap path on plain bodies.
func TestReadBody_DirectTooLarge(t *testing.T) {
	resp := &http.Response{
		Body:   io.NopCloser(strings.NewReader(strings.Repeat("x", 2048))),
		Header: http.Header{},
	}
	_, err := readBody(resp, 1024)
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("expected ErrBodyTooLarge, got %v", err)
	}
}

// testCtx returns a short-deadline background context for tests.
func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}