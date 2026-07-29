package facade

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	cacheutil "github.com/bizshuk/yfin/utils/cache"
)

type foundationFetcher interface {
	FetchFoundation(context.Context, string, string, string) (any, error)
}

// FoundationArtifact is one foundation fetch and its yfin-owned cache
// location.
type FoundationArtifact struct {
	Data         any
	Cached       bool
	RelativePath string
}

// CachedFoundationClient adds the established tiered raw-cache policy to all
// foundation commands.
type CachedFoundationClient struct {
	upstream foundationFetcher
	store    *cacheutil.JSONStore
	now      func() time.Time
}

// NewCachedFoundationClient builds a cache-backed foundation client.
func NewCachedFoundationClient(rawRoot string) (*CachedFoundationClient, error) {
	return newCachedFoundationClient(rawRoot, NewClient(), time.Now)
}

func newCachedFoundationClient(
	rawRoot string,
	upstream foundationFetcher,
	now func() time.Time,
) (*CachedFoundationClient, error) {
	if upstream == nil {
		return nil, errors.New("facade: foundation upstream is required")
	}
	if now == nil {
		return nil, errors.New("facade: foundation clock is required")
	}
	store, err := cacheutil.NewJSONStore(rawRoot)
	if err != nil {
		return nil, err
	}
	return &CachedFoundationClient{
		upstream: upstream,
		store:    store,
		now:      now,
	}, nil
}

// Fetch returns a fresh-enough cached artifact or fetches and persists it.
func (c *CachedFoundationClient) Fetch(
	ctx context.Context,
	command string,
	symbol string,
	runID string,
	force bool,
) (FoundationArtifact, error) {
	command = strings.TrimSpace(command)
	if !isFoundationCommand(command) {
		return FoundationArtifact{}, fmt.Errorf(
			"facade: unsupported foundation command %q",
			command,
		)
	}
	commandComponent, err := safeCacheComponent(command)
	if err != nil {
		return FoundationArtifact{}, err
	}
	providerSymbol := strings.TrimSpace(symbol)
	symbolComponent, err := safeCacheComponent(providerSymbol)
	if err != nil {
		return FoundationArtifact{}, err
	}
	now := c.now().UTC()

	if cacheutil.ShouldSkip(
		commandComponent,
		symbolComponent,
		force,
		c.store.Root(),
		now,
	) {
		return c.loadLatest(commandComponent, symbolComponent)
	}

	data, err := c.upstream.FetchFoundation(
		ctx,
		commandComponent,
		providerSymbol,
		runID,
	)
	if err != nil {
		return FoundationArtifact{}, err
	}
	relativePath := filepath.Join(
		commandComponent,
		fmt.Sprintf(
			"%s.%s.json",
			symbolComponent,
			now.Format("2006-01-02"),
		),
	)
	if err := c.store.Write(relativePath, data); err != nil {
		return FoundationArtifact{}, err
	}
	return FoundationArtifact{
		Data:         data,
		RelativePath: relativePath,
	}, nil
}

func (c *CachedFoundationClient) loadLatest(
	command string,
	symbol string,
) (FoundationArtifact, error) {
	paths, err := c.store.List(command)
	if err != nil {
		return FoundationArtifact{}, err
	}
	prefix := symbol + "."
	for index := len(paths) - 1; index >= 0; index-- {
		path := paths[index]
		if !strings.HasPrefix(filepath.Base(path), prefix) {
			continue
		}
		var data any
		if err := c.store.Read(path, &data); err != nil {
			return FoundationArtifact{}, err
		}
		return FoundationArtifact{
			Data:         data,
			Cached:       true,
			RelativePath: path,
		}, nil
	}
	return FoundationArtifact{}, fmt.Errorf(
		"facade: fresh foundation cache missing for %s/%s",
		command,
		symbol,
	)
}
