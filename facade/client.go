// client.go — the facade.Client handle: the struct (MIC-inference cache +
// `sync.RWMutex`), its 2 constructors, and the helpers shared across every
// surface. The methods live beside it by surface: chart-API `Fetch*` in
// client_yahoo.go, ScaledDecimal `Fetch*Norm` in client_norm.go, and HTML
// `Scrape*` in client_scrape.go.
package facade

import (
	"context"
	"strings"
	"sync"

	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/svc/scrape"
	"github.com/bizshuk/yfin/svc/yahoo"
	"github.com/bizshuk/yfin/utils/httpx"
)

// Client provides a high-level interface for fetching Yahoo Finance data
type Client struct {
	yahooClient  *yahoo.Client
	scrapeClient scrape.Client
	micCache     map[string]string // Cache for MIC inference to avoid repeated API calls
	micCacheMu   sync.RWMutex      // Mutex for MIC cache
}

// NewClient creates a new Yahoo Finance client with default configuration
func NewClient() *Client {
	config := httpx.DefaultConfig()
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, "")
	scrapeClient, _ := scrape.NewClient(scrape.DefaultConfig(), httpClient)

	return &Client{
		yahooClient:  yahooClient,
		scrapeClient: scrapeClient,
		micCache:     make(map[string]string),
	}
}

// NewClientWithConfig creates a new Yahoo Finance client with custom configuration
func NewClientWithConfig(config *httpx.Config) *Client {
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, config.BaseURL)
	scrapeClient, _ := scrape.NewClient(scrape.DefaultConfig(), httpClient)

	return &Client{
		yahooClient:  yahooClient,
		scrapeClient: scrapeClient,
		micCache:     make(map[string]string),
	}
}

// inferMICForSymbol attempts to infer the MIC code for a symbol by fetching company info
// Uses caching to avoid repeated API calls for the same symbol
func (c *Client) inferMICForSymbol(ctx context.Context, symbol string) string {
	// Check cache first
	c.micCacheMu.RLock()
	if mic, found := c.micCache[symbol]; found {
		c.micCacheMu.RUnlock()
		return mic
	}
	c.micCacheMu.RUnlock()

	// Cache miss - fetch company info
	companyInfo, err := c.FetchCompanyInfo(ctx, symbol, "mic-inference")
	if err != nil {
		// If we can't fetch company info, cache empty string to avoid repeated failures
		c.micCacheMu.Lock()
		c.micCache[symbol] = ""
		c.micCacheMu.Unlock()
		return ""
	}

	mic := model.InferMIC(companyInfo.Exchange, companyInfo.FullExchangeName)

	// Cache the result
	c.micCacheMu.Lock()
	c.micCache[symbol] = mic
	c.micCacheMu.Unlock()

	return mic
}

// isAuthenticationError checks if an error indicates authentication is required
func isAuthenticationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "authentication")
}
