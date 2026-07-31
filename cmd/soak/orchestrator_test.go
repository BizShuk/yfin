// orchestrator_test.go — orchestrator construction, input parsing, and
// deterministic short-run coverage.
package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bizshuk/yfin/facade"
)

func TestNewOrchestrator(t *testing.T) {
	// Create a temporary universe file
	tempDir := t.TempDir()
	universeFile := filepath.Join(tempDir, "test_universe.txt")

	universeContent := `# Test universe
AAPL
MSFT
GOOGL`

	if err := os.WriteFile(universeFile, []byte(universeContent), 0o644); err != nil {
		t.Fatalf("Failed to create test universe file: %v", err)
	}

	soakConfig := &SoakConfig{
		UniverseFile:  universeFile,
		Endpoints:     "key-statistics,news",
		Fallback:      "scrape-only",
		Duration:      10 * time.Second,
		Concurrency:   2,
		QPS:           1.0,
		ProbeInterval: 1 * time.Hour,
		FailureRate:   0.0, // Disable failure injection for test
		MemoryCheck:   true,
	}

	orchestrator, err := NewOrchestrator(facade.NewClient(), discardLogger(), soakConfig)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Verify orchestrator was created correctly
	if orchestrator == nil {
		t.Fatal("Orchestrator is nil")
	}

	if len(orchestrator.tickers) != 3 {
		t.Errorf("Expected 3 tickers, got %d", len(orchestrator.tickers))
	}

	expectedTickers := []string{"AAPL", "MSFT", "GOOGL"}
	for i, expected := range expectedTickers {
		if i >= len(orchestrator.tickers) || orchestrator.tickers[i] != expected {
			t.Errorf("Expected ticker %s at index %d, got %s", expected, i, orchestrator.tickers[i])
		}
	}

	if len(orchestrator.endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(orchestrator.endpoints))
	}
}

func TestLoadTickerUniverse(t *testing.T) {
	tempDir := t.TempDir()
	universeFile := filepath.Join(tempDir, "test_universe.txt")

	universeContent := `# Test universe with comments
AAPL
# This is a comment
MSFT

GOOGL
# Another comment
TSLA`

	if err := os.WriteFile(universeFile, []byte(universeContent), 0o644); err != nil {
		t.Fatalf("Failed to create test universe file: %v", err)
	}

	tickers, err := loadTickerUniverse(universeFile)
	if err != nil {
		t.Fatalf("Failed to load ticker universe: %v", err)
	}

	expected := []string{"AAPL", "MSFT", "GOOGL", "TSLA"}
	if len(tickers) != len(expected) {
		t.Errorf("Expected %d tickers, got %d", len(expected), len(tickers))
	}

	for i, expectedTicker := range expected {
		if i >= len(tickers) || tickers[i] != expectedTicker {
			t.Errorf("Expected ticker %s at index %d, got %s", expectedTicker, i, tickers[i])
		}
	}
}

func TestParseEndpoints(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "key-statistics,financials,news",
			expected: []string{"key-statistics", "financials", "news"},
		},
		{
			input:    "key-statistics, financials , news ",
			expected: []string{"key-statistics", "financials", "news"},
		},
		{
			input:    "single",
			expected: []string{"single"},
		},
	}

	for _, test := range tests {
		result := parseEndpoints(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("For input %q, expected %d endpoints, got %d", test.input, len(test.expected), len(result))
			continue
		}

		for i, expected := range test.expected {
			if result[i] != expected {
				t.Errorf("For input %q, expected endpoint %s at index %d, got %s", test.input, expected, i, result[i])
			}
		}
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
