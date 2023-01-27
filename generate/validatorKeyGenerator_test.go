package generate

import (
	"testing"

	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorKeyGenerator_GenerateKeysShouldWork(t *testing.T) {
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	keygen := signing.NewKeyGenerator(suite)
	vkg, err := NewValidatorKeyGenerator(keygen)
	require.Nil(t, err)

	numKeys := uint(10)
	keys, err := vkg.GenerateKeys(numKeys)
	require.Equal(t, int(numKeys), len(keys))
	for _, key := range keys {
		assert.Equal(t, 96, len(key.PubKeyBytes))
		assert.Equal(t, 32, len(key.PrivKeyBytes))
	}
}
