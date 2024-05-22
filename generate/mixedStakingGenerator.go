package generate

import (
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-deploy-go/data"
	mxData "github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/sharding"
)

type mixedStakingGenerator struct {
	*delegatedBaseGenerator
	numDelegatedNodes  uint
	maxNumNodesOnOwner uint
}

// NewMixedStakingGenerator will create a mixed (direct + delegated) staking generator
func NewMixedStakingGenerator(arg ArgMixedStakingGenerator) (*mixedStakingGenerator, error) {
	err := checkDelegatedStakingArgument(arg.ArgDelegatedStakingGenerator)
	if err != nil {
		return nil, err
	}
	if arg.NumDelegatedNodes == 0 {
		return nil, fmt.Errorf("%w for the NumDelegatedNodes", ErrInvalidValue)
	}
	if arg.MaxNumNodesOnOwner == 0 {
		return nil, fmt.Errorf("%w for MaxNumNodesOnOwner", ErrInvalidValue)
	}
	if check.IfNil(arg.IntRandomizer) {
		return nil, ErrNilRandomizer
	}

	msg := &mixedStakingGenerator{
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
		numDelegatedNodes:  arg.NumDelegatedNodes,
		maxNumNodesOnOwner: arg.MaxNumNodesOnOwner,
	}

	err = msg.prepareFieldsFromArguments(arg.ArgDelegatedStakingGenerator, arg.IntRandomizer)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Generate will generate data for direct stake method
func (msg *mixedStakingGenerator) Generate() (*data.GeneratorOutput, error) {
	validatorBlsKeys, observerBlsKeys, err := msg.generateValidatorAndObservers()
	if err != nil {
		return nil, err
	}

	delegators, err := msg.wkg.GenerateAdditionalKeys(int(msg.numDelegators))
	if err != nil {
		return nil, err
	}

	additionalKeys, err := msg.wkg.GenerateAdditionalKeys(int(msg.numAdditionalWalletKeys))
	if err != nil {
		return nil, err
	}

	delegatedUsedBalance := msg.prepareDelegators(delegators, int(msg.numDelegatedNodes))
	walletKeys, stakedUsedBalance, err := msg.generateWalletKeys(validatorBlsKeys)
	if err != nil {
		return nil, err
	}

	usedBalance := big.NewInt(0).Add(delegatedUsedBalance, stakedUsedBalance)
	balance := big.NewInt(0).Sub(msg.totalSupply, usedBalance)
	if balance.Cmp(zero) < 0 {
		return nil, fmt.Errorf("%w, total supply: %s, usedBalance: %s", ErrTotalSupplyTooSmall,
			msg.totalSupply.String(), usedBalance.String())
	}

	walletBalance, remainder := msg.computeWalletBalance(len(walletKeys)+len(additionalKeys), balance)
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
	gen.InitialAccounts = msg.computeInitialAccounts(walletKeys, additionalKeys, delegators)
	gen.InitialNodes = msg.computeInitialNodes(validatorBlsKeys, walletKeys)

	return gen, nil
}

func (msg *mixedStakingGenerator) generateWalletKeys(validatorBlsKeys []*data.BlsKey) ([]*data.WalletKey, *big.Int, error) {
	// the first msg.numDelegatedNodes are considered delegated. The rest are considered staked
	stakedNodes := validatorBlsKeys[msg.numDelegatedNodes:]
	walletKeys, err := msg.wkg.GenerateKeys(stakedNodes, int(msg.maxNumNodesOnOwner))
	if err != nil {
		return nil, nil, err
	}

	stakedValue := big.NewInt(int64(len(stakedNodes)))
	stakedValue.Mul(stakedValue, msg.wkg.NodePrice())

	return walletKeys, stakedValue, nil
}

func (msg *mixedStakingGenerator) computeInitialAccounts(
	walletKeys []*data.WalletKey,
	additionalKeys []*data.WalletKey,
	delegators []*data.WalletKey,
) []mxData.InitialAccount {
	initialAccounts := make([]mxData.InitialAccount, 0, len(walletKeys)+len(additionalKeys))

	for _, key := range delegators {
		account := mxData.InitialAccount{
			Address:      msg.walletPubKeyConverter.Encode(key.PubKeyBytes),
			Supply:       big.NewInt(0).Add(key.Balance, key.DelegatedValue),
			Balance:      big.NewInt(0).Set(key.Balance),
			StakingValue: big.NewInt(0),
			Delegation: &mxData.DelegationData{
				Address: msg.walletPubKeyConverter.Encode(key.DelegatedPubKeyBytes),
				Value:   big.NewInt(0).Set(key.DelegatedValue),
			},
		}

		initialAccounts = append(initialAccounts, account)
	}

	for _, key := range walletKeys {
		account := mxData.InitialAccount{
			Address:      msg.walletPubKeyConverter.Encode(key.PubKeyBytes),
			Supply:       big.NewInt(0).Add(key.Balance, key.StakedValue),
			Balance:      big.NewInt(0).Set(key.Balance),
			StakingValue: key.StakedValue,
			Delegation: &mxData.DelegationData{
				Address: "",
				Value:   big.NewInt(0),
			},
		}

		initialAccounts = append(initialAccounts, account)
	}

	for _, key := range additionalKeys {
		account := mxData.InitialAccount{
			Address:      msg.walletPubKeyConverter.Encode(key.PubKeyBytes),
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

	adjustInitialAccounts(initialAccounts)

	return initialAccounts
}

func (msg *mixedStakingGenerator) computeInitialNodes(validators []*data.BlsKey, walletKeys []*data.WalletKey) []*sharding.InitialNode {
	initialNodes := make([]*sharding.InitialNode, 0, len(validators))

	// delegated nodes
	for i := uint(0); i < msg.numDelegatedNodes; i++ {
		blsKey := validators[i]

		initialNode := &sharding.InitialNode{
			PubKey:        msg.validatorPubKeyConverter.Encode(blsKey.PubKeyBytes),
			Address:       msg.delegationScPkString,
			InitialRating: msg.initialRating,
		}
		initialNodes = append(initialNodes, initialNode)
	}

	// staked nodes
	for _, key := range walletKeys {
		initialNodes = append(initialNodes, msg.computeInitialNodesForWalletKey(key)...)
	}

	return initialNodes
}

// IsInterfaceNil returns true if there is no value under the interface
func (msg *mixedStakingGenerator) IsInterfaceNil() bool {
	return msg == nil
}
