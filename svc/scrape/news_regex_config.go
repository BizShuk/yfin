// news_regex_config.go — the YAML-driven regex pattern set the news
// extractors run on, plus its process-wide lazy loader. Split out of
// extract_news.go so pattern configuration is separable from parsing.
package scrape

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// NewsRegexConfig holds the regex patterns for news extraction
type NewsRegexConfig struct {
	ArticleContainer string `yaml:"article_container"`
	Title            string `yaml:"title"`
	ArticleLink      string `yaml:"article_link"`
	PublishingInfo   string `yaml:"publishing_info"`
	ImageURL         string `yaml:"image_url"`
	RelatedTickers   string `yaml:"related_tickers"`
	NextPageHint     string `yaml:"next_page_hint"`

	RelativeTime struct {
		Minutes   string `yaml:"minutes"`
		Hours     string `yaml:"hours"`
		Days      string `yaml:"days"`
		Weeks     string `yaml:"weeks"`
		Yesterday string `yaml:"yesterday"`
	} `yaml:"relative_time"`

	URLCleanup struct {
		UTMParams      string `yaml:"utm_params"`
		TrackingParams string `yaml:"tracking_params"`
		Fragment       string `yaml:"fragment"`
		QuerySeparator string `yaml:"query_separator"`
	} `yaml:"url_cleanup"`
}

var newsRegexConfig *NewsRegexConfig

// LoadNewsRegexConfig loads the news regex patterns from YAML file
func LoadNewsRegexConfig() error {
	if newsRegexConfig != nil {
		return nil // Already loaded
	}

	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to get current file path")
	}

	configPath := filepath.Join(filepath.Dir(filename), "regex", "news.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read news regex config file: %w", err)
	}

	newsRegexConfig = &NewsRegexConfig{}
	if err := yaml.Unmarshal(data, newsRegexConfig); err != nil {
		return fmt.Errorf("failed to parse news regex config YAML: %w", err)
	}

	return nil
}

// newsMetrics is the news-specific subset of the old scrape.Metrics —
// outcome counter + parse latency. Kept here (next to the parser it
// instruments) because they have no use outside ParseNews.
//
// IMPORTANT (test-isolation constraint): newNewsMetrics registers
// `yfin_scrape_news_total` and `yfin_scrape_news_parse_latency_ms`
// against `prometheus.DefaultRegisterer` via `promauto`, guarded by a
// package-level `sync.Once`. Tests that reset `DefaultRegisterer` (e.g.
// `prometheus.DefaultRegisterer = prometheus.NewRegistry()`) MUST
// invoke `ParseNews` (or otherwise call `newNewsMetrics`) BEFORE the
// reset, or the once-fired registration will continue pointing at the
// old registry and the new one will not have these metrics. Changing
// this contract to accept a `prometheus.Registerer` is deliberately
