package generate

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-deploy-go/data"
)

type validatorKeyGenerator struct {
	keyGen crypto.KeyGenerator
}

// NewValidatorKeyGenerator will create a new instance for the validator key generator
func NewValidatorKeyGenerator(keyGen crypto.KeyGenerator) (*validatorKeyGenerator, error) {
	if check.IfNil(keyGen) {
		return nil, ErrNilKeyGenerator
	}

	return &validatorKeyGenerator{
		keyGen: keyGen,
	}, nil
}

// GenerateKeys will generate the number of keys provided
func (vkg *validatorKeyGenerator) GenerateKeys(numKeys uint) ([]*data.BlsKey, error) {
	keys := make([]*data.BlsKey, 0, numKeys)

	var err error
	for i := uint(0); i < numKeys; i++ {
		sk, pk := vkg.keyGen.GeneratePair()
		blsKey := &data.BlsKey{}

		blsKey.PrivKeyBytes, err = sk.ToByteArray()
		if err != nil {
			return nil, fmt.Errorf("%w at index %d", err, i)
		}

		blsKey.PubKeyBytes, err = pk.ToByteArray()
		if err != nil {
			return nil, fmt.Errorf("%w at index %d", err, i)
		}

		keys = append(keys, blsKey)
	}

	return keys, nil
}
