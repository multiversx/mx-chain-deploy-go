package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
	txSignKeyFormat = cli.StringFlag{
		Name:  "tx-sign-key-format",
		Usage: "This flag specifies the format for transactions sign keys",
		Value: "bech32",
	}
	blockSignKeyFormat = cli.StringFlag{
		Name:  "block-sign-key-format",
		Usage: "This flag specifies the format for blocks sign keys",
		Value: "hex",
	}
	totalSupply = cli.StringFlag{
		Name:  "total-supply",
		Usage: "Total supply available",
		Value: "20000000000000000000000000000",
	}
	nodePrice = cli.StringFlag{
		Name:  "node-price",
		Usage: "The cost for a node",
		Value: "500000000000000000000000",
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
	numAdditionalAccountsInGenesis = cli.IntFlag{
		Name:  "num-aditional-accounts",
		Usage: "Number of additional accounts which will be added in genesis",
		Value: 0,
	}
	numOfMetachainObservers = cli.IntFlag{
		Name:  "num-of-observers-in-metachain",
		Usage: "Number of initial metachain observers, private/public keys, to generate",
		Value: 1,
	}
	hysteresis = cli.Float64Flag{
		Name: "hysteresis",
		Usage: "Hysteresis value - multiplied with numOfNodesPerShard to compute number of nodes allowed in the " +
			"waiting list of each shard",
		Value: 0.0,
	}
	adaptivity = cli.BoolFlag{
		Name:  "adaptivity",
		Usage: "Adaptivity value - should be set to true if shard merging and splitting is enabled",
	}
	chainID = cli.StringFlag{
		Name:  "chain-id",
		Usage: "Chain ID flag",
		Value: "testnet",
	}
	txgenFile = cli.BoolFlag{
		Name:  "txgen",
		Usage: "If set, will generate the accounts.json file needed for txgen",
	}

	walletKeyFileName     = "./walletKey.pem"
	validatorKeyFileName  = "./validatorKey.pem"
	genesisFilename       = "./genesis.json"
	nodesSetupFilename    = "./nodesSetup.json"
	txgenAccountsFileName = "./accounts.json"

	errInvalidNumPrivPubKeys = errors.New("invalid number of private/public keys to generate")
	errInvalidMintValue      = errors.New("invalid mint value for generated public keys")
	errInvalidNumOfNodes     = errors.New("invalid number of nodes in shard/metachain or in the consensus group")
	errCreatingKeygen        = errors.New("cannot create key gen")

	log = logger.GetOrCreate("main")
)

type txgenAccount struct {
	PubKey        string   `json:"pubKey"`
	PrivKey       string   `json:"privKey"`
	LastNonce     uint64   `json:"lastNonce"`
	Balance       *big.Int `json:"balance"`
	TokenBalance  *big.Int `json:"tokenBalance"`
	CanReuseNonce bool     `json:"canReuseNonce"`
}

// The resulting binary will be used to generate 2 files: genesis.json and privkeys.pem
// Those files are used to mass-deploy nodes and thus, ensuring that all nodes have the same data to work with
// The 2 optional flags are used to specify how many private/public keys to generate and the initial minting for each
// public generated key
func main() {
	_ = logger.SetLogLevel("*:DEBUG")
	app := cli.NewApp()
	cli.AppHelpTemplate = fileGenHelpTemplate
	app.Name = "Deploy Preparation Tool"
	app.Version = "v0.0.1"
	app.Usage = "This binary will generate a initialBalancesSk.pem, initialNodesSk.pem, genesis.json and nodesSetup.json" +
		" files, to be used in mass deployment"
	app.Flags = []cli.Flag{
		txSignKeyFormat,
		blockSignKeyFormat,
		totalSupply,
		nodePrice,
		numOfShards,
		numOfNodesPerShard,
		consensusGroupSize,
		numOfObserversPerShard,
		numOfMetachainNodes,
		metachainConsensusGroupSize,
		numOfMetachainObservers,
		numAdditionalAccountsInGenesis,
		hysteresis,
		adaptivity,
		chainID,
		txgenFile,
	}
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
		log.Error(err.Error())
		os.Exit(1)
	}
}

func getIdentifierAndPrivateKey(keyGen crypto.KeyGenerator, pubKeyConverter state.PubkeyConverter) (string, []byte, error) {
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
	pkString := pubKeyConverter.Encode(pkBytes)

	return pkString, skHex, nil
}

func generateFiles(ctx *cli.Context) error {
	var err error

	startTime := time.Now()
	txSignKeyFormatValue := ctx.GlobalString(txSignKeyFormat.Name)
	blockSignKeyFormatValue := ctx.GlobalString(blockSignKeyFormat.Name)
	numOfShards := ctx.GlobalInt(numOfShards.Name)
	numOfNodesPerShard := ctx.GlobalInt(numOfNodesPerShard.Name)
	consensusGroupSize := ctx.GlobalInt(consensusGroupSize.Name)
	numOfObserversPerShard := ctx.GlobalInt(numOfObserversPerShard.Name)
	numOfMetachainNodes := ctx.GlobalInt(numOfMetachainNodes.Name)
	metachainConsensusGroupSize := ctx.GlobalInt(metachainConsensusGroupSize.Name)
	numOfMetachainObservers := ctx.GlobalInt(numOfMetachainObservers.Name)
	numOfAdditionalAccounts := ctx.GlobalInt(numAdditionalAccountsInGenesis.Name)
	hysteresisValue := ctx.GlobalFloat64(hysteresis.Name)
	adaptivityValue := ctx.GlobalBool(adaptivity.Name)
	chainID := ctx.GlobalString(chainID.Name)
	generateTxgenFile := ctx.IsSet(txgenFile.Name)

	pubKeyConverterTxs, errPkC := factory.NewPubkeyConverter(config.PubkeyConfig{
		Length: 32, // TODO: use a constant after it is defined in elrond-go
		Type:   txSignKeyFormatValue,
	})
	if errPkC != nil {
		return errPkC
	}

	pubKeyConverterBlocks, errPkC := factory.NewPubkeyConverter(config.PubkeyConfig{
		Length: 96, // TODO: use a constant after it is defined in elrond-go
		Type:   blockSignKeyFormatValue,
	})
	if errPkC != nil {
		return errPkC
	}

	totalAddressesWithBalances := numOfShards*(numOfNodesPerShard+numOfObserversPerShard) +
		numOfMetachainNodes + numOfMetachainObservers

	invalidNumPrivPubKey := totalAddressesWithBalances < 1 ||
		numOfShards < 1 ||
		numOfNodesPerShard < 1 ||
		numOfMetachainNodes < 1
	if invalidNumPrivPubKey {
		return errInvalidNumPrivPubKeys
	}

	invalidNumOfNodes := consensusGroupSize < 1 ||
		consensusGroupSize > numOfNodesPerShard ||
		numOfObserversPerShard < 0 ||
		metachainConsensusGroupSize < 1 ||
		metachainConsensusGroupSize > numOfMetachainNodes ||
		numOfMetachainObservers < 0
	if invalidNumOfNodes {
		return errInvalidNumOfNodes
	}

	totalSupplyString := ctx.GlobalString(totalSupply.Name)
	totalSupplyValue, err := convertToBigInt(totalSupplyString)
	if err != nil {
		return err
	}

	nodePriceString := ctx.GlobalString(nodePrice.Name)
	nodePriceValue, err := convertToBigInt(nodePriceString)
	if err != nil {
		return err
	}

	var (
		walletKeyFile        *os.File
		validatorKeyFile     *os.File
		genesisFile          *os.File
		nodesFile            *os.File
		txgenAccountsFile    *os.File
		pkString             string
		skHex                []byte
		suite                crypto.Suite
		balancesKeyGenerator crypto.KeyGenerator
	)

	defer func() {
		err = walletKeyFile.Close()
		log.LogIfError(err)
		err = validatorKeyFile.Close()
		log.LogIfError(err)
		err = genesisFile.Close()
		log.LogIfError(err)
		err = nodesFile.Close()
		log.LogIfError(err)
	}()

	err = os.Remove(walletKeyFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	walletKeyFile, err = os.OpenFile(walletKeyFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	err = os.Remove(validatorKeyFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	validatorKeyFile, err = os.OpenFile(validatorKeyFileName, os.O_CREATE|os.O_WRONLY, 0666)
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

	if generateTxgenFile {
		err = os.Remove(txgenAccountsFileName)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		txgenAccountsFile, err = os.OpenFile(txgenAccountsFileName, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
	}

	genesisList := make([]*genesis.InitialAccount, 0, totalAddressesWithBalances)

	var initialNodes []*sharding.InitialNode
	nodes := &sharding.NodesSetup{
		StartTime:                   0,
		RoundDuration:               4000,
		ConsensusGroupSize:          uint32(consensusGroupSize),
		MinNodesPerShard:            uint32(numOfNodesPerShard),
		MetaChainConsensusGroupSize: uint32(metachainConsensusGroupSize),
		MetaChainMinNodes:           uint32(numOfMetachainNodes),
		Hysteresis:                  float32(hysteresisValue),
		Adaptivity:                  adaptivityValue,
		ChainID:                     chainID,
		InitialNodes:                initialNodes,
	}

	txgenAccounts := make(map[uint32][]*txgenAccount)

	suite = ed25519.NewEd25519()
	balancesKeyGenerator = signing.NewKeyGenerator(suite)

	shardCoordinator, err := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	if err != nil {
		return err
	}

	numObservers := numOfShards*numOfObserversPerShard + numOfMetachainObservers
	numValidators := totalAddressesWithBalances - numObservers

	//initialTotalBalance = totalSupply - (numValidators * nodePriceValue)
	initialTotalBalance := big.NewInt(0).Set(totalSupplyValue)
	staked := big.NewInt(0).Mul(big.NewInt(int64(numValidators)), nodePriceValue)
	initialTotalBalance.Sub(initialTotalBalance, staked)

	// initialNodeBalance = initialTotalBalance / (totalAddressesWithBalances + numOfAdditionalAccounts)
	initialNodeBalance := big.NewInt(0).Set(initialTotalBalance)
	initialNodeBalance.Div(initialNodeBalance,
		big.NewInt(int64(totalAddressesWithBalances+numOfAdditionalAccounts)))
	log.Debug("supply values",
		"total supply", totalSupplyValue.String(),
		"staked", staked.String(),
		"initial total balance", initialTotalBalance.String(),
	)
	log.Debug("nodes",
		"num nodes with balance", totalAddressesWithBalances,
		"num additional accounts", numOfAdditionalAccounts,
		"num validators", numValidators,
		"num observers", numObservers,
		"initial node balance", initialNodeBalance.String(),
	)

	log.Info("started generating...")
	for i := 0; i < totalAddressesWithBalances; i++ {
		pkString, skHex, err = getIdentifierAndPrivateKey(balancesKeyGenerator, pubKeyConverterTxs)
		if err != nil {
			return err
		}

		supply := big.NewInt(0).Set(nodePriceValue)
		supply.Add(supply, big.NewInt(0).Set(initialNodeBalance))
		ia := &genesis.InitialAccount{
			Address:      pkString,
			Supply:       supply,
			Balance:      big.NewInt(0).Set(initialNodeBalance),
			StakingValue: big.NewInt(0).Set(nodePriceValue),
			Delegation: &genesis.DelegationData{
				Address: "",
				Value:   big.NewInt(0),
			},
		}
		genesisList = append(genesisList, ia)

		err = core.SaveSkToPemFile(walletKeyFile, pkString, skHex)
		if err != nil {
			return err
		}

		keyGen := getNodesKeyGen()
		if keyGen == nil {
			return errCreatingKeygen
		}

		pkHexForNode, skHexForNode, err := getIdentifierAndPrivateKey(keyGen, pubKeyConverterBlocks)
		if err != nil {
			return err
		}

		err = core.SaveSkToPemFile(validatorKeyFile, pkHexForNode, skHexForNode)
		if err != nil {
			return err
		}

		if i < numValidators {
			nodes.InitialNodes = append(nodes.InitialNodes, &sharding.InitialNode{
				PubKey:  pkHexForNode,
				Address: pkString,
			})
		} else {
			ia.StakingValue = big.NewInt(0)
			ia.Supply = big.NewInt(0).Set(ia.Balance)
		}
	}

	for i := 0; i < numOfAdditionalAccounts; i++ {
		pkString, skHex, err = getIdentifierAndPrivateKey(balancesKeyGenerator, pubKeyConverterTxs)
		if err != nil {
			return err
		}

		ia := &genesis.InitialAccount{
			Address:      pkString,
			Supply:       big.NewInt(0).Set(initialNodeBalance),
			Balance:      big.NewInt(0).Set(initialNodeBalance),
			StakingValue: big.NewInt(0),
			Delegation: &genesis.DelegationData{
				Address: "",
				Value:   big.NewInt(0),
			},
		}
		genesisList = append(genesisList, ia)

		if generateTxgenFile {
			pkBytes, _ := pubKeyConverterTxs.Decode(pkString)
			address := state.NewAddress(pkBytes)
			sId := shardCoordinator.ComputeId(address)
			txgenAccounts[sId] = append(txgenAccounts[sId], &txgenAccount{
				PubKey:        pkString,
				PrivKey:       string(skHex),
				LastNonce:     0,
				Balance:       big.NewInt(0).Set(initialNodeBalance),
				TokenBalance:  big.NewInt(0),
				CanReuseNonce: true,
			})
		}
	}

	//take the remainder and set it on the first node
	// initialBalance = initialTotalBalance - (totalAddressesWithBalances + numOfAdditionalAccounts - 1) * initialNodeBalance
	iaFirst := genesisList[0]
	initialBalance := big.NewInt(0).Set(initialNodeBalance)
	subs := big.NewInt(int64(totalAddressesWithBalances+numOfAdditionalAccounts) - 1)
	subs.Mul(subs, big.NewInt(0).Set(initialNodeBalance))
	initialBalance = big.NewInt(0).Set(initialTotalBalance)
	initialBalance.Sub(initialBalance, subs)

	iaFirst.Balance = big.NewInt(0).Set(initialBalance)
	initialBalance.Add(initialBalance, iaFirst.StakingValue)
	iaFirst.Supply = initialBalance

	genesisBuff, err := json.MarshalIndent(genesisList, "", "  ")
	if err != nil {
		return err
	}

	_, err = genesisFile.Write(genesisBuff)
	if err != nil {
		return err
	}

	nodesBuff, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	_, err = nodesFile.Write(nodesBuff)
	if err != nil {
		return err
	}

	if generateTxgenFile {
		txgenAccountsBuff, err := json.MarshalIndent(txgenAccounts, "", "  ")
		if err != nil {
			return err
		}

		_, err = txgenAccountsFile.Write(txgenAccountsBuff)
		if err != nil {
			return err
		}
	}

	log.Info("elapsed time in seconds", time.Since(startTime).Seconds())
	log.Info("files generated successfully!")
	return nil
}

func convertToBigInt(value string) (*big.Int, error) {
	valueNumber, isNumber := big.NewInt(0).SetString(value, 10)
	if !isNumber {
		return nil, errInvalidMintValue
	}

	if valueNumber.Cmp(big.NewInt(0)) < 0 {
		return nil, errInvalidMintValue
	}

	return valueNumber, nil
}

func getNodesKeyGen() crypto.KeyGenerator {
	suite := mcl.NewSuiteBLS12()

	return signing.NewKeyGenerator(suite)
}
