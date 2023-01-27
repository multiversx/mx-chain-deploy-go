package generate

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-deploy-go/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockDelegatedStakingGeneratorArguments() ArgDelegatedStakingGenerator {
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
	arg.WalletPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, logger.GetOrCreate("test"))
	arg.ValidatorPubKeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(96)
	arg.TotalSupply = big.NewInt(0)
	arg.TotalSupply.SetString("20000000000000000000000000", 10)
	arg.NodePrice = big.NewInt(0)
	arg.NodePrice.SetString("2500000000000000000000", 10)

	return arg
}

func TestDelegatedStakingGenerator_GenerateShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockDelegatedStakingGeneratorArguments()
	arg.NumValidatorBlsKeys = 33
	arg.NumObserverBlsKeys = 3
	arg.NumDelegators = 47
	arg.NumAdditionalWalletKeys = 3

	dsg, err := NewDelegatedGenerator(arg)
	require.Nil(t, err)

	generatedOutput, err := dsg.Generate()
	require.Nil(t, err)

	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.ValidatorBlsKeys))
	assert.Equal(t, int(arg.NumObserverBlsKeys), len(generatedOutput.ObserverBlsKeys))
	expectedNumInitialAccounts := arg.NumValidatorBlsKeys + arg.NumObserverBlsKeys + arg.NumDelegators
	assert.Equal(t, int(expectedNumInitialAccounts), len(generatedOutput.InitialAccounts))
	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.WalletKeys))
	assert.Equal(t, int(arg.NumAdditionalWalletKeys), len(generatedOutput.AdditionalKeys))
	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.InitialNodes))
	assert.Equal(t, int(arg.NumDelegators), len(generatedOutput.DelegatorKeys))

	iac, _ := check.NewInitialAccountsChecker(arg.NodePrice, arg.TotalSupply)
	assert.Nil(t, err, iac.CheckInitialAccounts(generatedOutput.InitialAccounts))
}

func TestDelegatedStakingGenerator_GenerateWithRichestAccountShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockDelegatedStakingGeneratorArguments()
	arg.NumValidatorBlsKeys = 33
	arg.NumObserverBlsKeys = 3
	arg.NumDelegators = 47
	arg.NumAdditionalWalletKeys = 3
	arg.RichestAccountMode = true

	dsg, err := NewDelegatedGenerator(arg)
	require.Nil(t, err)

	generatedOutput, err := dsg.Generate()
	require.Nil(t, err)

	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.ValidatorBlsKeys))
	assert.Equal(t, int(arg.NumObserverBlsKeys), len(generatedOutput.ObserverBlsKeys))
	expectedNumInitialAccounts := arg.NumValidatorBlsKeys + arg.NumObserverBlsKeys + arg.NumDelegators
	assert.Equal(t, int(expectedNumInitialAccounts), len(generatedOutput.InitialAccounts))
	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.WalletKeys))
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
