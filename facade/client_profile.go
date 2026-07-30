// client_profile.go — Yahoo host policies registered on the shared httpx
// connection pool during facade client composition.
package facade

import (
	"net/http"
	"strings"

	"github.com/bizshuk/yfin/svc/scrape"
	"github.com/bizshuk/yfin/utils/httpx"
)

type hostProfileRegistrar interface {
	UseProfile(httpx.HostProfile)
}

func registerYahooHostProfiles(registrar hostProfileRegistrar, config *httpx.Config) {
	for _, profile := range yahooHostProfiles(config) {
		registrar.UseProfile(profile)
	}
}

func yahooHostProfiles(config *httpx.Config) []httpx.HostProfile {
	return yahooHostProfilesWithScrape(config, nil)
}

func yahooHostProfilesWithScrape(config, scrapeHTTP *httpx.Config) []httpx.HostProfile {
	if config == nil {
		config = httpx.DefaultConfig()
	}
	if scrapeHTTP == nil {
		scrapeHTTP = scrape.DefaultConfig().HTTP
	}

	queryProfile := func(host string) httpx.HostProfile {
		return httpx.HostProfile{
			Host:       host,
			EndpointFn: yahooEndpoint,
			QPS:        config.QPS,
			Burst:      config.Burst,
			MaxBody:    config.MaxBodyBytes,
			UserAgent:  config.UserAgent,
		}
	}

	return []httpx.HostProfile{
		queryProfile("query1.finance.yahoo.com"),
		queryProfile("query2.finance.yahoo.com"),
		yahooWebHostProfile(scrapeHTTP),
		queryProfile("fc.yahoo.com"),
	}
}

func yahooWebHostProfile(config *httpx.Config) httpx.HostProfile {
	if config == nil {
		config = scrape.DefaultConfig().HTTP
	}
	return httpx.HostProfile{
		Host:       "finance.yahoo.com",
		EndpointFn: yahooEndpoint,
		QPS:        config.QPS,
		Burst:      config.Burst,
		MaxBody:    config.MaxBodyBytes,
		UserAgent:  config.UserAgent,
		Headers: http.Header{
			"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
			"Accept-Language": {"en-US,en;q=0.9"},
		},
	}
}

func yahooEndpoint(path string) string {
	switch {
	case strings.Contains(path, "/v8/finance/chart/"):
		return "bars_1d"
	case strings.Contains(path, "/v1/finance/quote"):
		return "quote"
	case strings.Contains(path, "/v11/finance/quoteSummary/"),
		strings.Contains(path, "/v1/finance/quoteSummary/"):
		return "fundamentals"
	default:
		return httpx.CommonEndpointFn(path)
	}
}
