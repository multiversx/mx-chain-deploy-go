package generate

import (
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-deploy-go/data"
	"github.com/multiversx/mx-chain-deploy-go/generate/disabled"
	mxData "github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/sharding"
)

type delegatedStakingGenerator struct {
	*delegatedBaseGenerator
}

// NewDelegatedGenerator will create a delegated staking generator
func NewDelegatedGenerator(arg ArgDelegatedStakingGenerator) (*delegatedStakingGenerator, error) {
	err := checkDelegatedStakingArgument(arg)
	if err != nil {
		return nil, err
	}

	dsg := &delegatedStakingGenerator{
		delegatedBaseGenerator: &delegatedBaseGenerator{
			baseGenerator: &baseGenerator{
				numValidatorBlsKeys:      arg.NumValidatorBlsKeys,
				numObserverBlsKeys:       arg.NumObserverBlsKeys,
				richestAccountMode:       arg.RichestAccountMode,
				numAdditionalWalletKeys:  arg.NumAdditionalWalletKeys,
				totalSupply:              arg.TotalSupply,
				walletPubKeyConverter:    arg.WalletPubKeyConverter,
				validatorPubKeyConverter: arg.ValidatorPubKeyConverter,
			},
			numDelegators: arg.NumDelegators,
		},
	}

	err = dsg.prepareFieldsFromArguments(arg, &disabled.NilRandomizer{})
	if err != nil {
		return nil, err
	}

	return dsg, nil
}

// Generate will generate data for direct stake method
func (dsg *delegatedStakingGenerator) Generate() (*data.GeneratorOutput, error) {
	validatorBlsKeys, observerBlsKeys, err := dsg.generateValidatorAndObservers()
	if err != nil {
		return nil, err
	}

	walletKeys, err := dsg.wkg.GenerateAdditionalKeys(len(validatorBlsKeys))
	if err != nil {
		return nil, err
	}

	delegators, err := dsg.wkg.GenerateAdditionalKeys(int(dsg.numDelegators))
	if err != nil {
		return nil, err
	}

	additionalKeys, err := dsg.wkg.GenerateAdditionalKeys(int(dsg.numAdditionalWalletKeys))
	if err != nil {
		return nil, err
	}

	if len(walletKeys) == 0 {
		return nil, ErrInvalidNumberOfWalletKeys
	}

	usedBalance := dsg.prepareDelegators(delegators, len(validatorBlsKeys))
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
		DelegatorKeys:    delegators,
	}
	gen.InitialAccounts = dsg.computeInitialAccounts(walletKeys, additionalKeys, delegators)
	gen.InitialNodes = dsg.computeInitialNodes(validatorBlsKeys)

	return gen, nil
}

func (dsg *delegatedStakingGenerator) computeInitialAccounts(
	walletKeys []*data.WalletKey,
	additionalKeys []*data.WalletKey,
	delegators []*data.WalletKey,
) []mxData.InitialAccount {
	initialAccounts := make([]mxData.InitialAccount, 0, len(walletKeys)+len(additionalKeys))

	for _, key := range delegators {
		account := mxData.InitialAccount{
			Address:      dsg.walletPubKeyConverter.SilentEncode(key.PubKeyBytes, log),
			Supply:       big.NewInt(0).Add(key.Balance, key.DelegatedValue),
			Balance:      big.NewInt(0).Set(key.Balance),
			StakingValue: big.NewInt(0),
			Delegation: &mxData.DelegationData{
				Address: dsg.walletPubKeyConverter.SilentEncode(key.DelegatedPubKeyBytes, log),
				Value:   big.NewInt(0).Set(key.DelegatedValue),
			},
		}

		initialAccounts = append(initialAccounts, account)
	}

	for _, key := range walletKeys {
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

func (dsg *delegatedStakingGenerator) computeInitialNodes(validators []*data.BlsKey) []*sharding.InitialNode {
	initialNodes := make([]*sharding.InitialNode, 0, len(validators))

	for _, blsKey := range validators {
		initialNode := &sharding.InitialNode{
			PubKey:        dsg.validatorPubKeyConverter.SilentEncode(blsKey.PubKeyBytes, log),
			Address:       dsg.delegationScPkString,
			InitialRating: dsg.initialRating,
		}
		initialNodes = append(initialNodes, initialNode)
	}

	return initialNodes
}

// IsInterfaceNil returns true if there is no value under the interface
func (dsg *delegatedStakingGenerator) IsInterfaceNil() bool {
	return dsg == nil
}
