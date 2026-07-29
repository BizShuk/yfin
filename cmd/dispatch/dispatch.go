// dispatch.go adapts the public foundation facade to the yfin batch CLI.
package dispatch

import (
	"context"

	"github.com/bizshuk/yfin/facade"
)

// FetchContext bundles everything a command needs to fetch its data.
type FetchContext struct {
	Root  *facade.Client
	RunID string
}

// fetchFunc fetches a single command's data; result must be JSON-marshalable.
type fetchFunc func(ctx context.Context, fc *FetchContext, symbol string) (any, error)

var commandOrder = facade.FoundationCommands()

var commandRegistry = newCommandRegistry(commandOrder)

func newCommandRegistry(commands []string) map[string]fetchFunc {
	registry := make(map[string]fetchFunc, len(commands))
	for _, name := range commands {
		command := name
		registry[command] = func(
			ctx context.Context,
			fetchContext *FetchContext,
			symbol string,
		) (any, error) {
			return fetchContext.Root.FetchFoundation(
				ctx,
				command,
				symbol,
				fetchContext.RunID,
			)
		}
	}
	return registry
}
