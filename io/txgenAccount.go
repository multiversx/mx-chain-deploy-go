package io

import "math/big"

type txgenAccount struct {
	PubKey        string   `json:"pubKey"`
	PrivKey       string   `json:"privKey"`
	LastNonce     uint64   `json:"lastNonce"`
	Balance       *big.Int `json:"balance"`
	TokenBalance  *big.Int `json:"tokenBalance"`
	CanReuseNonce bool     `json:"canReuseNonce"`
}
