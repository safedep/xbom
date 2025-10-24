package signatures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignatures(t *testing.T) {
	sigs, err := LoadAllSignatures()

	assert.NoError(t, err, "Signatures should be valid")
	assert.Equal(t, len(sigs), 0, "No signature is actually loaded here")
}
