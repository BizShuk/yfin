package httpx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircuitGroupContextNormalizesValue(t *testing.T) {
	ctx := WithCircuitGroup(context.Background(), "  Yahoo-Auth  ")
	assert.Equal(t, "yahoo-auth", circuitGroupFromContext(ctx))
}

func TestCircuitGroupContextIgnoresBlankValue(t *testing.T) {
	base := context.Background()
	ctx := WithCircuitGroup(base, "   ")
	assert.Equal(t, "", circuitGroupFromContext(ctx))
	assert.Equal(t, base, ctx)
}
