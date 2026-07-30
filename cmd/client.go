// client.go — shared builder used by every sub-package:
// `CreateClient` / `CreateClientFromConfig` (facade.NewClientWithConfig with
// cmd.Global flag overrides).
//
// facade is the single handler bridging cmd/ → svc/; both builders return
// *facade.Client.
//
// Capacity: 2 builder functions.
package cmd

import (
	"fmt"

	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/utils/httpx"
)

// CreateClient builds the yfin CLI's fetch handle. Returns *facade.Client —
// the same SDK surface external consumers use. CLI flag overrides
// (`--qps`, `--retry-max`, `--timeout`) are applied on top of the
// YAML HTTP settings so `yfin --qps=10` wins over `cfg.qps=2`.
//
// Historical note: prior to the *httpx.Config consolidation this builder
// did a per-field copy from `config.HTTPConfig` into `*httpx.Config`
// (`httpConfigToHttpx`). That mapper is gone — `(*Config).GetHTTPConfig`
// now returns the assembled `*httpx.Config` directly.
func CreateClient() (*facade.Client, error) {
	loader := config.NewLoader(Global.ConfigFile)
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return CreateClientFromConfig(cfg)
}

// CreateClientFromConfig builds the CLI client from one already-loaded config
// snapshot. Commands that also consume non-HTTP settings use this path so all
// runtime policy comes from the same load.
func CreateClientFromConfig(cfg *config.Config) (*facade.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}
	// Get the post-load assembled HTTP config (already an *httpx.Config
	// after the consolidation). Apply CLI flag overrides on top so
	// `yfin --qps=10` wins over the yaml default.
	httpxConfig := cfg.GetHTTPConfig()
	if httpxConfig == nil {
		// Load() guarantees this is non-nil; this fallback exists only
		// for defensive symmetry with httpx.NewClient's nil-config path.
		httpxConfig = httpx.DefaultConfig()
	}
	if Global.QPS > 0 {
		httpxConfig.QPS = Global.QPS
	}
	if Global.RetryMax > 0 {
		httpxConfig.MaxAttempts = Global.RetryMax
	}
	if Global.Timeout > 0 {
		httpxConfig.Timeout = Global.Timeout
	}

	scrapeConfig := cfg.GetScrapeConfig()
	if scrapeConfig != nil && scrapeConfig.HTTP != nil {
		if Global.QPS > 0 {
			scrapeConfig.HTTP.QPS = Global.QPS
		}
		if Global.RetryMax > 0 {
			scrapeConfig.HTTP.MaxAttempts = Global.RetryMax
		}
		if Global.Timeout > 0 {
			scrapeConfig.HTTP.Timeout = Global.Timeout
		}
	}

	client := facade.NewClientWithConfig(httpxConfig)
	if err := client.ConfigureScrape(scrapeConfig); err != nil {
		return nil, err
	}
	return client, nil
}
