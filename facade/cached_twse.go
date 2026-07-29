package facade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"sort"
	"sync"

	cacheutil "github.com/bizshuk/yfin/utils/cache"
)

type twseDispatcher interface {
	Dispatch(context.Context, string, string, url.Values) (any, error)
}

// CachedTwseClient adds persistent read-through caching to TWSE dispatches.
type CachedTwseClient struct {
	upstream twseDispatcher
	store    *cacheutil.JSONStore
	mu       sync.Mutex
}

// NewCachedTwseClient builds a cache-backed TWSE client.
func NewCachedTwseClient(rawRoot string) (*CachedTwseClient, error) {
	return newCachedTwseClient(rawRoot, NewTwseClient())
}

func newCachedTwseClient(rawRoot string, upstream twseDispatcher) (*CachedTwseClient, error) {
	if upstream == nil {
		return nil, errors.New("facade: TWSE upstream is required")
	}
	store, err := cacheutil.NewJSONStore(rawRoot)
	if err != nil {
		return nil, err
	}
	return &CachedTwseClient{upstream: upstream, store: store}, nil
}

// Dispatch returns an exact cached request or fetches and persists it.
func (c *CachedTwseClient) Dispatch(
	ctx context.Context,
	endpoint string,
	date string,
	options url.Values,
) (any, error) {
	path, err := twseCachePath(endpoint, date, options)
	if err != nil {
		return nil, err
	}
	if value, err := c.read(path); err == nil {
		return value, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if value, err := c.read(path); err == nil {
		return value, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	value, err := c.upstream.Dispatch(ctx, endpoint, date, options)
	if err != nil {
		return nil, err
	}
	if err := c.store.Write(path, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (c *CachedTwseClient) read(path string) (any, error) {
	var value any
	if err := c.store.Read(path, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func twseCachePath(endpoint, date string, options url.Values) (string, error) {
	endpointComponent, err := safeCacheComponent(endpoint)
	if err != nil {
		return "", err
	}
	dateComponent, err := safeCacheComponent(date)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(canonicalValues(options)))
	return filepath.Join(
		"twse",
		endpointComponent,
		fmt.Sprintf("%s_%s.json", dateComponent, hex.EncodeToString(digest[:])),
	), nil
}

func canonicalValues(values url.Values) string {
	canonical := make(url.Values, len(values))
	for key, items := range values {
		canonical[key] = append([]string(nil), items...)
		sort.Strings(canonical[key])
	}
	return canonical.Encode()
}
