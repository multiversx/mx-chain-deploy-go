package generate

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

// ArgDirectStakingGenerator is the argument used in direct staking mechanism
type ArgDirectStakingGenerator struct {
	KeyGeneratorForValidators crypto.KeyGenerator
	KeyGeneratorForWallets    crypto.KeyGenerator
	WalletPubKeyConverter     core.PubkeyConverter
	ValidatorPubKeyConverter  core.PubkeyConverter
	NumValidatorBlsKeys       uint
	NumObserverBlsKeys        uint
	RichestAccountMode        bool
	MaxNumNodesOnOwner        uint
	NumAdditionalWalletKeys   uint
	IntRandomizer             IntRandomizer
	NodePrice                 *big.Int
	TotalSupply               *big.Int
	InitialRating             uint64
}

// ArgDelegatedStakingGenerator is the argument used in delegated staking mechanism
type ArgDelegatedStakingGenerator struct {
	KeyGeneratorForValidators crypto.KeyGenerator
	KeyGeneratorForWallets    crypto.KeyGenerator
	WalletPubKeyConverter     core.PubkeyConverter
	ValidatorPubKeyConverter  core.PubkeyConverter
	NumValidatorBlsKeys       uint
	NumObserverBlsKeys        uint
	RichestAccountMode        bool
	NumAdditionalWalletKeys   uint
	NodePrice                 *big.Int
	TotalSupply               *big.Int
	InitialRating             uint64
	DelegationOwnerPkString   string
	DelegationOwnerNonce      uint64
	VmType                    string
	NumDelegators             uint
}

// ArgMixedStakingGenerator is the argument used in mixed staking mechanism
type ArgMixedStakingGenerator struct {
	ArgDelegatedStakingGenerator
	NumDelegatedNodes  uint
	MaxNumNodesOnOwner uint
	IntRandomizer      IntRandomizer
}
