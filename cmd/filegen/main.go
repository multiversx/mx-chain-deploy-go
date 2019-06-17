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
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
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
		Name:  "num-of-addresses-with-balances",
		Usage: "Number of addresses, private/public keys, with balances to generate",
		Value: 132,
	}
	mintValue = cli.Uint64Flag{
		Name:  "mint-value",
		Usage: "Initial minting for all public keys generated",
		Value: 1000000000,
	}
	numOfShards = cli.IntFlag{
		Name:  "num-of-shards",
		Usage: "Number of initial shards",
		Value: 5,
	}
	numOfNodesPerShard = cli.IntFlag{
		Name:  "num-of-nodes-in-each-shard",
		Usage: "Number of initial nodes in each shard, private/public keys, to generate",
		Value: 21,
	}
	consensusGroupSize = cli.IntFlag{
		Name:  "consensus-group-size",
		Usage: "Consensus group size",
		Value: 15,
	}
	numOfObserversPerShard = cli.IntFlag{
		Name:  "num-of-observers-in-each-shard",
		Usage: "Number of initial observers in each shard, private/public keys, to generate",
		Value: 1,
	}
	numOfMetachainNodes = cli.IntFlag{
		Name:  "num-of-metachain-nodes",
		Usage: "Number of initial metachain nodes, private/public keys, to generate",
		Value: 21,
	}
	metachainConsensusGroupSize = cli.IntFlag{
		Name:  "metachain-consensus-group-size",
		Usage: "Metachain consensus group size",
		Value: 15,
	}
	numOfMetachainObservers = cli.IntFlag{
		Name:  "num-of-observers-in-metachain",
		Usage: "Number of initial metachain observers, private/public keys, to generate",
		Value: 1,
	}
	consensusType = cli.StringFlag{
		Name:  "consensus-type",
		Usage: "Consensus type to be used and for which, private/public keys, to generate",
		Value: "bls",
	}

	initialBalancesSkFileName      = "./initialBalancesSk.pem"
	initialBalancesSkPlainFileName = "./initialBalancesSkPlain.txt"
	initialBalancesPkPlainFileName = "./initialBalancesPkPlain.txt"
	initialNodesSkFileName         = "./initialNodesSk.pem"
	initialNodesSkPlainFileName    = "./initialNodesSkPlain.txt"
	initialNodesPkPlainFileName    = "./initialNodesPkPlain.txt"
	genesisFilename                = "./genesis.json"
	nodesSetupFilename             = "./nodesSetup.json"

	errInvalidNumPrivPubKeys = errors.New("invalid number of private/public keys to generate")
	errInvalidMintValue      = errors.New("invalid mint value for generated public keys")
	errInvalidNumberOfNodes  = errors.New("invalid number of nodes in shard/metachain or in the consensus group")
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
	app.Flags = []cli.Flag{
		numAddressesWithBalances,
		mintValue,
		numOfShards,
		numOfNodesPerShard,
		consensusGroupSize,
		numOfObserversPerShard,
		numOfMetachainNodes,
		metachainConsensusGroupSize,
		numOfMetachainObservers,
		consensusType}
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
	numOfShards := ctx.GlobalInt(numOfShards.Name)
	numOfNodesPerShard := ctx.GlobalInt(numOfNodesPerShard.Name)
	consensusGroupSize := ctx.GlobalInt(consensusGroupSize.Name)
	numOfObserversPerShard := ctx.GlobalInt(numOfObserversPerShard.Name)
	numOfMetachainNodes := ctx.GlobalInt(numOfMetachainNodes.Name)
	metachainConsensusGroupSize := ctx.GlobalInt(metachainConsensusGroupSize.Name)
	numOfMetachainObservers := ctx.GlobalInt(numOfMetachainObservers.Name)

	var totalAddressesWithBalances int
	if ctx.GlobalIsSet(numAddressesWithBalances.Name) {
		totalAddressesWithBalances = ctx.GlobalInt(numAddressesWithBalances.Name)
	} else {
		totalAddressesWithBalances = numOfShards*(numOfNodesPerShard+numOfObserversPerShard) + numOfMetachainNodes + numOfMetachainObservers
	}

	totalNumOfNodes := numOfShards*numOfNodesPerShard + numOfMetachainNodes

	if totalAddressesWithBalances < 1 ||
		numOfShards < 1 ||
		numOfNodesPerShard < 1 ||
		numOfMetachainNodes < 1 {
		return errInvalidNumPrivPubKeys
	}

	if consensusGroupSize < 1 ||
		consensusGroupSize > numOfNodesPerShard ||
		numOfObserversPerShard < 0 ||
		metachainConsensusGroupSize < 1 ||
		metachainConsensusGroupSize > numOfMetachainNodes ||
		numOfMetachainObservers < 0 {
		return errInvalidNumberOfNodes
	}

	initialMint := ctx.GlobalUint64(mintValue.Name)
	if initialMint < 0 {
		return errInvalidMintValue
	}

	consensusType := ctx.GlobalString(consensusType.Name)

	var err error
	var initialBalancesSkFile,
		initialBalancesSkPlainFile,
		initialBalancesPkPlainFile,
		initialNodesSkFile,
		initialNodesSkPlainFile,
		initialNodesPkPlainFile,
		genesisFile,
		nodesFile *os.File
	var pkHex string
	var skHex []byte
	var suite crypto.Suite
	var generator crypto.KeyGenerator

	defer func() {
		err = initialBalancesSkFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = initialBalancesSkPlainFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = initialBalancesPkPlainFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = initialNodesSkFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = initialNodesSkPlainFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = initialNodesPkPlainFile.Close()
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

	err = os.Remove(initialBalancesSkFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	initialBalancesSkFile, err = os.OpenFile(initialBalancesSkFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	err = os.Remove(initialBalancesSkPlainFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	initialBalancesSkPlainFile, err = os.OpenFile(initialBalancesSkPlainFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	err = os.Remove(initialBalancesPkPlainFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	initialBalancesPkPlainFile, err = os.OpenFile(initialBalancesPkPlainFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	err = os.Remove(initialNodesSkFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	initialNodesSkFile, err = os.OpenFile(initialNodesSkFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	err = os.Remove(initialNodesSkPlainFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	initialNodesSkPlainFile, err = os.OpenFile(initialNodesSkPlainFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	err = os.Remove(initialNodesPkPlainFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	initialNodesPkPlainFile, err = os.OpenFile(initialNodesPkPlainFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	err = os.Remove(genesisFilename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	genesisFile, err = os.OpenFile(genesisFilename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	err = os.Remove(nodesSetupFilename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	nodesFile, err = os.OpenFile(nodesSetupFilename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	genesis := &sharding.Genesis{
		InitialBalances: make([]*sharding.InitialBalance, totalAddressesWithBalances),
	}

	nodes := &sharding.NodesSetup{
		StartTime:                   0,
		RoundDuration:               4000,
		ConsensusGroupSize:          uint32(consensusGroupSize),
		MinNodesPerShard:            uint32(numOfNodesPerShard),
		MetaChainActive:             true,
		MetaChainConsensusGroupSize: uint32(metachainConsensusGroupSize),
		MetaChainMinNodes:           uint32(numOfMetachainNodes),
		InitialNodes:                make([]*sharding.InitialNode, totalNumOfNodes),
	}

	suite = kyber.NewBlakeSHA256Ed25519()
	generator = signing.NewKeyGenerator(suite)

	shardsObserversStartIndex := totalAddressesWithBalances - numOfShards*numOfObserversPerShard

	shardCoordinator, err := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	if err != nil {
		return err
	}

	for i := 0; i < totalAddressesWithBalances; i++ {
		pkHex, skHex, err = getIdentifierAndPrivateKey(generator)
		if err != nil {
			return err
		}

		if i >= shardsObserversStartIndex {
			shardId := uint32((i - shardsObserversStartIndex) / numOfObserversPerShard)
			pk, _ := hex.DecodeString(pkHex)
			for shardCoordinator.ComputeId(state.NewAddress(pk)) != shardId {
				pkHex, skHex, err = getIdentifierAndPrivateKey(generator)
				if err != nil {
					return err
				}

				pk, _ = hex.DecodeString(pkHex)
			}
		}

		genesis.InitialBalances[i] = &sharding.InitialBalance{
			PubKey:  pkHex,
			Balance: fmt.Sprintf("%d", initialMint),
		}

		err = core.SaveSkToPemFile(initialBalancesSkFile, pkHex, skHex)
		if err != nil {
			return err
		}

		_, err = initialBalancesSkPlainFile.Write(append(skHex, '\n'))
		if err != nil {
			return err
		}

		_, err = initialBalancesPkPlainFile.Write(append([]byte(pkHex), '\n'))
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

	// TODO: A factory which returns the suite according to consensus type should be created in elrond-go-sandbox project
	// Ex: crypto.NewSuite(consensusType) crypto.Suite
	switch consensusType {
	case "bls":
		suite = kyber.NewSuitePairingBn256()
	case "bn":
		suite = kyber.NewBlakeSHA256Ed25519()
	default:
		suite = nil
	}

	generator = signing.NewKeyGenerator(suite)

	for i := 0; i < totalNumOfNodes; i++ {
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

		_, err = initialNodesSkPlainFile.Write(append(skHex, '\n'))
		if err != nil {
			return err
		}

		_, err = initialNodesPkPlainFile.Write(append([]byte(pkHex), '\n'))
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
