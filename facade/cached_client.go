package facade

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/bizshuk/yfin/model"
	cacheutil "github.com/bizshuk/yfin/utils/cache"
)

type dailyBarsFetcher interface {
	FetchDailyBars(context.Context, string, time.Time, time.Time, bool, string) (*BarBatch, error)
}

const dailyBarsCacheVersion = "v4"

// CachedClient adds persistent read-through caching to Yahoo daily bars.
type CachedClient struct {
	upstream dailyBarsFetcher
	store    *cacheutil.JSONStore
	mu       sync.Mutex
}

// NewCachedClient builds a cache-backed Yahoo client.
func NewCachedClient(rawRoot string) (*CachedClient, error) {
	return newCachedClient(rawRoot, NewClient())
}

func newCachedClient(rawRoot string, upstream dailyBarsFetcher) (*CachedClient, error) {
	if upstream == nil {
		return nil, errors.New("facade: daily-bars upstream is required")
	}
	store, err := cacheutil.NewJSONStore(rawRoot)
	if err != nil {
		return nil, err
	}
	return &CachedClient{upstream: upstream, store: store}, nil
}

// FetchDailyBars returns an exact cached request or fetches and persists it.
func (c *CachedClient) FetchDailyBars(
	ctx context.Context,
	symbol string,
	start time.Time,
	end time.Time,
	adjusted bool,
	runID string,
) (*BarBatch, error) {
	path, err := dailyBarsCachePath(symbol, start, end, adjusted)
	if err != nil {
		return nil, err
	}
	if batch, err := c.readDailyBars(path); err == nil {
		return batch, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if batch, err := c.readDailyBars(path); err == nil {
		return batch, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	batch, err := c.upstream.FetchDailyBars(ctx, symbol, start, end, adjusted, runID)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, nil
	}
	if err := c.store.Write(path, batch); err != nil {
		return nil, err
	}
	return batch, nil
}

// LoadDailyBars merges all cached daily-bar requests for symbol.
func (c *CachedClient) LoadDailyBars(symbol string) (*BarBatch, error) {
	component, err := safeCacheComponent(symbol)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join("yahoo", "bars", dailyBarsCacheVersion, component)
	paths, err := c.store.List(dir)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("facade: no cached bars for %s: %w", symbol, fs.ErrNotExist)
	}

	byDate := make(map[string]model.Bar)
	result := &BarBatch{Symbol: symbol}
	for _, path := range paths {
		var batch BarBatch
		if err := c.store.Read(path, &batch); err != nil {
			return nil, err
		}
		if batch.Symbol != "" {
			result.Symbol = batch.Symbol
		}
		if batch.MIC != "" {
			result.MIC = batch.MIC
		}
		if batch.FirstTradeDate != "" {
			result.FirstTradeDate = batch.FirstTradeDate
		}
		for _, bar := range batch.Bars {
			byDate[bar.Date] = bar
		}
	}
	result.Bars = make([]model.Bar, 0, len(byDate))
	for _, bar := range byDate {
		result.Bars = append(result.Bars, bar)
	}
	sort.Slice(result.Bars, func(i, j int) bool {
		return result.Bars[i].Date < result.Bars[j].Date
	})
	return result, nil
}

func (c *CachedClient) readDailyBars(path string) (*BarBatch, error) {
	var batch BarBatch
	if err := c.store.Read(path, &batch); err != nil {
		return nil, err
	}
	return &batch, nil
}

func dailyBarsCachePath(symbol string, start, end time.Time, adjusted bool) (string, error) {
	component, err := safeCacheComponent(symbol)
	if err != nil {
		return "", err
	}
	request := fmt.Sprintf(
		"%s_%s_%t.json",
		start.UTC().Format("20060102T150405Z"),
		end.UTC().Format("20060102T150405Z"),
		adjusted,
	)
	return filepath.Join(
		"yahoo",
		"bars",
		dailyBarsCacheVersion,
		component,
		request,
	), nil
}

func safeCacheComponent(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("facade: cache component is required")
	}
	var builder strings.Builder
	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsDigit(character) ||
			character == '-' || character == '_' || character == '.' {
			builder.WriteRune(character)
			continue
		}
		fmt.Fprintf(&builder, "%%%X", character)
	}
	component := builder.String()
	if component == "." || component == ".." {
		return "", fmt.Errorf("facade: invalid cache component %q", value)
	}
	return component, nil
}
