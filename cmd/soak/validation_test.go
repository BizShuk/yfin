package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateSoakConfigRejectsIncompatibleEndpointStrategy(t *testing.T) {
	config := validSoakConfig()
	config.Fallback = "api-only"
	config.Endpoints = "profile"

	err := validateSoakConfig(config)

	require.ErrorContains(t, err, "not supported")
}

func TestValidateSoakConfigAcceptsAutoSurfaceUnion(t *testing.T) {
	config := validSoakConfig()
	config.Fallback = "auto"
	config.Endpoints = "quote,profile,fundamentals"

	require.NoError(t, validateSoakConfig(config))
}

func validSoakConfig() *SoakConfig {
	return &SoakConfig{
		UniverseFile:  "universe.txt",
		Endpoints:     "quote",
		Fallback:      "auto",
		Duration:      time.Minute,
		Concurrency:   1,
		QPS:           1,
		ProbeInterval: time.Minute,
	}
}
