// +build sgx_enclave

package attestation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuote(t *testing.T) {
	result, err := Quote("hi dad")
	assert.NoError(t, err)
	assert.Equal(t, "hi dad 0x0", result)
}
