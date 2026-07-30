package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateClientFromConfigRejectsNil(t *testing.T) {
	_, err := CreateClientFromConfig(nil)
	require.ErrorContains(t, err, "configuration is required")
}
