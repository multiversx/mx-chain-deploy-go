package generate

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-deploy-go/data"
	"github.com/multiversx/mx-chain-go/sharding"
)

type baseGenerator struct {
	vkg                      *validatorKeyGenerator
	wkg                      *walletKeyGenerator
	numValidatorBlsKeys      uint
	numObserverBlsKeys       uint
	richestAccountMode       bool
	numAdditionalWalletKeys  uint
	totalSupply              *big.Int
	walletPubKeyConverter    core.PubkeyConverter
	validatorPubKeyConverter core.PubkeyConverter
	initialRating            uint32
}

func (bg *baseGenerator) computeWalletBalance(numTotalWalletKeys int, balance *big.Int) (*big.Int, *big.Int) {
	// walletBalance = balance / numTotalWalletKeys
	walletBalance := big.NewInt(0).Set(balance)
	walletBalance.Div(walletBalance, big.NewInt(int64(numTotalWalletKeys)))
	// remainder = balance % numTotalWalletKeys
	remainder := big.NewInt(0).Set(balance)
	remainder.Mod(remainder, big.NewInt(int64(numTotalWalletKeys)))

	shouldReturnComputedWalletBalance := walletBalance.Cmp(minimumInitialBalance) <= 0 || !bg.richestAccountMode
	if shouldReturnComputedWalletBalance {
		return walletBalance, remainder
	}

	// remainder = balance - (minimumInitialBalance * minimumInitialBalance)
	remainder = big.NewInt(0).Set(balance)
	totalWalletBalance := big.NewInt(int64(numTotalWalletKeys))
	totalWalletBalance.Mul(totalWalletBalance, minimumInitialBalance)
	remainder.Sub(remainder, totalWalletBalance)

	return minimumInitialBalance, remainder
}

func (bg *baseGenerator) generateValidatorAndObservers() ([]*data.BlsKey, []*data.BlsKey, error) {
	validatorBlsKeys, err := bg.vkg.GenerateKeys(bg.numValidatorBlsKeys)
	if err != nil {
		return nil, nil, err
	}

	observerBlsKeys, err := bg.vkg.GenerateKeys(bg.numObserverBlsKeys)
	if err != nil {
		return nil, nil, err
	}

	return validatorBlsKeys, observerBlsKeys, nil
}

func (bg *baseGenerator) computeInitialNodesForWalletKey(key *data.WalletKey) []*sharding.InitialNode {
	initialNodes := make([]*sharding.InitialNode, 0, len(key.BlsKeys))

	for _, blsKey := range key.BlsKeys {
		validatorPubKey, _ := bg.validatorPubKeyConverter.Encode(blsKey.PubKeyBytes)
		walletAddress, _ := bg.walletPubKeyConverter.Encode(key.PubKeyBytes)

		initialNode := &sharding.InitialNode{
			PubKey:        validatorPubKey,
			Address:       walletAddress,
			InitialRating: bg.initialRating,
		}
		initialNodes = append(initialNodes, initialNode)
	}

	return initialNodes
}
