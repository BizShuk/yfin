package facade

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeFoundationFetcher struct {
	calls   int
	symbols []string
}

func (f *fakeFoundationFetcher) FetchFoundation(
	_ context.Context,
	_ string,
	symbol string,
	_ string,
) (any, error) {
	f.calls++
	f.symbols = append(f.symbols, symbol)
	return map[string]string{"symbol": "AAPL"}, nil
}

func TestCachedFoundationClientMissFetchesAndThenHitsCache(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	upstream := &fakeFoundationFetcher{}
	client, err := newCachedFoundationClient(root, upstream, func() time.Time {
		return now
	})
	if err != nil {
		t.Fatal(err)
	}

	first, err := client.Fetch(
		context.Background(),
		"info",
		"AAPL",
		"first-run",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if first.Cached {
		t.Fatal("first Fetch() reported a cache hit")
	}
	if first.RelativePath != filepath.Join("info", "AAPL.2026-07-29.json") {
		t.Fatalf("first RelativePath = %q", first.RelativePath)
	}

	now = time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	second, err := client.Fetch(
		context.Background(),
		"info",
		"AAPL",
		"second-run",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !second.Cached {
		t.Fatal("second Fetch() did not report the monthly cache hit")
	}
	if upstream.calls != 1 {
		t.Fatalf("upstream calls = %d, want 1", upstream.calls)
	}
	if _, err := os.Stat(filepath.Join(root, first.RelativePath)); err != nil {
		t.Fatalf("cached artifact: %v", err)
	}
}

func TestCachedFoundationClientForceRefreshes(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	upstream := &fakeFoundationFetcher{}
	client, err := newCachedFoundationClient(root, upstream, func() time.Time {
		return now
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Fetch(context.Background(), "info", "AAPL", "first", false); err != nil {
		t.Fatal(err)
	}
	now = now.Add(24 * time.Hour)
	result, err := client.Fetch(context.Background(), "info", "AAPL", "forced", true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Cached {
		t.Fatal("forced Fetch() reported a cache hit")
	}
	if upstream.calls != 2 {
		t.Fatalf("upstream calls = %d, want 2", upstream.calls)
	}
	if result.RelativePath != filepath.Join("info", "AAPL.2026-07-30.json") {
		t.Fatalf("forced RelativePath = %q", result.RelativePath)
	}
}

func TestCachedFoundationClientPreservesProviderSymbol(t *testing.T) {
	root := t.TempDir()
	upstream := &fakeFoundationFetcher{}
	client, err := newCachedFoundationClient(
		root,
		upstream,
		func() time.Time {
			return time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Fetch(
		context.Background(),
		"info",
		"^GSPC",
		"index-run",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(upstream.symbols) != 1 || upstream.symbols[0] != "^GSPC" {
		t.Fatalf("upstream symbols = %v, want original symbol", upstream.symbols)
	}
	wantPath := filepath.Join("info", "%5EGSPC.2026-07-29.json")
	if result.RelativePath != wantPath {
		t.Fatalf("RelativePath = %q, want %q", result.RelativePath, wantPath)
	}
}
