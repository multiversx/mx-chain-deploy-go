package generating

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-deploy-go/core"
	"github.com/ElrondNetwork/elrond-deploy-go/data"
	"github.com/ElrondNetwork/elrond-deploy-go/generating/disabled"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	elrondData "github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type delegatedStakingGenerator struct {
	*baseGenerator
	delegationScPkString string
	delegationScPkBytes  []byte
	numDelegators        uint
}

// NewDelegatedGenerator will create a delegated staking generator
func NewDelegatedGenerator(arg ArgDelegatedStakingGenerator) (*delegatedStakingGenerator, error) {
	if check.IfNil(arg.WalletPubKeyConverter) {
		return nil, fmt.Errorf("%w for the WalletPubKeyConverter", ErrNilPubKeyConverter)
	}
	if check.IfNil(arg.ValidatorPubKeyConverter) {
		return nil, fmt.Errorf("%w for the ValidatorPubKeyConverter", ErrNilPubKeyConverter)
	}
	if arg.NumDelegators == 0 {
		return nil, fmt.Errorf("%w for the NumDelegators", ErrInvalidValue)
	}

	dsg := &delegatedStakingGenerator{
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
	}
	var err error
	dsg.vkg, err = NewValidatorKeyGenerator(arg.KeyGeneratorForValidators)
	if err != nil {
		return nil, err
	}

	dsg.wkg, err = NewWalletKeyGenerator(arg.KeyGeneratorForWallets, &disabled.NilRandomizer{}, arg.NodePrice)
	if err != nil {
		return nil, err
	}

	dsg.delegationScPkString, err = core.GenerateSCAddress(
		arg.DelegationOwnerPkString,
		arg.DelegationOwnerNonce,
		arg.VmType,
		arg.WalletPubKeyConverter,
	)
	if err != nil {
		return nil, err
	}

	dsg.delegationScPkBytes, err = arg.WalletPubKeyConverter.Decode(dsg.delegationScPkString)
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

	usedBalance := dsg.prepareDelegators(delegators, validatorBlsKeys)
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

func (dsg *delegatedStakingGenerator) prepareDelegators(delegators []*data.WalletKey, validators []*data.BlsKey) *big.Int {
	//totalDelegated = len(validators) * nodePrice
	totalDelegated := big.NewInt(int64(len(validators)))
	totalDelegated.Mul(totalDelegated, dsg.wkg.NodePrice())
	//delegated = totalDelegated / len(delegators)
	delegated := big.NewInt(0).Set(totalDelegated)
	delegated.Div(delegated, big.NewInt(int64(len(delegators))))
	//remainder = totalDelegated % len(delegators)
	remainder := big.NewInt(0).Set(totalDelegated)
	remainder.Mod(remainder, big.NewInt(int64(len(delegators))))

	for i, wallet := range delegators {
		wallet.DelegatedPubKeyBytes = make([]byte, len(dsg.delegationScPkBytes))
		copy(wallet.DelegatedPubKeyBytes, dsg.delegationScPkBytes)
		wallet.DelegatedValue = big.NewInt(0).Set(delegated)
		if i == 0 {
			wallet.DelegatedValue.Add(wallet.DelegatedValue, remainder)
		}
		//give a little balance to each delegator so it can claim the rewards
		wallet.Balance = minimumInitialBalance
		totalDelegated.Add(totalDelegated, minimumInitialBalance)
	}

	return totalDelegated
}

func (dsg *delegatedStakingGenerator) computeInitialAccounts(
	walletKeys []*data.WalletKey,
	additionalKeys []*data.WalletKey,
	delegators []*data.WalletKey,
) []elrondData.InitialAccount {
	initialAccounts := make([]elrondData.InitialAccount, 0, len(walletKeys)+len(additionalKeys))

	for _, key := range delegators {
		account := elrondData.InitialAccount{
			Address:      dsg.walletPubKeyConverter.Encode(key.PubKeyBytes),
			Supply:       big.NewInt(0).Add(key.Balance, key.DelegatedValue),
			Balance:      big.NewInt(0).Set(key.Balance),
			StakingValue: big.NewInt(0),
			Delegation: &elrondData.DelegationData{
				Address: dsg.walletPubKeyConverter.Encode(key.DelegatedPubKeyBytes),
				Value:   big.NewInt(0).Set(key.DelegatedValue),
			},
		}

		initialAccounts = append(initialAccounts, account)
	}

	for _, key := range walletKeys {
		account := elrondData.InitialAccount{
			Address:      dsg.walletPubKeyConverter.Encode(key.PubKeyBytes),
			Supply:       big.NewInt(0).Set(key.Balance),
			Balance:      big.NewInt(0).Set(key.Balance),
			StakingValue: big.NewInt(0),
			Delegation: &elrondData.DelegationData{
				Address: "",
				Value:   big.NewInt(0),
			},
		}

		initialAccounts = append(initialAccounts, account)
	}

	for _, key := range additionalKeys {
		account := elrondData.InitialAccount{
			Address:      dsg.walletPubKeyConverter.Encode(key.PubKeyBytes),
			Supply:       big.NewInt(0).Set(key.Balance),
			Balance:      big.NewInt(0).Set(key.Balance),
			StakingValue: big.NewInt(0),
			Delegation: &elrondData.DelegationData{
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
			PubKey:        dsg.validatorPubKeyConverter.Encode(blsKey.PubKeyBytes),
			Address:       dsg.delegationScPkString,
			InitialRating: dsg.initialRating,
		}
		initialNodes = append(initialNodes, initialNode)
	}

	return initialNodes
}

// IsInterfaceNil returns if underlying object is true
func (dsg *delegatedStakingGenerator) IsInterfaceNil() bool {
	return dsg == nil
}
