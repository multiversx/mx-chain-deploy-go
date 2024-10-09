package generate

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-deploy-go/data"
)

type walletKeyGenerator struct {
	keyGen              crypto.KeyGenerator
	randomizer          IntRandomizer
	nodePrice           *big.Int
	mut                 sync.Mutex
	crtShard            byte
	numShards           uint32
	generateInAllShards bool
}

// NewWalletKeyGenerator will create a new instance for the wallet key generator
func NewWalletKeyGenerator(
	keyGen crypto.KeyGenerator,
	randomizer IntRandomizer,
	nodePrice *big.Int,
	numShards uint32,
	generateInAllShards bool,
) (*walletKeyGenerator, error) {
	if check.IfNil(keyGen) {
		return nil, ErrNilKeyGenerator
	}
	if check.IfNil(randomizer) {
		return nil, ErrNilRandomizer
	}
	if nodePrice == nil {
		return nil, ErrNilNodePrice
	}
	if numShards == 0 {
		return nil, ErrNumShardsIsZero
	}

	return &walletKeyGenerator{
		keyGen:              keyGen,
		randomizer:          randomizer,
		nodePrice:           nodePrice,
		generateInAllShards: generateInAllShards,
		numShards:           numShards,
	}, nil
}

// GenerateKeys will generate the number of keys provided
func (wkg *walletKeyGenerator) GenerateKeys(blsKeys []*data.BlsKey, maxNumKeysOnOwner int) ([]*data.WalletKey, error) {
	if maxNumKeysOnOwner < 1 {
		return nil, fmt.Errorf("%w for maxNumKeysOnOwner", ErrInvalidValue)
	}

	keys := make([]*data.WalletKey, 0)
	blsKeysPool := make([]*data.BlsKey, len(blsKeys))
	copy(blsKeysPool, blsKeys)

	for {
		if len(blsKeysPool) == 0 {
			break
		}

		numKeysOnOwner := maxNumKeysOnOwner
		if maxNumKeysOnOwner > 1 {
			// create a random number
			numKeysOnOwner = wkg.randomizer.Intn(maxNumKeysOnOwner-1) + 1
			if numKeysOnOwner > len(blsKeysPool) {
				numKeysOnOwner = len(blsKeysPool)
			}
		}

		extractedBlsKeys := blsKeysPool[:numKeysOnOwner]
		blsKeysPool = blsKeysPool[numKeysOnOwner:]

		walletKey, err := wkg.generateWalletKey()
		if err != nil {
			return nil, err
		}
		walletKey.BlsKeys = extractedBlsKeys
		walletKey.StakedValue = big.NewInt(0).Mul(wkg.nodePrice, big.NewInt(int64(len(extractedBlsKeys))))

		keys = append(keys, walletKey)
	}

	return keys, nil
}

func (wkg *walletKeyGenerator) generateWalletKey() (*data.WalletKey, error) {
	var err error

	sk, pkBytes, err := wkg.generateKey()
	if err != nil {
		return nil, err
	}

	walletKey := &data.WalletKey{
		PubKeyBytes: pkBytes,
	}

	walletKey.PrivKeyBytes, err = sk.ToByteArray()
	if err != nil {
		return nil, err
	}

	return walletKey, nil
}

func (wkg *walletKeyGenerator) generateKey() (crypto.PrivateKey, []byte, error) {
	wkg.mut.Lock()
	defer func() {
		wkg.crtShard++
		wkg.crtShard = wkg.crtShard % byte(wkg.numShards)

		wkg.mut.Unlock()
	}()

	for {
		sk, pk := wkg.keyGen.GeneratePair()

		pkBytes, err := pk.ToByteArray()
		if err != nil {
			return nil, nil, err
		}

		if wkg.generateInAllShards && pkBytes[len(pkBytes)-1] != wkg.crtShard {
			continue
		}

		return sk, pkBytes, nil
	}
}

// GenerateAdditionalKeys will generate the additional wallet keys
func (wkg *walletKeyGenerator) GenerateAdditionalKeys(numKeys int) ([]*data.WalletKey, error) {
	keys := make([]*data.WalletKey, 0, numKeys)

	for i := 0; i < numKeys; i++ {
		walletKey, err := wkg.generateWalletKey()
		if err != nil {
			return nil, err
		}

		keys = append(keys, walletKey)
	}

	return keys, nil
}

// NodePrice returns the initial node price
func (wkg *walletKeyGenerator) NodePrice() *big.Int {
	return wkg.nodePrice
}
