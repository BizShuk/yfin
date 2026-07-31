// validation.go — validates the standalone soak runner's execution policy.
package main

import (
	"fmt"
)

var apiSoakEndpoints = map[string]struct{}{
	"quote":        {},
	"daily-bars":   {},
	"fundamentals": {},
}

var scrapeSoakEndpoints = map[string]struct{}{
	"key-statistics":   {},
	"financials":       {},
	"analysis":         {},
	"profile":          {},
	"news":             {},
	"balance-sheet":    {},
	"cash-flow":        {},
	"analyst-insights": {},
}

func validateSoakConfig(config *SoakConfig) error {
	if config == nil {
		return fmt.Errorf("soak config is required")
	}
	if config.Duration <= 0 {
		return fmt.Errorf("duration must be greater than zero")
	}
	if config.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than zero")
	}
	if config.QPS <= 0 {
		return fmt.Errorf("QPS must be greater than zero")
	}
	if config.ProbeInterval <= 0 {
		return fmt.Errorf("probe interval must be greater than zero")
	}
	if config.FailureRate < 0 || config.FailureRate > 1 {
		return fmt.Errorf("failure rate must be between 0 and 1")
	}
	if config.Fallback != "auto" && config.Fallback != "api-only" && config.Fallback != "scrape-only" {
		return fmt.Errorf("unsupported fallback strategy %q", config.Fallback)
	}

	endpoints := parseEndpoints(config.Endpoints)
	if len(endpoints) == 0 {
		return fmt.Errorf("at least one endpoint is required")
	}
	seen := make(map[string]struct{}, len(endpoints))
	for _, endpoint := range endpoints {
		if _, duplicate := seen[endpoint]; duplicate {
			return fmt.Errorf("duplicate endpoint %q", endpoint)
		}
		seen[endpoint] = struct{}{}
		if !supportsEndpoint(config.Fallback, endpoint) {
			return fmt.Errorf("endpoint %q is not supported by %s strategy", endpoint, config.Fallback)
		}
	}
	return nil
}

func supportsEndpoint(strategy, endpoint string) bool {
	_, api := apiSoakEndpoints[endpoint]
	_, scrape := scrapeSoakEndpoints[endpoint]
	switch strategy {
	case "api-only":
		return api
	case "scrape-only":
		return scrape
	case "auto":
		return api || scrape
	default:
		return false
	}
}
