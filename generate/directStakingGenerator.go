package generate

import (
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-deploy-go/data"
	mxData "github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/sharding"
)

var minimumInitialBalance = big.NewInt(1000000000000000000) // 1eGLD
var zero = big.NewInt(0)

type directStakingGenerator struct {
	*baseGenerator
	maxNumNodesOnOwner uint
}

// NewDirectStakingGenerator will create a direct staking generator
func NewDirectStakingGenerator(arg ArgDirectStakingGenerator) (*directStakingGenerator, error) {
	if arg.MaxNumNodesOnOwner == 0 {
		return nil, fmt.Errorf("%w for MaxNumNodesOnOwner", ErrInvalidValue)
	}
	if check.IfNil(arg.WalletPubKeyConverter) {
		return nil, fmt.Errorf("%w for the WalletPubKeyConverter", ErrNilPubKeyConverter)
	}
	if check.IfNil(arg.ValidatorPubKeyConverter) {
		return nil, fmt.Errorf("%w for the ValidatorPubKeyConverter", ErrNilPubKeyConverter)
	}
	if check.IfNil(arg.IntRandomizer) {
		return nil, ErrNilRandomizer
	}

	dsg := &directStakingGenerator{
		baseGenerator: &baseGenerator{
			numValidatorBlsKeys:      arg.NumValidatorBlsKeys,
			numObserverBlsKeys:       arg.NumObserverBlsKeys,
			richestAccountMode:       arg.RichestAccountMode,
			numAdditionalWalletKeys:  arg.NumAdditionalWalletKeys,
			totalSupply:              arg.TotalSupply,
			walletPubKeyConverter:    arg.WalletPubKeyConverter,
			validatorPubKeyConverter: arg.ValidatorPubKeyConverter,
		},
		maxNumNodesOnOwner: arg.MaxNumNodesOnOwner,
	}
	var err error
	dsg.vkg, err = NewValidatorKeyGenerator(arg.KeyGeneratorForValidators)
	if err != nil {
		return nil, err
	}

	dsg.wkg, err = NewWalletKeyGenerator(arg.KeyGeneratorForWallets, arg.IntRandomizer, arg.NodePrice)
	if err != nil {
		return nil, err
	}

	return dsg, nil
}

// Generate will generate data for direct stake method
func (dsg *directStakingGenerator) Generate() (*data.GeneratorOutput, error) {
	validatorBlsKeys, observerBlsKeys, err := dsg.generateValidatorAndObservers()
	if err != nil {
		return nil, err
	}

	walletKeys, err := dsg.wkg.GenerateKeys(validatorBlsKeys, int(dsg.maxNumNodesOnOwner))
	if err != nil {
		return nil, err
	}

	additionalKeys, err := dsg.wkg.GenerateAdditionalKeys(int(dsg.numAdditionalWalletKeys))
	if err != nil {
		return nil, err
	}

	if len(walletKeys)+len(additionalKeys) == 0 {
		return nil, ErrInvalidNumberOfWalletKeys
	}

	usedBalance := dsg.computeUsedBalance(walletKeys)
	balance := big.NewInt(0).Sub(dsg.totalSupply, usedBalance)
	if balance.Cmp(zero) < 0 {
		return nil, fmt.Errorf("%w, total supply: %s, usedBalance: %s", ErrTotalSupplyTooSmall,
			dsg.totalSupply.String(), usedBalance.String())
	}

	walletBalance, remainder := dsg.computeWalletBalance(len(walletKeys)+len(additionalKeys), balance)

	for i, key := range walletKeys {
		key.Balance = big.NewInt(0).Set(walletBalance)
		if i == 0 {
			key.Balance.Add(key.Balance, remainder)
		}
	}

	for _, key := range additionalKeys {
		key.Balance = big.NewInt(0).Set(walletBalance)
	}

	gen := &data.GeneratorOutput{
		ValidatorBlsKeys: validatorBlsKeys,
		ObserverBlsKeys:  observerBlsKeys,
		WalletKeys:       walletKeys,
		AdditionalKeys:   additionalKeys,
	}
	gen.InitialAccounts = dsg.computeInitialAccounts(walletKeys, additionalKeys)
	gen.InitialNodes = dsg.computeInitialNodes(walletKeys)

	return gen, nil
}

func (dsg *directStakingGenerator) computeUsedBalance(walletKeys []*data.WalletKey) *big.Int {
	staked := big.NewInt(0)
	for _, key := range walletKeys {
		staked.Add(staked, key.StakedValue)
	}

	return staked
}

func (dsg *directStakingGenerator) computeInitialAccounts(
	walletKeys []*data.WalletKey,
	additionalKeys []*data.WalletKey,
) []mxData.InitialAccount {
	initialAccounts := make([]mxData.InitialAccount, 0, len(walletKeys)+len(additionalKeys))
	for _, key := range walletKeys {
		account := mxData.InitialAccount{
			Address:      dsg.walletPubKeyConverter.SilentEncode(key.PubKeyBytes, log),
			Supply:       big.NewInt(0).Add(key.Balance, key.StakedValue),
			Balance:      big.NewInt(0).Set(key.Balance),
			StakingValue: big.NewInt(0).Set(key.StakedValue),
			Delegation: &mxData.DelegationData{
				Address: "",
				Value:   big.NewInt(0),
			},
		}

		initialAccounts = append(initialAccounts, account)
	}

	for _, key := range additionalKeys {
		account := mxData.InitialAccount{
			Address:      dsg.walletPubKeyConverter.SilentEncode(key.PubKeyBytes, log),
			Supply:       big.NewInt(0).Set(key.Balance),
			Balance:      big.NewInt(0).Set(key.Balance),
			StakingValue: big.NewInt(0),
			Delegation: &mxData.DelegationData{
				Address: "",
				Value:   big.NewInt(0),
			},
		}

		initialAccounts = append(initialAccounts, account)
	}

	return initialAccounts
}

func (dsg *directStakingGenerator) computeInitialNodes(walletKeys []*data.WalletKey) []*sharding.InitialNode {
	initialNodes := make([]*sharding.InitialNode, 0)
	for _, key := range walletKeys {
		initialNodes = append(initialNodes, dsg.computeInitialNodesForWalletKey(key)...)
	}

	return initialNodes
}

// IsInterfaceNil returns true if there is no value under the interface
func (dsg *directStakingGenerator) IsInterfaceNil() bool {
	return dsg == nil
}
