package generate

import (
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-deploy-go/core"
	"github.com/multiversx/mx-chain-deploy-go/data"
)

type delegatedBaseGenerator struct {
	*baseGenerator
	delegationScPkString string
	delegationScPkBytes  []byte
	numDelegators        uint
}

func checkDelegatedStakingArgument(arg ArgDelegatedStakingGenerator) error {
	if check.IfNil(arg.WalletPubKeyConverter) {
		return fmt.Errorf("%w for the WalletPubKeyConverter", ErrNilPubKeyConverter)
	}
	if check.IfNil(arg.ValidatorPubKeyConverter) {
		return fmt.Errorf("%w for the ValidatorPubKeyConverter", ErrNilPubKeyConverter)
	}
	if arg.NumDelegators == 0 {
		return fmt.Errorf("%w for the NumDelegators", ErrInvalidValue)
	}

	return nil
}

func (dbs *delegatedBaseGenerator) prepareFieldsFromArguments(arg ArgDelegatedStakingGenerator, randomizer IntRandomizer) error {
	var err error
	dbs.vkg, err = NewValidatorKeyGenerator(arg.KeyGeneratorForValidators)
	if err != nil {
		return err
	}

	dbs.wkg, err = NewWalletKeyGenerator(arg.KeyGeneratorForWallets, randomizer, arg.NodePrice)
	if err != nil {
		return err
	}

	dbs.delegationScPkString, err = core.GenerateSCAddress(
		arg.DelegationOwnerPkString,
		arg.DelegationOwnerNonce,
		arg.VmType,
		arg.WalletPubKeyConverter,
	)
	if err != nil {
		return err
	}

	dbs.delegationScPkBytes, err = arg.WalletPubKeyConverter.Decode(dbs.delegationScPkString)
	if err != nil {
		return err
	}

	return nil
}

func (dbs *delegatedBaseGenerator) prepareDelegators(delegators []*data.WalletKey, numDelegated int) *big.Int {
	// totalDelegated = numDelegated * nodePrice
	totalDelegated := big.NewInt(int64(numDelegated))
	totalDelegated.Mul(totalDelegated, dbs.wkg.NodePrice())
	// delegated = totalDelegated / len(delegators)
	delegated := big.NewInt(0).Set(totalDelegated)
	delegated.Div(delegated, big.NewInt(int64(len(delegators))))
	// remainder = totalDelegated % len(delegators)
	remainder := big.NewInt(0).Set(totalDelegated)
	remainder.Mod(remainder, big.NewInt(int64(len(delegators))))

	for i, wallet := range delegators {
		wallet.DelegatedPubKeyBytes = make([]byte, len(dbs.delegationScPkBytes))
		copy(wallet.DelegatedPubKeyBytes, dbs.delegationScPkBytes)
		wallet.DelegatedValue = big.NewInt(0).Set(delegated)
		if i == 0 {
			wallet.DelegatedValue.Add(wallet.DelegatedValue, remainder)
		}
		// give a little balance to each delegator so it can claim the rewards
		wallet.Balance = minimumInitialBalance
		totalDelegated.Add(totalDelegated, minimumInitialBalance)
	}

	return totalDelegated
}
