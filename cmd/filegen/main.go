package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/urfave/cli"
)

const delegatedStakeType = "delegated"
const stakedType = "direct"
const defaultRoundDuration = 5000

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
	outputDirectoryFlag = cli.StringFlag{
		Name:  "output-directory",
		Usage: "specifies the directory where all files will be saved",
		Value: "./output",
	}
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
		Value: "20000000000000000000000000",
	}
	nodePrice = cli.StringFlag{
		Name:  "node-price",
		Usage: "The cost for a node",
		Value: "2500000000000000000000",
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
	initialRating = cli.Uint64Flag{
		Name:  "initial-rating",
		Usage: "The initial rating to be used for each node",
		Value: 5000001,
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
	transactionVersion = cli.UintFlag{
		Name:  "tx-version",
		Usage: "Transaction Version flag flag",
		Value: 1,
	}
	txgenFile = cli.BoolFlag{
		Name:  "txgen",
		Usage: "If set, will generate the accounts.json file needed for txgen",
	}
	stakeType = cli.StringFlag{
		Name: "stake-type",
		Usage: "defines the 2 possible ways to stake the nodes: 'direct' as in direct staking " +
			"and 'delegated' that will stake the nodes through delegation",
		Value: "direct",
	}
	delegationOwnerPublicKey = cli.StringFlag{
		Name:  "delegation-owner-pk",
		Usage: "defines the delegation owner public key, encoded in bech32 format",
		Value: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
	}
	numDelegators = cli.UintFlag{
		Name:  "num-delegators",
		Usage: "number of delegators if the stake-type is of type `delegated`",
		Value: 100,
	}
	richestAccount = cli.BoolFlag{
		Name: "richest-account",
		Usage: "if this flag is set, all the remaining balance will be credited to a new account. " +
			"This flag is useful in tests involving automated stake events. All other account will still" +
			"receive 1eGLD in order to complete some transactions (unstake, for instance)",
	}

	walletKeyFileName     = "walletKey.pem"
	validatorKeyFileName  = "validatorKey.pem"
	genesisFilename       = "genesis.json"
	nodesSetupFilename    = "nodesSetup.json"
	txgenAccountsFileName = "accounts.json"
	delegatorsFileName    = "delegators.pem"
	vmType                = "0500"
	delegationOwnerNonce  = uint64(0)

	errInvalidNumPrivPubKeys = errors.New("invalid number of private/public keys to generate")
	errInvalidMintValue      = errors.New("invalid mint value for generated public keys")
	errInvalidNumOfNodes     = errors.New("invalid number of nodes in shard/metachain or in the consensus group")
	errCreatingKeygen        = errors.New("cannot create key gen")

	log                         = logger.GetOrCreate("main")
	zero                        = big.NewInt(0)
	initialBalanceForDelegators = big.NewInt(1000000000000000000) //1eGLD
	minimumInitialBalance       = big.NewInt(1000000000000000000) //1eGLD
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
	app := cli.NewApp()
	cli.AppHelpTemplate = fileGenHelpTemplate
	app.Name = "Deploy Preparation Tool"
	app.Version = "v0.0.1"
	app.Usage = "This binary will generate a initialBalancesSk.pem, initialNodesSk.pem, genesis.json and nodesSetup.json" +
		" files, to be used in mass deployment"
	app.Flags = []cli.Flag{
		outputDirectoryFlag,
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
		initialRating,
		hysteresis,
		adaptivity,
		chainID,
		transactionVersion,
		txgenFile,
		stakeType,
		delegationOwnerPublicKey,
		numDelegators,
		richestAccount,
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

func getIdentifierAndPrivateKey(keyGen crypto.KeyGenerator, pubKeyConverter core.PubkeyConverter) (string, []byte, error) {
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
	outputDirectory := ctx.GlobalString(outputDirectoryFlag.Name)
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
	initialRating := ctx.GlobalUint64(initialRating.Name)
	hysteresisValue := ctx.GlobalFloat64(hysteresis.Name)
	adaptivityValue := ctx.GlobalBool(adaptivity.Name)
	chainID := ctx.GlobalString(chainID.Name)
	txVersion := ctx.GlobalUint(transactionVersion.Name)
	generateTxgenFile := ctx.IsSet(txgenFile.Name)
	numDelegatorsValue := ctx.GlobalUint(numDelegators.Name)
	withRichestAccount := ctx.GlobalBool(richestAccount.Name)

	err = prepareOutputDirectory(outputDirectory)
	if err != nil {
		return err
	}

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

	numValidatorsOnAShard := int(math.Ceil(float64(numOfNodesPerShard) * (1 + hysteresisValue)))
	numShardValidators := numOfShards * numValidatorsOnAShard
	numValidatorsOnMeta := int(math.Ceil(float64(numOfMetachainNodes) * (1 + hysteresisValue)))

	totalAddressesWithBalances := numShardValidators + numValidatorsOnMeta +
		numOfShards*numOfObserversPerShard + numOfMetachainObservers

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
		delegatorsFile       *os.File
		suite                crypto.Suite
		balancesKeyGenerator crypto.KeyGenerator
	)

	defer func() {
		closeFile(walletKeyFile)
		closeFile(validatorKeyFile)
		closeFile(genesisFile)
		closeFile(nodesFile)
		closeFile(txgenAccountsFile)
		closeFile(delegatorsFile)
	}()

	walletKeyFile, err = createNewFile(outputDirectory, walletKeyFileName)
	if err != nil {
		return err
	}

	validatorKeyFile, err = createNewFile(outputDirectory, validatorKeyFileName)
	if err != nil {
		return err
	}

	genesisFile, err = createNewFile(outputDirectory, genesisFilename)
	if err != nil {
		return err
	}

	nodesFile, err = createNewFile(outputDirectory, nodesSetupFilename)
	if err != nil {
		return err
	}

	if generateTxgenFile {
		txgenAccountsFile, err = createNewFile(outputDirectory, txgenAccountsFileName)
		if err != nil {
			return err
		}
	}

	genesisList := make([]*data.InitialAccount, 0)

	var initialNodes []*sharding.InitialNode
	nodes := &sharding.NodesSetup{
		StartTime:                   0,
		RoundDuration:               defaultRoundDuration,
		ConsensusGroupSize:          uint32(consensusGroupSize),
		MinNodesPerShard:            uint32(numOfNodesPerShard),
		MetaChainConsensusGroupSize: uint32(metachainConsensusGroupSize),
		MetaChainMinNodes:           uint32(numOfMetachainNodes),
		Hysteresis:                  float32(hysteresisValue),
		Adaptivity:                  adaptivityValue,
		ChainID:                     chainID,
		MinTransactionVersion:       uint32(txVersion),
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

	stakeTypeString := ctx.GlobalString(stakeType.Name)
	if stakeTypeString == stakedType {
		numDelegatorsValue = 0
	}

	//initialTotalBalance = totalSupply - (numValidators * nodePriceValue) - (numDelegators * initialBalanceForDelegators)
	initialTotalBalance := big.NewInt(0).Set(totalSupplyValue)
	staked := big.NewInt(0).Mul(big.NewInt(int64(numValidators)), nodePriceValue)
	initialTotalBalance.Sub(initialTotalBalance, staked)
	totalDelegatorsBalance := big.NewInt(0).Set(initialBalanceForDelegators)
	totalDelegatorsBalance.Mul(totalDelegatorsBalance, big.NewInt(int64(numDelegatorsValue)))
	initialTotalBalance.Sub(initialTotalBalance, totalDelegatorsBalance)

	initialNodeBalance := big.NewInt(0).Set(initialTotalBalance)
	richestAccount := big.NewInt(0)
	totalNodes := big.NewInt(int64(totalAddressesWithBalances + numOfAdditionalAccounts))
	if !withRichestAccount {
		// initialNodeBalance = initialTotalBalance / (totalAddressesWithBalances + numOfAdditionalAccounts)
		initialNodeBalance.Div(initialNodeBalance, totalNodes)
	} else {
		initialNodeBalance.Set(minimumInitialBalance)
	}
	log.Info("supply values",
		"total supply", totalSupplyValue.String(),
		"staked", staked.String(),
		"initial total balance", initialTotalBalance.String(),
		"initial node balance", initialNodeBalance.String(),
	)
	log.Info("nodes",
		"num nodes with balance", totalAddressesWithBalances,
		"num additional accounts", numOfAdditionalAccounts,
		"num validators", numValidators,
		"num observers", numObservers,
	)

	delegationScAddress := ""
	stakedValue := big.NewInt(0).Set(nodePriceValue)
	if stakeTypeString == delegatedStakeType {
		if numDelegatorsValue == 0 {
			return fmt.Errorf("can not have 0 delegators")
		}

		delegatorsFile, err = createNewFile(outputDirectory, delegatorsFileName)
		if err != nil {
			return err
		}

		delegationScAddress, err = generateScDelegationAddress(ctx.GlobalString(delegationOwnerPublicKey.Name), pubKeyConverterTxs)
		if err != nil {
			return fmt.Errorf("%w when generationg resulted delegation address", err)
		}

		stakedValue = big.NewInt(0)

		genesisList, err = manageDelegators(
			genesisList,
			numValidators,
			nodePriceValue,
			numDelegatorsValue,
			balancesKeyGenerator,
			pubKeyConverterTxs,
			delegatorsFile,
			stakedValue,
			delegationScAddress,
			initialBalanceForDelegators,
		)
		if err != nil {
			return err
		}
	}

	log.Info("started generating...")
	for i := 0; i < totalAddressesWithBalances; i++ {
		isValidator := i < numValidators

		ia, node, err := createInitialAccount(
			balancesKeyGenerator,
			pubKeyConverterTxs,
			pubKeyConverterBlocks,
			initialNodeBalance,
			stakedValue,
			delegationScAddress,
			walletKeyFile,
			validatorKeyFile,
			isValidator,
			stakeTypeString,
			uint32(initialRating),
		)
		if err != nil {
			return err
		}

		genesisList = append(genesisList, ia)
		if node != nil {
			nodes.InitialNodes = append(nodes.InitialNodes, node)
		}
	}

	for i := 0; i < numOfAdditionalAccounts; i++ {
		ia, txGenAccount, shardID, err := createAdditionalNode(
			balancesKeyGenerator,
			pubKeyConverterTxs,
			initialNodeBalance,
			shardCoordinator,
			generateTxgenFile,
		)
		if err != nil {
			return err
		}

		if txGenAccount != nil {
			txgenAccounts[shardID] = append(txgenAccounts[shardID], txGenAccount)
		}

		genesisList = append(genesisList, ia)
	}

	if withRichestAccount {
		ia, _, _, err := createAdditionalNode(
			balancesKeyGenerator,
			pubKeyConverterTxs,
			richestAccount,
			shardCoordinator,
			generateTxgenFile,
		)
		if err != nil {
			return err
		}

		genesisList = append(genesisList, ia)
	}

	remainder := computeRemainder(
		initialTotalBalance,
		totalAddressesWithBalances+numOfAdditionalAccounts,
		initialNodeBalance,
	)

	firstInitialAccount := genesisList[0]
	firstInitialAccount.Supply.Add(firstInitialAccount.Supply, remainder)
	firstInitialAccount.Balance.Add(firstInitialAccount.Balance, remainder)

	err = checkValues(genesisList, totalSupplyValue)
	if err != nil {
		return err
	}

	err = writeDataInFile(genesisFile, genesisList)
	if err != nil {
		return err
	}

	err = writeDataInFile(nodesFile, nodes)
	if err != nil {
		return err
	}

	if generateTxgenFile {
		err = writeDataInFile(txgenAccountsFile, txgenAccounts)
		if err != nil {
			return err
		}
	}

	log.Info("elapsed time", "value", time.Since(startTime))
	log.Info("files generated successfully!")
	return nil
}

func createInitialAccount(
	balancesKeyGenerator crypto.KeyGenerator,
	pubKeyConverterTxs core.PubkeyConverter,
	pubKeyConverterBlocks core.PubkeyConverter,
	initialNodeBalance *big.Int,
	stakedValue *big.Int,
	delegationScAddress string,
	walletKeyFile *os.File,
	validatorKeyFile *os.File,
	isValidator bool,
	stakeTypeString string,
	initialRating uint32,
) (*data.InitialAccount, *sharding.InitialNode, error) {

	pkString, skHex, err := getIdentifierAndPrivateKey(balancesKeyGenerator, pubKeyConverterTxs)
	if err != nil {
		return nil, nil, err
	}

	supply := big.NewInt(0).Set(big.NewInt(0).Set(initialNodeBalance))
	supply.Add(supply, stakedValue)
	ia := &data.InitialAccount{
		Address:      pkString,
		Supply:       supply,
		Balance:      big.NewInt(0).Set(initialNodeBalance),
		StakingValue: stakedValue,
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}

	err = core.SaveSkToPemFile(walletKeyFile, pkString, skHex)
	if err != nil {
		return nil, nil, err
	}

	keyGen := getNodesKeyGen()
	if keyGen == nil {
		return nil, nil, errCreatingKeygen
	}

	pkHexForNode, skHexForNode, err := getIdentifierAndPrivateKey(keyGen, pubKeyConverterBlocks)
	if err != nil {
		return nil, nil, err
	}

	err = core.SaveSkToPemFile(validatorKeyFile, pkHexForNode, skHexForNode)
	if err != nil {
		return nil, nil, err
	}

	var node *sharding.InitialNode
	if isValidator {
		node = &sharding.InitialNode{
			PubKey:        pkHexForNode,
			Address:       pkString,
			InitialRating: initialRating,
		}

		if stakeTypeString == delegatedStakeType {
			node.Address = delegationScAddress
		}
	} else {
		ia.StakingValue = big.NewInt(0)
		ia.Supply = big.NewInt(0).Set(ia.Balance)
		ia.Delegation.Address = ""
		ia.Delegation.Value = big.NewInt(0)
	}

	return ia, node, nil
}

func createAdditionalNode(
	balancesKeyGenerator crypto.KeyGenerator,
	pubKeyConverterTxs core.PubkeyConverter,
	initialNodeBalance *big.Int,
	shardCoordinator sharding.Coordinator,
	generateTxgenFile bool,
) (*data.InitialAccount, *txgenAccount, uint32, error) {

	pkString, skHex, err := getIdentifierAndPrivateKey(balancesKeyGenerator, pubKeyConverterTxs)
	if err != nil {
		return nil, nil, 0, err
	}

	ia := &data.InitialAccount{
		Address:      pkString,
		Supply:       big.NewInt(0).Set(initialNodeBalance),
		Balance:      big.NewInt(0).Set(initialNodeBalance),
		StakingValue: big.NewInt(0),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}

	var txGen *txgenAccount
	var shId uint32
	if generateTxgenFile {
		pkBytes, _ := pubKeyConverterTxs.Decode(pkString)
		shId = shardCoordinator.ComputeId(pkBytes)

		txGen = &txgenAccount{
			PubKey:        pkString,
			PrivKey:       string(skHex),
			LastNonce:     0,
			Balance:       big.NewInt(0).Set(initialNodeBalance),
			TokenBalance:  big.NewInt(0),
			CanReuseNonce: true,
		}
	}

	return ia, txGen, shId, nil
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

func closeFile(f *os.File) {
	if f != nil {
		err := f.Close()
		log.LogIfError(err)
	}
}

func createNewFile(outputDirectory string, fileName string) (*os.File, error) {
	filePath := filepath.Join(outputDirectory, fileName)
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
}

func writeDataInFile(file *os.File, data interface{}) error {
	buff, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(buff)

	return err
}

func generateScDelegationAddress(pkString string, converter core.PubkeyConverter) (string, error) {
	blockchainHook, err := generateBlockchainHook(converter)
	if err != nil {
		return "", err
	}

	pk, err := converter.Decode(pkString)
	if err != nil {
		return "", err
	}

	vmTypeBytes, err := hex.DecodeString(vmType)
	if err != nil {
		return "", err
	}

	scResultingAddressBytes, err := blockchainHook.NewAddress(pk, delegationOwnerNonce, vmTypeBytes)
	if err != nil {
		return "", err
	}

	return converter.Encode(scResultingAddressBytes), nil
}

func generateBlockchainHook(converter core.PubkeyConverter) (process.BlockChainHookHandler, error) {
	builtInFuncs := builtInFunctions.NewBuiltInFunctionContainer()
	arg := hooks.ArgBlockChainHook{
		Accounts:         &mock.AccountsStub{},
		PubkeyConv:       converter,
		StorageService:   &mock.ChainStorerMock{},
		BlockChain:       &mock.BlockChainMock{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		Marshalizer:      &mock.MarshalizerMock{},
		Uint64Converter:  &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions: builtInFuncs,
	}

	return hooks.NewBlockChainHookImpl(arg)
}

func prepareOutputDirectory(outputDirectory string) error {
	_, err := os.Stat(outputDirectory)
	if os.IsNotExist(err) {
		return os.MkdirAll(outputDirectory, 0755)
	}

	return err
}

func computeRemainder(total *big.Int, numUnit int, valPerUnit *big.Int) *big.Int {
	subs := big.NewInt(int64(numUnit))
	subs.Mul(subs, valPerUnit)
	remainder := big.NewInt(0).Set(total)
	remainder.Sub(remainder, subs)

	return remainder
}

func checkValues(genesisList []*data.InitialAccount, totalSupplyValue *big.Int) error {
	supply := big.NewInt(0)
	staked := big.NewInt(0)
	balance := big.NewInt(0)
	delegation := big.NewInt(0)
	for idx, ia := range genesisList {
		supply.Add(supply, ia.Supply)
		staked.Add(staked, ia.StakingValue)
		balance.Add(balance, ia.Balance)
		delegation.Add(delegation, ia.Delegation.Value)

		accountSupply := big.NewInt(0).Set(ia.Supply)
		accountSupply.Sub(accountSupply, ia.StakingValue)
		accountSupply.Sub(accountSupply, ia.Balance)
		accountSupply.Sub(accountSupply, ia.Delegation.Value)

		if accountSupply.Cmp(zero) != 0 {
			return fmt.Errorf("error generation account (supply mismatch), "+
				"index %d, address %s, supply %s, staked %s, balance %s, delegated %s",
				idx, ia.Address, ia.Supply, ia.StakingValue, ia.Balance, ia.Delegation.Value)
		}
	}

	supplyCopy := big.NewInt(0).Set(supply)
	supplyCopy.Sub(supplyCopy, staked)
	supplyCopy.Sub(supplyCopy, balance)
	supplyCopy.Sub(supplyCopy, delegation)

	if supplyCopy.Cmp(zero) != 0 {
		return fmt.Errorf("supply does not match supply %s, staked %s, balance %s, delegated %s",
			supply.String(), staked.String(), balance.String(), delegation.String())
	}

	if supply.Cmp(totalSupplyValue) != 0 {
		return fmt.Errorf("supply does not match expected %s, got %s",
			totalSupplyValue.String(), supply.String())
	}

	log.Info("values",
		"total supply", supply.String(),
		"total staked", staked.String(),
		"total balance", balance.String(),
		"total delegated", delegation.String(),
	)

	return nil
}

func manageDelegators(
	genesisList []*data.InitialAccount,
	numValidators int,
	nodePrice *big.Int,
	numDelegators uint,
	balancesKeyGenerator crypto.KeyGenerator,
	pubKeyConverterTxs core.PubkeyConverter,
	delegatorsFile *os.File,
	stakedValue *big.Int,
	delegationScAddress string,
	initialBalanceForDelegators *big.Int,
) ([]*data.InitialAccount, error) {
	totalDelegationNeeded := big.NewInt(int64(numValidators))
	totalDelegationNeeded.Mul(totalDelegationNeeded, nodePrice)
	delegatorValue := big.NewInt(0).Set(totalDelegationNeeded)
	delegatorValue.Div(delegatorValue, big.NewInt(int64(numDelegators)))

	remainder := computeRemainder(totalDelegationNeeded, int(numDelegators), delegatorValue)

	log.Info("delegators info",
		"total delegation needed", totalDelegationNeeded.String(),
		"num delegators", numDelegators,
		"delegator value", delegatorValue.String(),
		"remainder", remainder.String(),
	)

	for i := uint(0); i < numDelegators; i++ {
		pk, sk, err := getIdentifierAndPrivateKey(balancesKeyGenerator, pubKeyConverterTxs)
		if err != nil {
			return nil, err
		}

		err = core.SaveSkToPemFile(delegatorsFile, pk, sk)
		if err != nil {
			return nil, err
		}

		ia := &data.InitialAccount{
			Address:      pk,
			Supply:       big.NewInt(0).Set(delegatorValue),
			Balance:      big.NewInt(0),
			StakingValue: stakedValue,
			Delegation: &data.DelegationData{
				Address: delegationScAddress,
				Value:   big.NewInt(0).Set(delegatorValue),
			},
		}
		ia.Supply.Add(ia.Supply, initialBalanceForDelegators)
		ia.Balance.Add(ia.Balance, initialBalanceForDelegators)

		if i == 0 {
			//treat the remainder on the first delegator
			ia.Supply.Add(ia.Supply, remainder)
			ia.Delegation.Value.Add(ia.Delegation.Value, remainder)
		}

		genesisList = append(genesisList, ia)
	}

	return genesisList, nil
}
