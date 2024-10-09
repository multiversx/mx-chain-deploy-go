package factory

import (
	"fmt"
	"math/big"

	mxCore "github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-deploy-go/core"
	"github.com/multiversx/mx-chain-deploy-go/generate"
)

// ArgDataGenerator is the argument used by the data generator method factory
type ArgDataGenerator struct {
	KeyGeneratorForValidators  crypto.KeyGenerator
	KeyGeneratorForWallets     crypto.KeyGenerator
	WalletPubKeyConverter      mxCore.PubkeyConverter
	ValidatorPubKeyConverter   mxCore.PubkeyConverter
	NumValidatorBlsKeys        uint
	NumObserverBlsKeys         uint
	RichestAccountMode         bool
	MaxNumNodesOnOwner         uint
	NumAdditionalWalletKeys    uint
	IntRandomizer              generate.IntRandomizer
	NodePrice                  *big.Int
	TotalSupply                *big.Int
	InitialRating              uint64
	GenerationType             string
	DelegationOwnerPkString    string
	DelegationOwnerNonce       uint64
	VmType                     string
	NumDelegators              uint
	NumDelegatedNodes          uint
	NumShards                  uint32
	GenerateWalletsInAllShards bool
}

// CreateDataGenerator will attempt to create a data generator instance
func CreateDataGenerator(arg ArgDataGenerator) (DataGenerator, error) {
	switch arg.GenerationType {
	case core.StakedType:
		return stakedTypeDataGenerator(arg)
	case core.DelegatedStakeType:
		return delegatedTypeDataGenerator(arg)
	case core.MixedType:
		return mixedTypeDataGenerator(arg)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownGenerationType, arg.GenerationType)
	}
}

func stakedTypeDataGenerator(arg ArgDataGenerator) (DataGenerator, error) {
	argDirectStaking := generate.ArgDirectStakingGenerator{
		KeyGeneratorForValidators:  arg.KeyGeneratorForValidators,
		KeyGeneratorForWallets:     arg.KeyGeneratorForWallets,
		WalletPubKeyConverter:      arg.WalletPubKeyConverter,
		ValidatorPubKeyConverter:   arg.ValidatorPubKeyConverter,
		NumValidatorBlsKeys:        arg.NumValidatorBlsKeys,
		NumObserverBlsKeys:         arg.NumObserverBlsKeys,
		RichestAccountMode:         arg.RichestAccountMode,
		MaxNumNodesOnOwner:         arg.MaxNumNodesOnOwner,
		NumAdditionalWalletKeys:    arg.NumAdditionalWalletKeys,
		IntRandomizer:              arg.IntRandomizer,
		NodePrice:                  arg.NodePrice,
		TotalSupply:                arg.TotalSupply,
		InitialRating:              arg.InitialRating,
		NumShards:                  arg.NumShards,
		GenerateWalletsInAllShards: arg.GenerateWalletsInAllShards,
	}

	return generate.NewDirectStakingGenerator(argDirectStaking)
}

func delegatedTypeDataGenerator(arg ArgDataGenerator) (DataGenerator, error) {
	argDelegatedStaking := generate.ArgDelegatedStakingGenerator{
		KeyGeneratorForValidators:  arg.KeyGeneratorForValidators,
		KeyGeneratorForWallets:     arg.KeyGeneratorForWallets,
		WalletPubKeyConverter:      arg.WalletPubKeyConverter,
		ValidatorPubKeyConverter:   arg.ValidatorPubKeyConverter,
		NumValidatorBlsKeys:        arg.NumValidatorBlsKeys,
		NumObserverBlsKeys:         arg.NumObserverBlsKeys,
		RichestAccountMode:         arg.RichestAccountMode,
		NumAdditionalWalletKeys:    arg.NumAdditionalWalletKeys,
		NodePrice:                  arg.NodePrice,
		TotalSupply:                arg.TotalSupply,
		InitialRating:              arg.InitialRating,
		DelegationOwnerPkString:    arg.DelegationOwnerPkString,
		DelegationOwnerNonce:       arg.DelegationOwnerNonce,
		VmType:                     arg.VmType,
		NumDelegators:              arg.NumDelegators,
		NumShards:                  arg.NumShards,
		GenerateWalletsInAllShards: arg.GenerateWalletsInAllShards,
	}

	return generate.NewDelegatedGenerator(argDelegatedStaking)
}

func mixedTypeDataGenerator(arg ArgDataGenerator) (DataGenerator, error) {
	argDelegatedStaking := generate.ArgDelegatedStakingGenerator{
		KeyGeneratorForValidators:  arg.KeyGeneratorForValidators,
		KeyGeneratorForWallets:     arg.KeyGeneratorForWallets,
		WalletPubKeyConverter:      arg.WalletPubKeyConverter,
		ValidatorPubKeyConverter:   arg.ValidatorPubKeyConverter,
		NumValidatorBlsKeys:        arg.NumValidatorBlsKeys,
		NumObserverBlsKeys:         arg.NumObserverBlsKeys,
		RichestAccountMode:         arg.RichestAccountMode,
		NumAdditionalWalletKeys:    arg.NumAdditionalWalletKeys,
		NodePrice:                  arg.NodePrice,
		TotalSupply:                arg.TotalSupply,
		InitialRating:              arg.InitialRating,
		DelegationOwnerPkString:    arg.DelegationOwnerPkString,
		DelegationOwnerNonce:       arg.DelegationOwnerNonce,
		VmType:                     arg.VmType,
		NumDelegators:              arg.NumDelegators,
		NumShards:                  arg.NumShards,
		GenerateWalletsInAllShards: arg.GenerateWalletsInAllShards,
	}

	argMixedStaking := generate.ArgMixedStakingGenerator{
		ArgDelegatedStakingGenerator: argDelegatedStaking,
		NumDelegatedNodes:            arg.NumDelegatedNodes,
		MaxNumNodesOnOwner:           arg.MaxNumNodesOnOwner,
		IntRandomizer:                arg.IntRandomizer,
	}

	return generate.NewMixedStakingGenerator(argMixedStaking)
}
