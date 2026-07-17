// circuit_group.go — fixed breaker group for Yahoo HTML page fetches.
package scrape

import (
	"context"

	"github.com/bizshuk/yfin/utils/httpx"
)

const circuitGroupYahooWeb = "yahoo-web"

func yahooWebCircuitContext(ctx context.Context) context.Context {
	return httpx.WithCircuitGroup(ctx, circuitGroupYahooWeb)
}
