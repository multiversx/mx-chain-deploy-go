package data

import (
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
