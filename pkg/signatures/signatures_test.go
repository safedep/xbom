package signatures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignatures(t *testing.T) {
	_, err := LoadAllSignatures()
	assert.NoError(t, err, "Signatures should be valid")
}
