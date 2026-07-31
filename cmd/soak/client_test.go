package main

import (
	"testing"
	"time"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

func TestSoakHTTPConfigAppliesOverridesWithoutMutatingLoadedConfig(t *testing.T) {
	loaded := httpx.DefaultConfig()
	originalAttempts := loaded.MaxAttempts
	originalTimeout := loaded.Timeout
	originalQPS := loaded.QPS

	got, err := soakHTTPConfig(
		loaded,
		GlobalConfig{RetryMax: 7, Timeout: 4 * time.Second},
		3.5,
		0.25,
	)

	require.NoError(t, err)
	require.NotSame(t, loaded, got)
	require.Equal(t, 7, got.MaxAttempts)
	require.Equal(t, 4*time.Second, got.Timeout)
	require.Equal(t, 3.5, got.QPS)
	require.NotNil(t, got.WrapTransport)
	require.Equal(t, originalAttempts, loaded.MaxAttempts)
	require.Equal(t, originalTimeout, loaded.Timeout)
	require.Equal(t, originalQPS, loaded.QPS)
}

func TestSoakHTTPConfigRejectsMissingLoadedConfig(t *testing.T) {
	_, err := soakHTTPConfig(nil, GlobalConfig{}, 1, 0)

	require.ErrorContains(t, err, "HTTP config")
}
