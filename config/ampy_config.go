// Package config is a thin alias kept so existing callers can keep
// writing `config.Config`. New code should import
// github.com/bizshuk/yfin/config/types directly — every sub-config
// struct, the Loader, the default-writer, and the adapter methods
// (GetHTTPConfig / GetBusConfig / GetFXConfig / GetScrapeConfig /
// ValidateInterval / ValidateAdjustmentPolicy) live there.
//
// Capacity: 1 type alias (`Config`).
package config

import "github.com/bizshuk/yfin/config/types"

// Config is the canonical root config; see
// github.com/bizshuk/yfin/config/types for the full type tree.
type Config = types.Config