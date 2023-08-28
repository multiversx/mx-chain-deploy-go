package data

import (
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/sharding"
)

// GeneratorOutput represents the structure that will contain aggregated generated data
type GeneratorOutput struct {
	ValidatorBlsKeys []*BlsKey
	ObserverBlsKeys  []*BlsKey
	WalletKeys       []*WalletKey
	AdditionalKeys   []*WalletKey
	InitialAccounts  []data.InitialAccount
	InitialNodes     []*sharding.InitialNode
	DelegatorKeys    []*WalletKey
}
