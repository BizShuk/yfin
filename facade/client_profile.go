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
	if config == nil {
		config = httpx.DefaultConfig()
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

	scrapeHTTP := scrape.DefaultConfig().HTTP
	return []httpx.HostProfile{
		queryProfile("query1.finance.yahoo.com"),
		queryProfile("query2.finance.yahoo.com"),
		{
			Host:       "finance.yahoo.com",
			EndpointFn: yahooEndpoint,
			QPS:        scrapeHTTP.QPS,
			Burst:      scrapeHTTP.Burst,
			MaxBody:    scrapeHTTP.MaxBodyBytes,
			UserAgent:  scrapeHTTP.UserAgent,
			Headers: http.Header{
				"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
				"Accept-Language": {"en-US,en;q=0.9"},
			},
		},
		queryProfile("fc.yahoo.com"),
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
