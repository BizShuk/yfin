package scrape_test

import (
	"testing"

	"github.com/bizshuk/yfin/svc/scrape"
)

func TestLoadFinancialsRegexConfig(t *testing.T) {
	err := scrape.LoadFinancialsRegexConfig()
	if err != nil {
		t.Fatalf("LoadFinancialsRegexConfig failed: %v", err)
	}

	// Note: financialsRegexConfig is internal and not accessible from external tests
	// The fact that LoadFinancialsRegexConfig() doesn't return an error indicates success
}
