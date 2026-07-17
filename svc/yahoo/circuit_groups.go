// circuit_groups.go — fixed bounded breaker-family names for Yahoo request
// surfaces. URL and response handling remain in endpoint files.
package yahoo

import (
	"context"

	"github.com/bizshuk/yfin/utils/httpx"
)

const (
	circuitGroupAuth       = "yahoo-auth"
	circuitGroupChart      = "yahoo-chart"
	circuitGroupTimeseries = "yahoo-timeseries"
	circuitGroupOptions    = "yahoo-options"
	circuitGroupNews       = "yahoo-news"
	circuitGroupWeb        = "yahoo-web"
)

func circuitContext(ctx context.Context, group string) context.Context {
	return httpx.WithCircuitGroup(ctx, group)
}
