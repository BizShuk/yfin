package httpx

import (
	"context"
	"io"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRateLimiterFractionalQPS(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	limiter := NewRateLimiter(0.7, 1)
	limiter.now = func() time.Time { return now }
	limiter.lastTime = now

	require.NoError(t, limiter.Wait(context.Background()))
	now = now.Add(2 * time.Second)
	require.NoError(t, limiter.Wait(context.Background()))
	assert.InDelta(t, 0.7, limiter.rate, 0.000001)
}

func TestNewRateLimiterRejectsNonPositiveConfig(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, "httpx: rate limit qps must be greater than zero", func() {
		NewRateLimiter(0, 1)
	})
	assert.PanicsWithValue(t, "httpx: rate limit burst must be greater than zero", func() {
		NewRateLimiter(1, 0)
	})
}

func TestRateLimiterQueuesConcurrentWaiters(t *testing.T) {
	limiter := NewRateLimiter(20, 1)
	require.NoError(t, limiter.Wait(context.Background()))

	started := time.Now()
	type waitResult struct {
		duration time.Duration
		err      error
	}
	completed := make(chan waitResult, 2)
	for range 2 {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			err := limiter.Wait(ctx)
			completed <- waitResult{duration: time.Since(started), err: err}
		}()
	}

	first, second := <-completed, <-completed
	require.NoError(t, first.err)
	require.NoError(t, second.err)
	times := []time.Duration{first.duration, second.duration}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	assert.GreaterOrEqual(t, times[1]-times[0], 30*time.Millisecond)
}

func TestRateLimiterRegistryIsolatesHosts(t *testing.T) {
	t.Parallel()

	registry := newRateLimiterRegistry()
	hostA := registry.forHost("QUERY1.FINANCE.YAHOO.COM", 1, 1)
	hostAAgain := registry.forHost("query1.finance.yahoo.com:443", 1, 1)
	hostB := registry.forHost("finance.yahoo.com", 0.7, 1)

	assert.Same(t, hostA, hostAAgain)
	assert.NotSame(t, hostA, hostB)
	assert.InDelta(t, 0.7, hostB.rate, 0.000001)

	require.NoError(t, hostA.Wait(context.Background()))
	require.NoError(t, hostB.Wait(context.Background()))
}

func TestClientRateLimiterIsScopedByHost(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	config.QPS = 1
	config.Burst = 1
	config.MaxAttempts = 1
	client := NewClient(config)
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("ok")),
			Request:    req,
		}, nil
	})

	requestA, err := http.NewRequest(http.MethodGet, "https://query1.finance.yahoo.com/a", nil)
	require.NoError(t, err)
	response, err := client.Do(context.Background(), requestA)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	requestB, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://finance.yahoo.com/b", nil)
	require.NoError(t, err)
	response, err = client.Do(ctx, requestB)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
}
