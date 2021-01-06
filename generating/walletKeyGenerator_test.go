package generating

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-deploy-go/data"
	"github.com/ElrondNetwork/elrond-deploy-go/mock"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletKeyGenerator_GenerateKeysShouldWork(t *testing.T) {
	t.Parallel()

	suite := ed25519.NewEd25519()
	keygen := signing.NewKeyGenerator(suite)

	numBlsKeys := 45
	blsKeys := make([]*data.BlsKey, 0)
	for i := 0; i < numBlsKeys; i++ {
		blsKeys = append(blsKeys, &data.BlsKey{
			PubKeyBytes: []byte("pubkey"),
		})
	}

	numToReturn := -1
	intRandomizer := &mock.IntRandomizerStub{
		IntnCalled: func(n int) int {
			numToReturn++
			return numToReturn
		},
	}

	nodePrice := big.NewInt(2500)
	vkg, err := NewWalletKeyGenerator(keygen, intRandomizer, nodePrice)
	require.Nil(t, err)

	keys, err := vkg.GenerateKeys(blsKeys, 2)
	assert.Equal(t, 9, len(keys))

	for i, key := range keys {
		assert.Equal(t, i+1, len(key.BlsKeys))
		assert.Equal(t, big.NewInt(0).Mul(nodePrice, big.NewInt(int64(i+1))), key.StakedValue)
	}
}

func TestWalletKeyGenerator_GenerateKeysWithOneBlsKeyShouldWork(t *testing.T) {
	t.Parallel()

	suite := ed25519.NewEd25519()
	keygen := signing.NewKeyGenerator(suite)

	numBlsKeys := 45
	blsKeys := make([]*data.BlsKey, 0)
	for i := 0; i < numBlsKeys; i++ {
		blsKeys = append(blsKeys, &data.BlsKey{
			PubKeyBytes: []byte("pubkey"),
		})
	}

	nodePrice := big.NewInt(2500)
	vkg, err := NewWalletKeyGenerator(keygen, &mock.IntRandomizerStub{}, nodePrice)
	require.Nil(t, err)

	keys, err := vkg.GenerateKeys(blsKeys, 1)
	assert.Equal(t, numBlsKeys, len(keys))

	for _, key := range keys {
		assert.Equal(t, 1, len(key.BlsKeys))
		assert.Equal(t, nodePrice, key.StakedValue)
	}
}

func TestWalletKeyGenerator_GenerateAdditionalKeysShouldWork(t *testing.T) {
	t.Parallel()

	suite := ed25519.NewEd25519()
	keygen := signing.NewKeyGenerator(suite)

	nodePrice := big.NewInt(2500)
	vkg, err := NewWalletKeyGenerator(keygen, &mock.IntRandomizerStub{}, nodePrice)
	require.Nil(t, err)

	numKeys := 100
	keys, err := vkg.GenerateAdditionalKeys(numKeys)
	assert.Equal(t, numKeys, len(keys))
}
