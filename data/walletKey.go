package data

import "math/big"

// WalletKey will hold the data for a wallet key including bls key when it acts as an owner, balance and delegated info
type WalletKey struct {
	PubKeyBytes          []byte
	PrivKeyBytes         []byte
	BlsKeys              []*BlsKey
	Balance              *big.Int
	DelegatedValue       *big.Int
	DelegatedPubKeyBytes []byte
	StakedValue          *big.Int
}
