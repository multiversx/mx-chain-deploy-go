package generate

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-deploy-go/check"
	"github.com/multiversx/mx-chain-deploy-go/mock"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockDirectStakingGeneratorArguments() ArgDirectStakingGenerator {
	mclSuite := mcl.NewSuiteBLS12()
	edSuite := ed25519.NewEd25519()

	arg := ArgDirectStakingGenerator{
		KeyGeneratorForValidators: signing.NewKeyGenerator(mclSuite),
		KeyGeneratorForWallets:    signing.NewKeyGenerator(edSuite),
		NumValidatorBlsKeys:       0,
		NumObserverBlsKeys:        0,
		RichestAccountMode:        false,
		MaxNumNodesOnOwner:        0,
		NumAdditionalWalletKeys:   0,
		IntRandomizer:             &mock.IntRandomizerStub{},
		NodePrice:                 big.NewInt(2500),
		TotalSupply:               big.NewInt(20000000),
		InitialRating:             50,
	}
	arg.WalletPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, logger.GetOrCreate("test"))
	arg.ValidatorPubKeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(96)

	return arg
}

func TestDirectStakingGenerator_GenerateShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockDirectStakingGeneratorArguments()
	arg.NumValidatorBlsKeys = 33
	arg.NumObserverBlsKeys = 3
	arg.MaxNumNodesOnOwner = 1
	arg.NumAdditionalWalletKeys = 3

	dsg, err := NewDirectStakingGenerator(arg)
	require.Nil(t, err)

	generatedOutput, err := dsg.Generate()
	require.Nil(t, err)

	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.ValidatorBlsKeys))
	assert.Equal(t, int(arg.NumObserverBlsKeys), len(generatedOutput.ObserverBlsKeys))
	expectedNumInitialAccounts := arg.NumValidatorBlsKeys + arg.NumAdditionalWalletKeys
	assert.Equal(t, int(expectedNumInitialAccounts), len(generatedOutput.InitialAccounts))
	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.WalletKeys))
	assert.Equal(t, int(arg.NumAdditionalWalletKeys), len(generatedOutput.AdditionalKeys))
	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.InitialNodes))
	assert.Equal(t, 0, len(generatedOutput.DelegatorKeys))

	iac, _ := check.NewInitialAccountsChecker(arg.NodePrice, arg.TotalSupply)
	assert.Nil(t, err, iac.CheckInitialAccounts(generatedOutput.InitialAccounts))
}

func TestDirectStakingGenerator_GenerateWithRichestAccountShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockDirectStakingGeneratorArguments()
	arg.TotalSupply = big.NewInt(0)
	arg.TotalSupply.SetString("20000000000000000000000000", 10)
	arg.NodePrice = big.NewInt(0)
	arg.NodePrice.SetString("2500000000000000000000", 10)
	arg.NumValidatorBlsKeys = 33
	arg.NumObserverBlsKeys = 3
	arg.MaxNumNodesOnOwner = 1
	arg.NumAdditionalWalletKeys = 3
	arg.RichestAccountMode = true

	dsg, err := NewDirectStakingGenerator(arg)
	require.Nil(t, err)

	generatedOutput, err := dsg.Generate()
	require.Nil(t, err)

	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.ValidatorBlsKeys))
	assert.Equal(t, int(arg.NumObserverBlsKeys), len(generatedOutput.ObserverBlsKeys))
	expectedNumInitialAccounts := arg.NumValidatorBlsKeys + arg.NumAdditionalWalletKeys
	assert.Equal(t, int(expectedNumInitialAccounts), len(generatedOutput.InitialAccounts))
	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.WalletKeys))
	assert.Equal(t, int(arg.NumAdditionalWalletKeys), len(generatedOutput.AdditionalKeys))
	assert.Equal(t, int(arg.NumValidatorBlsKeys), len(generatedOutput.InitialNodes))
	assert.Equal(t, 0, len(generatedOutput.DelegatorKeys))

	iac, _ := check.NewInitialAccountsChecker(arg.NodePrice, arg.TotalSupply)
	assert.Nil(t, err, iac.CheckInitialAccounts(generatedOutput.InitialAccounts))
	for i, ia := range generatedOutput.InitialAccounts {
		if i == 0 {
			assert.NotEqual(t, minimumInitialBalance, ia.Balance)
		} else {
			assert.Equal(t, minimumInitialBalance, ia.Balance)
		}
	}
}
