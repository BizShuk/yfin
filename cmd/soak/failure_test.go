package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type soakRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn soakRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestFailureTransportDelegatesWhenFailureRateIsZero(t *testing.T) {
	var delegated int
	next := soakRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		delegated++
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Status:     "204 No Content",
			Header:     make(http.Header),
			Body:       http.NoBody,
			Request:    req,
		}, nil
	})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test", nil)
	require.NoError(t, err)

	resp, err := newFailureTransport(next, 0).RoundTrip(req)

	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	require.Equal(t, 1, delegated)
}

func TestFailureTransportInjectsRetryableResponseAtFullRate(t *testing.T) {
	next := soakRoundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("underlying transport must not run when failure rate is 1")
		return nil, nil
	})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test", nil)
	require.NoError(t, err)

	resp, err := newFailureTransport(next, 1).RoundTrip(req)

	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	require.Equal(t, "true", resp.Header.Get("X-YFin-Failure-Injected"))
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Contains(t, strings.ToLower(string(body)), "injected")
}
