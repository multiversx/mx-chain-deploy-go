package generate

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-deploy-go/data"
	"github.com/multiversx/mx-chain-deploy-go/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWalletKeyGenerator(t *testing.T) {
	t.Parallel()

	suite := ed25519.NewEd25519()
	keygen := signing.NewKeyGenerator(suite)
	numToReturn := -1
	intRandomizer := &mock.IntRandomizerStub{
		IntnCalled: func(n int) int {
			numToReturn++
			return numToReturn
		},
	}
	nodePrice := big.NewInt(2500)

	t.Run("nil key generator should error", func(t *testing.T) {
		t.Parallel()

		vkg, err := NewWalletKeyGenerator(nil, intRandomizer, nodePrice, 1, false)
		assert.Nil(t, vkg)
		assert.Equal(t, ErrNilKeyGenerator, err)
	})
	t.Run("nil randomizer should error", func(t *testing.T) {
		t.Parallel()

		vkg, err := NewWalletKeyGenerator(keygen, nil, nodePrice, 1, false)
		assert.Nil(t, vkg)
		assert.Equal(t, ErrNilRandomizer, err)
	})
	t.Run("nil node price should error", func(t *testing.T) {
		t.Parallel()

		vkg, err := NewWalletKeyGenerator(keygen, intRandomizer, nil, 1, false)
		assert.Nil(t, vkg)
		assert.Equal(t, ErrNilNodePrice, err)
	})
	t.Run("num shards is 0 should error", func(t *testing.T) {
		t.Parallel()

		vkg, err := NewWalletKeyGenerator(keygen, intRandomizer, nodePrice, 0, true)
		assert.Nil(t, vkg)
		assert.Equal(t, ErrNumShardsIsZero, err)

		vkg, err = NewWalletKeyGenerator(keygen, intRandomizer, nodePrice, 0, false)
		assert.Nil(t, vkg)
		assert.Equal(t, ErrNumShardsIsZero, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		vkg, err := NewWalletKeyGenerator(keygen, intRandomizer, nodePrice, 1, true)
		assert.NotNil(t, vkg)
		assert.Nil(t, err)

		vkg, err = NewWalletKeyGenerator(keygen, intRandomizer, nodePrice, 1, false)
		assert.NotNil(t, vkg)
		assert.Nil(t, err)
	})
}

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
	vkg, err := NewWalletKeyGenerator(keygen, intRandomizer, nodePrice, 1, false)
	require.Nil(t, err)

	keys, err := vkg.GenerateKeys(blsKeys, 2)
	assert.Equal(t, 9, len(keys))

	for i, key := range keys {
		assert.Equal(t, i+1, len(key.BlsKeys))
		assert.Equal(t, big.NewInt(0).Mul(nodePrice, big.NewInt(int64(i+1))), key.StakedValue)
	}
}

func TestWalletKeyGenerator_GenerateKeysWithOneKeyShouldWork(t *testing.T) {
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
	vkg, err := NewWalletKeyGenerator(keygen, &mock.IntRandomizerStub{}, nodePrice, 1, false)
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

	numToReturn := -1
	intRandomizer := &mock.IntRandomizerStub{
		IntnCalled: func(n int) int {
			numToReturn++
			return numToReturn
		},
	}

	t.Run("generateInAllShards is false should generate in all shards", func(t *testing.T) {
		t.Parallel()

		vkg, err := NewWalletKeyGenerator(keygen, intRandomizer, nodePrice, 1, false)
		require.Nil(t, err)

		numKeys := 100
		keys, err := vkg.GenerateAdditionalKeys(numKeys)
		assert.Equal(t, numKeys, len(keys))

		mapShards := make(map[byte]int)
		for _, key := range keys {
			mapShards[key.PubKeyBytes[len(key.PubKeyBytes)-1]]++
		}

		assert.Greater(t, len(mapShards), 1)
	})
	t.Run("generateInAllShards is true and num shard == 1 should generate in shard 0", func(t *testing.T) {
		t.Parallel()

		vkg, err := NewWalletKeyGenerator(keygen, intRandomizer, nodePrice, 1, true)
		require.Nil(t, err)

		numKeys := 100
		keys, err := vkg.GenerateAdditionalKeys(numKeys)
		assert.Equal(t, numKeys, len(keys))

		mapShards := make(map[byte]int)
		for _, key := range keys {
			mapShards[key.PubKeyBytes[len(key.PubKeyBytes)-1]]++
		}

		assert.Equal(t, 1, len(mapShards))
		assert.Equal(t, numKeys, mapShards[0])
	})
	t.Run("generateInAllShards is true and num shard == 2 should generate in shard 0 & 1", func(t *testing.T) {
		t.Parallel()

		vkg, err := NewWalletKeyGenerator(keygen, intRandomizer, nodePrice, 2, true)
		require.Nil(t, err)

		numKeys := 100
		keys, err := vkg.GenerateAdditionalKeys(numKeys)
		assert.Equal(t, numKeys, len(keys))

		mapShards := make(map[byte]int)
		for _, key := range keys {
			mapShards[key.PubKeyBytes[len(key.PubKeyBytes)-1]]++
		}

		assert.Equal(t, 2, len(mapShards))
		assert.Equal(t, numKeys/2, mapShards[0])
		assert.Equal(t, numKeys/2, mapShards[1])
	})
}
