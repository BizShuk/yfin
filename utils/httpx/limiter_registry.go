package httpx

import (
	"strings"
	"sync"
)

// rateLimiterRegistry owns one token bucket per normalized upstream host.
type rateLimiterRegistry struct {
	mu       sync.Mutex
	limiters map[string]*RateLimiter
}

func newRateLimiterRegistry() *rateLimiterRegistry {
	return &rateLimiterRegistry{limiters: make(map[string]*RateLimiter)}
}

func (r *rateLimiterRegistry) forHost(host string, qps float64, burst int) *RateLimiter {
	key := normalizeProfileHost(host)
	if key == "" {
		key = strings.ToLower(strings.TrimSpace(host))
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if limiter, ok := r.limiters[key]; ok {
		return limiter
	}
	limiter := NewRateLimiter(qps, burst)
	r.limiters[key] = limiter
	return limiter
}
