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
	numPairsWithBalances = cli.IntFlag{
		Name:  "num-pairs-with-balances",
		Usage: "Number of private/public keys with balances to generate",
		Value: 3,
	}
	mintValue = cli.Uint64Flag{
		Name:  "mint-value",
		Usage: "Initial minting for all public keys generated",
		Value: 1000000000,
	}
	numPairsEligible = cli.IntFlag{
		Name:  "num-pairs-eligible",
		Usage: "Number of private/public keys eligible to generate",
		Value: 21,
	}
	consensusBls = cli.BoolFlag{
		Name:  "consensus-bls",
		Usage: "Eligible nodes will be generated with bls private/public keys",
	}

	balancesPrivKeysFileName = "./balancesPrivateKeys.pem"
	eligiblePrivKeysFileName = "./eligiblePrivateKeys.pem"
	genesisFilename  = "./genesis.json"
	nodesFilename  = "./nodes.json"

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
	app.Usage = "This binary will generate a balancesPrivateKeys.pem, eligiblePrivateKeys.pem, genesis.json and nodes.json" +
		" files, to be used in mass deployment"
	app.Flags = []cli.Flag{numPairsWithBalances, mintValue, numPairsEligible, consensusBls}
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

func getIdentifierAndPrivateKey(blsGenerator crypto.KeyGenerator) (string, []byte, error) {
	blsSk, blsPk := blsGenerator.GeneratePair()
	blsSkBytes, err := blsSk.ToByteArray()
	if err != nil {
		return "", nil, err
	}

	blsPkBytes, err := blsPk.ToByteArray()
	if err != nil {
		return "", nil, err
	}

	blsSkHex := []byte(hex.EncodeToString(blsSkBytes))
	blsPkHex := hex.EncodeToString(blsPkBytes)
	return blsPkHex, blsSkHex, nil
}

func generateFiles(ctx *cli.Context) error {
	numPairsWithBalances := ctx.GlobalInt(numPairsWithBalances.Name)
	numPairsEligible := ctx.GlobalInt(numPairsEligible.Name)
	if numPairsWithBalances < 1 || numPairsEligible < 1 {
		return errInvalidNumPrivPubKeys
	}

	initialMint := ctx.GlobalUint64(mintValue.Name)
	if initialMint < 0 {
		return errInvalidMintValue
	}

	consensusBls := ctx.GlobalBool(consensusBls.Name)

	var err error
	var balancesPrivKeysFile, eligiblePrivKeysFile, genesisFile, nodesFile *os.File
	var pkHex string
	var skHex []byte

	defer func() {
		err = balancesPrivKeysFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = eligiblePrivKeysFile.Close()
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

	balancesPrivKeysFile, err = os.OpenFile(balancesPrivKeysFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	eligiblePrivKeysFile, err = os.OpenFile(eligiblePrivKeysFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	genesisFile, err = os.OpenFile(genesisFilename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	nodesFile, err = os.OpenFile(nodesFilename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	genesis := &sharding.Genesis{
		InitialBalances: make([]*sharding.InitialBalance, numPairsWithBalances),
	}

	nodes := &sharding.Nodes{
		StartTime:                   0,
		RoundDuration:               6000,
		ConsensusGroupSize:          uint32(numPairsEligible - 1),
		MinNodesPerShard:            uint32(numPairsEligible - 1),
		InitialNodes:                make([]*sharding.InitialNode, numPairsEligible),
		MetaChainActive:             true,
		MetaChainConsensusGroupSize: 1,
		MetaChainMinNodes:           1,
	}

	suite := kyber.NewBlakeSHA256Ed25519()
	generator := signing.NewKeyGenerator(suite)

	blsSuite := kyber.NewSuitePairingBn256()
	blsGenerator := signing.NewKeyGenerator(blsSuite)

	for i := 0; i < numPairsWithBalances; i++ {
		pkHex, skHex, err = getIdentifierAndPrivateKey(generator)
		if err != nil {
			return err
		}

		genesis.InitialBalances[i] = &sharding.InitialBalance{
			PubKey:  pkHex,
			Balance: fmt.Sprintf("%d", initialMint),
		}

		err = core.SaveSkToPemFile(balancesPrivKeysFile, pkHex, skHex)
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

	for i := 0; i < numPairsEligible; i++ {
		if consensusBls {
			pkHex, skHex, err = getIdentifierAndPrivateKey(blsGenerator)
		} else {
			pkHex, skHex, err = getIdentifierAndPrivateKey(generator)
		}

		if err != nil {
			return err
		}

		nodes.InitialNodes[i] = &sharding.InitialNode{
			PubKey: pkHex,
		}

		err = core.SaveSkToPemFile(eligiblePrivKeysFile, pkHex, skHex)
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
