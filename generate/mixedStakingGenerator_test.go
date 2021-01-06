package generate

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-deploy-go/check"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockMixedStakingGeneratorArguments() ArgMixedStakingGenerator {
	mclSuite := mcl.NewSuiteBLS12()
	edSuite := ed25519.NewEd25519()

	arg := ArgDelegatedStakingGenerator{
		KeyGeneratorForValidators: signing.NewKeyGenerator(mclSuite),
		KeyGeneratorForWallets:    signing.NewKeyGenerator(edSuite),
		NumValidatorBlsKeys:       0,
		NumObserverBlsKeys:        0,
		RichestAccountMode:        false,
		NumAdditionalWalletKeys:   0,
		InitialRating:             50,
		DelegationOwnerPkString:   "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
		DelegationOwnerNonce:      0,
		VmType:                    "0500",
		NumDelegators:             0,
	}
	arg.WalletPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32)
	arg.ValidatorPubKeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(96)
	arg.TotalSupply = big.NewInt(0)
	arg.TotalSupply.SetString("20000000000000000000000000", 10)
	arg.NodePrice = big.NewInt(0)
	arg.NodePrice.SetString("2500000000000000000000", 10)

	return ArgMixedStakingGenerator{
		ArgDelegatedStakingGenerator: arg,
		NumDelegatedNodes:            0,
		MaxNumNodesOnOwner:           0,
	}
}

func TestMixedStakingGenerator_GenerateShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockMixedStakingGeneratorArguments()
	arg.NumValidatorBlsKeys = 33
	arg.NumObserverBlsKeys = 3
	arg.NumDelegators = 47
	arg.NumDelegatedNodes = 9
	arg.MaxNumNodesOnOwner = 1
	arg.NumAdditionalWalletKeys = 3

	msg, err := NewMixedStakingGenerator(arg)
	require.Nil(t, err)

	generatedOutput, err := msg.Generate()
	require.Nil(t, err)

	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.ValidatorBlsKeys))
	assert.Equal(t, int(arg.NumObserverBlsKeys), len(generatedOutput.ObserverBlsKeys))
	expectedNumInitialAccounts := arg.NumValidatorBlsKeys + arg.NumObserverBlsKeys + arg.NumDelegators - arg.NumDelegatedNodes
	assert.Equal(t, int(expectedNumInitialAccounts), len(generatedOutput.InitialAccounts))
	assert.Equal(t, int(arg.NumValidatorBlsKeys-arg.NumDelegatedNodes), len(generatedOutput.WalletKeys))
	assert.Equal(t, int(arg.NumAdditionalWalletKeys), len(generatedOutput.AdditionalKeys))
	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.InitialNodes))
	assert.Equal(t, int(arg.NumDelegators), len(generatedOutput.DelegatorKeys))

	iac, _ := check.NewInitialAccountsChecker(arg.NodePrice, arg.TotalSupply)
	assert.Nil(t, err, iac.CheckInitialAccounts(generatedOutput.InitialAccounts))
}

func TestMixedStakingGenerator_GenerateWithRichestAccountShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockMixedStakingGeneratorArguments()
	arg.NumValidatorBlsKeys = 33
	arg.NumObserverBlsKeys = 3
	arg.NumDelegators = 47
	arg.NumDelegatedNodes = 9
	arg.MaxNumNodesOnOwner = 1
	arg.NumAdditionalWalletKeys = 3
	arg.RichestAccountMode = true

	msg, err := NewMixedStakingGenerator(arg)
	require.Nil(t, err)

	generatedOutput, err := msg.Generate()
	require.Nil(t, err)

	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.ValidatorBlsKeys))
	assert.Equal(t, int(arg.NumObserverBlsKeys), len(generatedOutput.ObserverBlsKeys))
	expectedNumInitialAccounts := arg.NumValidatorBlsKeys + arg.NumObserverBlsKeys + arg.NumDelegators - arg.NumDelegatedNodes
	assert.Equal(t, int(expectedNumInitialAccounts), len(generatedOutput.InitialAccounts))
	assert.Equal(t, int(arg.NumValidatorBlsKeys-arg.NumDelegatedNodes), len(generatedOutput.WalletKeys))
	assert.Equal(t, int(arg.NumAdditionalWalletKeys), len(generatedOutput.AdditionalKeys))
	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.InitialNodes))
	assert.Equal(t, int(arg.NumDelegators), len(generatedOutput.DelegatorKeys))

	iac, _ := check.NewInitialAccountsChecker(arg.NodePrice, arg.TotalSupply)
	assert.Nil(t, err, iac.CheckInitialAccounts(generatedOutput.InitialAccounts))
	for i, ia := range generatedOutput.InitialAccounts {
		if i == int(arg.NumDelegators) {
			assert.NotEqual(t, minimumInitialBalance, ia.Balance)
		} else {
			assert.Equal(t, minimumInitialBalance, ia.Balance)
		}
	}
}
