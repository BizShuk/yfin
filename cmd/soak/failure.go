// failure.go — process-local HTTP transport fault injection for soak runs.
package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

type failureTransport struct {
	next        http.RoundTripper
	failureRate float64
	random      func() float64
}

func newFailureTransport(next http.RoundTripper, failureRate float64) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	if failureRate <= 0 {
		return next
	}
	return &failureTransport{
		next:        next,
		failureRate: failureRate,
		random:      rand.Float64,
	}
}

func (transport *failureTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if transport.random() >= transport.failureRate {
		return transport.next.RoundTrip(request)
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Retry-After", "1")
	headers.Set("X-YFin-Failure-Injected", "true")
	return &http.Response{
		StatusCode:    http.StatusServiceUnavailable,
		Status:        "503 Service Unavailable",
		Header:        headers,
		Body:          io.NopCloser(strings.NewReader(`{"error":"injected soak transport failure"}`)),
		ContentLength: -1,
		Request:       request,
	}, nil
}
