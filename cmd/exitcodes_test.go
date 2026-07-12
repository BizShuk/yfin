// exitcodes_test.go — sanity check that the 5 `Exit*` constants keep the
// documented values. The codes are part of the yfin CLI's external contract
// (orchestrators / shell scripts depend on them); a silent regression would
// be hard to catch otherwise.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExitCodes(t *testing.T) {
	assert.Equal(t, 0, ExitSuccess)
	assert.Equal(t, 1, ExitGeneral)
	assert.Equal(t, 2, ExitPaidFeature)
	assert.Equal(t, 3, ExitConfigError)
	assert.Equal(t, 4, ExitPublishError)
}
