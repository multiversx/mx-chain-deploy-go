package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"os"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/urfave/cli"
)

var (
	fileGenHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`
	numAddressesWithBalances = cli.IntFlag{
		Name:  "num-addresses-with-balances",
		Usage: "Number of addresses, private/public keys, with balances to generate",
		Value: 21,
	}
	mintValue = cli.Uint64Flag{
		Name:  "mint-value",
		Usage: "Initial minting for all public keys generated",
		Value: 1000000000,
	}
	numNodes = cli.IntFlag{
		Name:  "num-nodes",
		Usage: "Number of initial nodes, private/public keys, to generate",
		Value: 21,
	}
	consensusType = cli.StringFlag{
		Name:  "consensus-type",
		Usage: "Consensus type to be used and for which, private/public keys, to generate",
		Value: "bls",
	}

	initialBalancesSkFileName = "./initialBalancesSk.pem"
	initialNodesSkFileName    = "./initialNodesSk.pem"
	genesisFilename           = "./genesis.json"
	nodesSetupFilename        = "./nodesSetup.json"

	errInvalidNumPrivPubKeys = errors.New("invalid number of private/public keys to generate")
	errInvalidMintValue      = errors.New("invalid mint value for generated public keys")
)

// The resulting binary will be used to generate 2 files: genesis.json and privkeys.pem
// Those files are used to mass-deploy nodes and thus, ensuring that all nodes have the same data to work with
// The 2 optional flags are used to specify how many private/public keys to generate and the initial minting for each
// public generated key
//TODO this should be refactor when genesis.json will hold only minting addresses
// and it will be a new config file for initial nodes (public keys list)
func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = fileGenHelpTemplate
	app.Name = "Deploy Preparation Tool"
	app.Version = "v0.0.1"
	app.Usage = "This binary will generate a initialBalancesSk.pem, initialNodesSk.pem, genesis.json and nodesSetup.json" +
		" files, to be used in mass deployment"
	app.Flags = []cli.Flag{numAddressesWithBalances, mintValue, numNodes, consensusType}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return generateFiles(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func getIdentifierAndPrivateKey(keyGen crypto.KeyGenerator) (string, []byte, error) {
	sk, pk := keyGen.GeneratePair()
	skBytes, err := sk.ToByteArray()
	if err != nil {
		return "", nil, err
	}

	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return "", nil, err
	}

	skHex := []byte(hex.EncodeToString(skBytes))
	pkHex := hex.EncodeToString(pkBytes)
	return pkHex, skHex, nil
}

func generateFiles(ctx *cli.Context) error {
	numAddressesWithBalances := ctx.GlobalInt(numAddressesWithBalances.Name)
	numNodes := ctx.GlobalInt(numNodes.Name)
	if numAddressesWithBalances < 1 || numNodes < 1 {
		return errInvalidNumPrivPubKeys
	}

	initialMint := ctx.GlobalUint64(mintValue.Name)
	if initialMint < 0 {
		return errInvalidMintValue
	}

	var err error
	var initialBalancesSkFile, initialNodesSkFile, genesisFile, nodesFile *os.File
	var pkHex string
	var skHex []byte
	var suite crypto.Suite
	var generator crypto.KeyGenerator

	defer func() {
		err = initialBalancesSkFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = initialNodesSkFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = genesisFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = nodesFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	os.Remove(initialBalancesSkFileName)
	initialBalancesSkFile, err = os.OpenFile(initialBalancesSkFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	os.Remove(initialNodesSkFileName)
	initialNodesSkFile, err = os.OpenFile(initialNodesSkFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	os.Remove(genesisFilename)
	genesisFile, err = os.OpenFile(genesisFilename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	os.Remove(nodesSetupFilename)
	nodesFile, err = os.OpenFile(nodesSetupFilename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	genesis := &sharding.Genesis{
		InitialBalances: make([]*sharding.InitialBalance, numAddressesWithBalances),
	}

	nodes := &sharding.NodesSetup{
		StartTime:                   0,
		RoundDuration:               6000,
		ConsensusGroupSize:          uint32(numNodes - 1),
		MinNodesPerShard:            uint32(numNodes - 1),
		InitialNodes:                make([]*sharding.InitialNode, numNodes),
		MetaChainActive:             true,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           1,
	}

	suite = kyber.NewBlakeSHA256Ed25519()
	generator = signing.NewKeyGenerator(suite)

	for i := 0; i < numAddressesWithBalances; i++ {
		pkHex, skHex, err = getIdentifierAndPrivateKey(generator)
		if err != nil {
			return err
		}

		genesis.InitialBalances[i] = &sharding.InitialBalance{
			PubKey:  pkHex,
			Balance: fmt.Sprintf("%d", initialMint),
		}

		err = core.SaveSkToPemFile(initialBalancesSkFile, pkHex, skHex)
		if err != nil {
			return err
		}
	}

	genesisBuff, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return err
	}

	_, err = genesisFile.Write(genesisBuff)
	if err != nil {
		return err
	}

	switch consensusType.Value {
	case "bls":
		suite = kyber.NewSuitePairingBn256()
	case "bn":
		suite = kyber.NewBlakeSHA256Ed25519()
	default:
		suite = nil
	}

	generator = signing.NewKeyGenerator(suite)

	for i := 0; i < numNodes; i++ {
		pkHex, skHex, err = getIdentifierAndPrivateKey(generator)

		if err != nil {
			return err
		}

		nodes.InitialNodes[i] = &sharding.InitialNode{
			PubKey: pkHex,
		}

		err = core.SaveSkToPemFile(initialNodesSkFile, pkHex, skHex)
		if err != nil {
			return err
		}
	}

	nodesBuff, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	_, err = nodesFile.Write(nodesBuff)
	if err != nil {
		return err
	}

	fmt.Println("Files generated successfully!")
	return nil
}
