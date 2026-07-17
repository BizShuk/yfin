// circuit_group.go — request-context contract for bounded logical circuit
// groups. Values never leave the process as HTTP headers.
package httpx

import (
	"context"
	"strings"
)

type circuitGroupContextKey struct{}

// WithCircuitGroup returns a context carrying a normalized logical breaker
// group. Blank groups preserve the original ungrouped context.
func WithCircuitGroup(ctx context.Context, group string) context.Context {
	group = normalizeCircuitGroup(group)
	if group == "" {
		return ctx
	}
	return context.WithValue(ctx, circuitGroupContextKey{}, group)
}

func circuitGroupFromContext(ctx context.Context) string {
	group, _ := ctx.Value(circuitGroupContextKey{}).(string)
	return normalizeCircuitGroup(group)
}

func normalizeCircuitGroup(group string) string {
	return strings.ToLower(strings.TrimSpace(group))
}
