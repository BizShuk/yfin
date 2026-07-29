package facade

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	cacheutil "github.com/bizshuk/yfin/utils/cache"
)

// RawCache exposes legacy JSON cache reads without leaking filesystem
// ownership to SDK consumers.
type RawCache struct {
	store *cacheutil.JSONStore
}

// NewRawCache opens the yfin-managed raw cache at rawRoot.
func NewRawCache(rawRoot string) (*RawCache, error) {
	store, err := cacheutil.NewJSONStore(rawRoot)
	if err != nil {
		return nil, err
	}
	return &RawCache{store: store}, nil
}

// ReadLatestJSON decodes the lexically latest JSON file whose base name has
// filenamePrefix. It returns the suffix between the prefix and ".json".
func (c *RawCache) ReadLatestJSON(relativeDir, filenamePrefix string, destination any) (string, error) {
	if filenamePrefix == "" {
		return "", fmt.Errorf("facade: raw-cache filename prefix is required")
	}
	paths, err := c.store.List(relativeDir)
	if err != nil {
		return "", err
	}
	matches := make([]string, 0, len(paths))
	for _, path := range paths {
		name := filepath.Base(path)
		if strings.HasPrefix(name, filenamePrefix) {
			matches = append(matches, path)
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("facade: no raw cache entry for %s: %w", filenamePrefix, fs.ErrNotExist)
	}
	sort.Strings(matches)
	latest := matches[len(matches)-1]
	if err := c.store.Read(latest, destination); err != nil {
		return "", err
	}
	name := filepath.Base(latest)
	asOf := strings.TrimSuffix(strings.TrimPrefix(name, filenamePrefix), ".json")
	return asOf, nil
}
