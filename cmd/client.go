// client.go — shared builder used by every sub-package:
// `CreateClient` (facade.NewClientWithConfig with cmd.Global flag overrides).
//
// facade is the single handler bridging cmd/ → svc/; CreateClient returns
// *facade.Client.
//
// Capacity: 1 builder function.
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
// ampy-config HTTP settings so `yfin --qps=10` wins over `cfg.qps=2`.
//
// Historical note: prior to the *httpx.Config consolidation this builder
// did a per-field copy from `config.HTTPConfig` into `*httpx.Config`
// (`httpConfigToHttpx`). That mapper is gone — `(*Config).GetHTTPConfig`
// now returns the assembled `*httpx.Config` directly.
func CreateClient() (*facade.Client, error) {
	// Determine effective config path
	effectivePath := Global.ConfigFile
	if effectivePath == "" {
		effectivePath = "config/effective.yaml"
	}

	// Load configuration using ampy-config
	loader := config.NewLoader(effectivePath)
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
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

	return facade.NewClientWithConfig(httpxConfig), nil
}
