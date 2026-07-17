// client_test.go — Tests `Client` retry, authority/group breaker outcomes, half-open transitions, rate limiting, and error classification.
package httpx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	config := DefaultConfig()
	config.BaseURL = server.URL
	config.MaxAttempts = 5
	config.BackoffBaseMs = 10 // Fast backoff for testing

	client := NewClient(config)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Do(ctx, req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestClientRetryReplaysPOSTBody(t *testing.T) {
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodies = append(bodies, string(body))
		if len(bodies) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxAttempts = 2
	config.BackoffBaseMs = 1
	config.BackoffJitterMs = 0
	client := NewClient(config)
	payload := []byte(`{"serviceConfig":{"snippetCount":10,"s":["AAPL"]}}`)
	req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader(payload))
	require.NoError(t, err)

	resp, err := client.Do(context.Background(), req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, []string{string(payload), string(payload)}, bodies)
}

func TestClientCircuitBreaker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.BaseURL = server.URL
	config.MaxAttempts = 1
	config.FailureThreshold = 2
	config.CircuitWindow = 100 * time.Millisecond
	config.ResetTimeout = 50 * time.Millisecond

	client := NewClient(config)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	breaker := client.circuitBreakers.forHost(req.URL.Host)

	// First request should fail
	_, err = client.Do(ctx, req)
	if err == nil {
		t.Error("Request 1 should have failed")
	}

	// Check circuit is still closed (1 failure, threshold is 2)
	if breaker.State() != StateClosed {
		t.Errorf("Expected circuit to be closed after 1 failure, got state %v", breaker.State())
	}

	// Second request should fail and open the circuit
	_, err = client.Do(ctx, req)
	if err == nil {
		t.Error("Request 2 should have failed")
	}

	// Check circuit is now open (2 failures, threshold is 2)
	if breaker.State() != StateOpen {
		t.Errorf("Expected circuit to be open after 2 failures, got state %v", breaker.State())
	}

	// Third request should be rejected by circuit breaker
	_, err = client.Do(ctx, req)
	if err == nil {
		t.Error("Request should have been rejected by circuit breaker")
	}
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}

	// Wait for circuit to reset (transition to half-open)
	time.Sleep(60 * time.Millisecond) // Wait longer than resetTimeout

	// The next request owns the single half-open probe and still fails.
	_, err = client.Do(ctx, req)
	if err == nil {
		t.Error("Request should still fail (server still returns 500)")
	}

	// After failure in half-open state, circuit should be open again
	if breaker.State() != StateOpen {
		t.Errorf("Expected circuit to be open again after failure in half-open state, got state %v", breaker.State())
	}
}

func TestClientRetryFailureRecordsOneBreakerOutcome(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxAttempts = 5
	config.BackoffBaseMs = 1
	config.BackoffJitterMs = 0
	config.MaxDelayMs = 2
	config.FailureThreshold = 2
	client := NewClient(config)
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	_, err = client.Do(context.Background(), req)
	require.Error(t, err)
	breaker := client.circuitBreakers.forHost(req.URL.Host)
	assert.Equal(t, 1, breaker.Failures())
	assert.Equal(t, StateClosed, breaker.State())
}

func TestClientCircuitBreakerIsScopedByHost(t *testing.T) {
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failing.Close()
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthy.Close()

	config := DefaultConfig()
	config.MaxAttempts = 1
	config.FailureThreshold = 1
	client := NewClient(config)
	failReq, err := http.NewRequest(http.MethodGet, failing.URL, nil)
	require.NoError(t, err)
	healthyReq, err := http.NewRequest(http.MethodGet, healthy.URL, nil)
	require.NoError(t, err)

	_, err = client.Do(context.Background(), failReq)
	require.Error(t, err)
	resp, err := client.Do(context.Background(), healthyReq)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func TestClientCircuitBreakerIsScopedByRequestGroup(t *testing.T) {
	var authHits, chartHits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			authHits++
			w.WriteHeader(http.StatusTooManyRequests)
		case "/chart":
			chartHits++
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxAttempts = 1
	config.FailureThreshold = 1
	config.QPS = 100
	config.Burst = 10
	client := NewClient(config)

	authCtx := WithCircuitGroup(context.Background(), "yahoo-auth")
	authReq, err := http.NewRequestWithContext(authCtx, http.MethodGet, server.URL+"/auth", nil)
	require.NoError(t, err)
	_, err = client.Do(context.Background(), authReq)
	require.Error(t, err)

	secondAuthReq, err := http.NewRequestWithContext(authCtx, http.MethodGet, server.URL+"/auth", nil)
	require.NoError(t, err)
	_, err = client.Do(context.Background(), secondAuthReq)
	require.ErrorIs(t, err, ErrCircuitOpen)

	chartCtx := WithCircuitGroup(context.Background(), "yahoo-chart")
	chartReq, err := http.NewRequestWithContext(chartCtx, http.MethodGet, server.URL+"/chart", nil)
	require.NoError(t, err)
	resp, err := client.Do(context.Background(), chartReq)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, 1, authHits)
	assert.Equal(t, 1, chartHits)
}

func TestClientHTTP404IsBreakerSuccess(t *testing.T) {
	statuses := []int{http.StatusInternalServerError, http.StatusNotFound}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		status := statuses[0]
		statuses = statuses[1:]
		w.WriteHeader(status)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxAttempts = 1
	config.FailureThreshold = 0
	config.FailureRateThreshold = 0.75
	config.MinimumRequests = 2
	client := NewClient(config)
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	_, _ = client.Do(context.Background(), req)
	_, _ = client.Do(context.Background(), req)

	breaker := client.circuitBreakers.forHost(req.URL.Host)
	assert.Equal(t, 2, breaker.Samples())
	assert.Equal(t, 1, breaker.Failures())
	assert.Equal(t, StateClosed, breaker.State())
}

func TestClientCancellationIsBreakerNeutral(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxAttempts = 1
	client := NewClient(config)
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err = client.Do(ctx, req)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, 0, client.circuitBreakers.forHost(req.URL.Host).Samples())
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(2.0, 2) // 2 QPS, burst of 2

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()

	// First two requests should be immediate (burst)
	for i := 0; i < 2; i++ {
		err := limiter.Wait(ctx)
		if err != nil {
			t.Errorf("Request %d failed: %v", i+1, err)
		}
	}

	// Third request should be throttled
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Request 3 failed: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 500*time.Millisecond {
		t.Errorf("Expected at least 500ms delay, got %v", elapsed)
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
		fatal     bool
	}{
		{
			name:      "429 error",
			err:       NewHTTPError(429, "Too Many Requests", nil),
			retryable: true,
			fatal:     false,
		},
		{
			name:      "500 error",
			err:       NewHTTPError(500, "Internal Server Error", nil),
			retryable: true,
			fatal:     false,
		},
		{
			name:      "400 error",
			err:       NewHTTPError(400, "Bad Request", nil),
			retryable: false,
			fatal:     true,
		},
		{
			name:      "404 error",
			err:       NewHTTPError(404, "Not Found", nil),
			retryable: false,
			fatal:     true,
		},
		{
			name:      "timeout error",
			err:       ErrTimeout,
			retryable: true,
			fatal:     false,
		},
		{
			name:      "decode error",
			err:       ErrDecode,
			retryable: false,
			fatal:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsRetryableError(tt.err) != tt.retryable {
				t.Errorf("IsRetryableError() = %v, want %v", IsRetryableError(tt.err), tt.retryable)
			}
			if IsFatalError(tt.err) != tt.fatal {
				t.Errorf("IsFatalError() = %v, want %v", IsFatalError(tt.err), tt.fatal)
			}
		})
	}
}

func TestClientDoReturnsTypedHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxAttempts = 1
	config.Burst = 100
	config.QPS = 100
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	_, err = NewClient(config).Do(context.Background(), req)
	var statusErr *HTTPError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", statusErr.StatusCode, http.StatusUnauthorized)
	}
}
