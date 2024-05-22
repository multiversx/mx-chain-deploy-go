package generate

import (
	mxCommonFactory "github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/config"
	mxData "github.com/multiversx/mx-chain-go/genesis/data"
)

func adjustInitialAccounts(initialAccounts []mxData.InitialAccount) {
	// Workaround. Does not care about private keys etc.
	// Filegen should be ran with: -num-aditional-accounts=4

	// Sponsor
	replaceAccount(initialAccounts, len(initialAccounts)-1, "erd1guzgwrg6mwvmftx4ppdg8dv2s239d6pt2crdxklcmz9ngq8zztxsl9ah7z")

	// Controller
	replaceAccount(initialAccounts, len(initialAccounts)-2, "erd10nht7cm9tqyq8r6aqdx2ec46ak94q8xjxdzwcngffvlzmaju97tszdca2y")

	// Web address
	replaceAccount(initialAccounts, len(initialAccounts)-3, "erd1wh9c0sjr2xn8hzf02lwwcr4jk2s84tat9ud2kaq6zr7xzpvl9l5q8awmex")

	// Faucet address
	replaceAccount(initialAccounts, len(initialAccounts)-4, "erd1sjs26q5pmngu7qjnpkcqgstdppvqul7vdqa0cru5ae5axkq37czqndz3vp")
}

func replaceAccount(initialAccounts []mxData.InitialAccount, index int, address string) {
	converter, err := mxCommonFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length: 32,
		Type:   "bech32",
	})
	if err != nil {
		panic(err)
	}

	initialAccounts[index].Address = address
	addressBytes, _ := converter.Decode(address)
	initialAccounts[index].SetAddressBytes(addressBytes)
}
