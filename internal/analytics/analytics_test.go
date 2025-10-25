package analytics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDisabled(t *testing.T) {
	t.Run("returns true if XBOM_DISABLE_TELEMETRY is set to true", func(t *testing.T) {
		err := os.Setenv(telemetryDisableEnvKey, "true")
		require.NoError(t, err)
		defer func() { _ = os.Unsetenv(telemetryDisableEnvKey) }()

		assert.True(t, IsDisabled())
	})

	t.Run("returns false if XBOM_DISABLE_TELEMETRY is not set", func(t *testing.T) {
		assert.False(t, IsDisabled())
	})
}

func TestCloseIsImmutable(t *testing.T) {
	Close()
	assert.Nil(t, globalPosthogClient)

	Close()
	assert.Nil(t, globalPosthogClient)
}
