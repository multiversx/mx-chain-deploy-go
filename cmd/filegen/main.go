package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/ElrondNetwork/elrond-deploy-go/checking"
	"github.com/ElrondNetwork/elrond-deploy-go/core"
	"github.com/ElrondNetwork/elrond-deploy-go/generating/factory"
	"github.com/ElrondNetwork/elrond-deploy-go/io"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	elrondCore "github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/random"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	elrondFactory "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/urfave/cli"
)

const defaultRoundDuration = 5000
const walletPubKeyFormat = "bech32"
const validatorPubKeyFormat = "hex"
const vmType = "0500"
const delegationOwnerNonce = uint64(0)

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
		Value: 3,
	}
	numOfNodesPerShard = cli.IntFlag{
		Name:  "num-of-nodes-in-each-shard",
		Usage: "Number of initial nodes in each shard, private/public keys, to generate",
		Value: 7,
	}
	consensusGroupSize = cli.IntFlag{
		Name:  "consensus-group-size",
		Usage: "Consensus group size",
		Value: 5,
	}
	numOfObserversPerShard = cli.IntFlag{
		Name:  "num-of-observers-in-each-shard",
		Usage: "Number of initial observers in each shard, private/public keys, to generate",
		Value: 1,
	}
	numOfMetachainNodes = cli.IntFlag{
		Name:  "num-of-metachain-nodes",
		Usage: "Number of initial metachain nodes, private/public keys, to generate",
		Value: 7,
	}
	metachainConsensusGroupSize = cli.IntFlag{
		Name:  "metachain-consensus-group-size",
		Usage: "Metachain consensus group size",
		Value: 7,
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
		Value: 0.2,
	}
	adaptivity = cli.BoolFlag{
		Name:  "adaptivity",
		Usage: "Adaptivity value - should be set to true if shard merging and splitting is enabled",
	}
	chainID = cli.StringFlag{
		Name:  "chain-id",
		Usage: "Chain ID flag",
		Value: "T",
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
		Value: "delegated",
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

	errInvalidNumPrivPubKeys = errors.New("invalid number of private/public keys to generate")
	errInvalidNumOfNodes     = errors.New("invalid number of nodes in shard/metachain or in the consensus group")
	log                      = logger.GetOrCreate("main")
)

// The resulting binary will be used to generate 2 files: genesis.json and privkeys.pem
// Those files are used to mass-deploy nodes and thus, ensuring that all nodes have the same data to work with
// The 2 optional flags are used to specify how many private/public keys to generate and the initial minting for each
// public generated key
func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = fileGenHelpTemplate
	app.Name = "Deploy Preparation Tool"
	app.Version = "v1.0.0"
	app.Usage = "This binary will generate genesis json and pem files" +
		" files, to be used in mass deployment"
	app.Flags = []cli.Flag{
		outputDirectoryFlag,
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
		return generate(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func generate(ctx *cli.Context) error {
	var err error

	startTime := time.Now()
	outputDirectory := ctx.GlobalString(outputDirectoryFlag.Name)
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
	stakeTypeString := ctx.GlobalString(stakeType.Name)
	delegationOwnerPkString := ctx.GlobalString(delegationOwnerPublicKey.Name)

	err = prepareOutputDirectory(outputDirectory)
	if err != nil {
		return err
	}

	numValidatorsOnAShard := int(math.Ceil(float64(numOfNodesPerShard) * (1 + hysteresisValue)))
	numShardValidators := numOfShards * numValidatorsOnAShard
	numValidatorsOnMeta := int(math.Ceil(float64(numOfMetachainNodes) * (1 + hysteresisValue)))
	numValidators := numShardValidators + numValidatorsOnMeta
	numObservers := numOfShards*numOfObserversPerShard + numOfMetachainObservers

	invalidNumPrivPubKey := numValidators < 1 ||
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
	totalSupplyValue, err := core.ConvertToBigInt(totalSupplyString)
	if err != nil {
		return err
	}

	nodePriceString := ctx.GlobalString(nodePrice.Name)
	nodePriceValue, err := core.ConvertToBigInt(nodePriceString)
	if err != nil {
		return err
	}

	validatorPubKeyConverter, walletPubKeyConverter, err := createPubKeyConverters()
	validatorKeyGenerator, walletKeyGenerator := createKeyGenerators()

	shardCoordinator, err := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	if err != nil {
		return err
	}

	argOutputHandler, err := io.CreateOutputHandlerArgument(
		outputDirectory,
		validatorPubKeyConverter,
		walletPubKeyConverter,
		shardCoordinator,
		generateTxgenFile,
		stakeTypeString == core.DelegatedStakeType,
	)
	if err != nil {
		return err
	}
	argOutputHandler.RoundDuration = defaultRoundDuration
	argOutputHandler.ConsensusGroupSize = consensusGroupSize
	argOutputHandler.NumOfNodesPerShard = numOfNodesPerShard
	argOutputHandler.MetachainConsensusGroupSize = metachainConsensusGroupSize
	argOutputHandler.NumOfMetachainNodes = numOfMetachainNodes
	argOutputHandler.HysteresisValue = float32(hysteresisValue)
	argOutputHandler.AdaptivityValue = adaptivityValue
	argOutputHandler.ChainID = chainID
	argOutputHandler.TxVersion = txVersion

	outputHandler, err := io.NewOutputHandler(argOutputHandler)
	if err != nil {
		return err
	}

	defer outputHandler.Close()

	argDataGenerator := factory.ArgDataGenerator{
		KeyGeneratorForValidators: validatorKeyGenerator,
		KeyGeneratorForWallets:    walletKeyGenerator,
		WalletPubKeyConverter:     walletPubKeyConverter,
		ValidatorPubKeyConverter:  validatorPubKeyConverter,
		NumValidatorBlsKeys:       uint(numValidators),
		NumObserverBlsKeys:        uint(numObservers),
		RichestAccountMode:        withRichestAccount,
		MaxNumNodesOnOwner:        1, //TODO in next PR: replace this with a flag
		NumAdditionalWalletKeys:   uint(numOfAdditionalAccounts),
		IntRandomizer:             &random.ConcurrentSafeIntRandomizer{},
		NodePrice:                 nodePriceValue,
		TotalSupply:               totalSupplyValue,
		InitialRating:             initialRating,
		GenerationType:            stakeTypeString,
		DelegationOwnerPkString:   delegationOwnerPkString,
		DelegationOwnerNonce:      delegationOwnerNonce,
		VmType:                    vmType,
		NumDelegators:             numDelegatorsValue,
	}

	dataGenerator, err := factory.CreateDataGenerator(argDataGenerator)
	if err != nil {
		return err
	}

	generatedOutput, err := dataGenerator.Generate()
	if err != nil {
		return err
	}

	initialAccountChecker, err := checking.NewInitialAccountsChecker(nodePriceValue, totalSupplyValue)
	if err != nil {
		return err
	}

	err = initialAccountChecker.CheckInitialAccounts(generatedOutput.InitialAccounts)
	if err != nil {
		return err
	}

	err = outputHandler.WriteData(*generatedOutput)
	if err != nil {
		return err
	}

	log.Info("elapsed time", "value", time.Since(startTime))
	log.Info("files generated successfully!")
	return nil
}

func prepareOutputDirectory(outputDirectory string) error {
	_, err := os.Stat(outputDirectory)
	if os.IsNotExist(err) {
		return os.MkdirAll(outputDirectory, 0755)
	}

	return err
}

func createPubKeyConverters() (elrondCore.PubkeyConverter, elrondCore.PubkeyConverter, error) {
	walletPubKeyConverter, err := elrondFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length: 32,
		Type:   walletPubKeyFormat,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("%w for walletPubKeyConverter", err)
	}

	validatorPubKeyConverter, err := elrondFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length: 96,
		Type:   validatorPubKeyFormat,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("%w for validatorPubKeyConverter", err)
	}

	return validatorPubKeyConverter, walletPubKeyConverter, nil
}

func createKeyGenerators() (crypto.KeyGenerator, crypto.KeyGenerator) {
	walletSuite := ed25519.NewEd25519()
	walletKeyGenerator := signing.NewKeyGenerator(walletSuite)

	validatorSuite := mcl.NewSuiteBLS12()
	validatorKeyGenerator := signing.NewKeyGenerator(validatorSuite)

	return validatorKeyGenerator, walletKeyGenerator
}

//func generateFiles(ctx *cli.Context) error {

//
//
//	numObservers := numOfShards*numOfObserversPerShard + numOfMetachainObservers
//	numValidators := totalAddressesWithBalances - numObservers
//

//
//	//initialTotalBalance = totalSupply - (numValidators * nodePriceValue) - (numDelegators * initialBalanceForDelegators)
//	initialTotalBalance := big.NewInt(0).Set(totalSupplyValue)
//	staked := big.NewInt(0).Mul(big.NewInt(int64(numValidators)), nodePriceValue)
//	initialTotalBalance.Sub(initialTotalBalance, staked)
//	totalDelegatorsBalance := big.NewInt(0).Set(initialBalanceForDelegators)
//	totalDelegatorsBalance.Mul(totalDelegatorsBalance, big.NewInt(int64(numDelegatorsValue)))
//	initialTotalBalance.Sub(initialTotalBalance, totalDelegatorsBalance)
//
//	initialNodeBalance := big.NewInt(0).Set(initialTotalBalance)
//	totalNodes := big.NewInt(int64(totalAddressesWithBalances + numOfAdditionalAccounts))
//	if !withRichestAccount {
//		// initialNodeBalance = initialTotalBalance / (totalAddressesWithBalances + numOfAdditionalAccounts)
//		initialNodeBalance.Div(initialNodeBalance, totalNodes)
//	} else {
//		initialNodeBalance.Set(minimumInitialBalance)
//	}
//	log.Info("supply values",
//		"total supply", totalSupplyValue.String(),
//		"staked", staked.String(),
//		"initial total balance", initialTotalBalance.String(),
//		"initial node balance", initialNodeBalance.String(),
//	)
//	log.Info("nodes",
//		"num nodes with balance", totalAddressesWithBalances,
//		"num additional accounts", numOfAdditionalAccounts,
//		"num validators", numValidators,
//		"num observers", numObservers,
//	)
//
//	delegationScAddress := ""
//	stakedValue := big.NewInt(0).Set(nodePriceValue)
//	if stakeTypeString == delegatedStakeType {
//		if numDelegatorsValue == 0 {
//			return fmt.Errorf("can not have 0 delegators")
//		}
//
//		delegatorsFile, err = createNewFile(outputDirectory, delegatorsFileName)
//		if err != nil {
//			return err
//		}
//
//		delegationScAddress, err = generateScDelegationAddress(ctx.GlobalString(delegationOwnerPublicKey.Name), pubKeyConverterTxs)
//		if err != nil {
//			return fmt.Errorf("%w when generationg resulted delegation address", err)
//		}
//
//		stakedValue = big.NewInt(0)
//
//		genesisList, err = manageDelegators(
//			genesisList,
//			numValidators,
//			nodePriceValue,
//			numDelegatorsValue,
//			balancesKeyGenerator,
//			pubKeyConverterTxs,
//			delegatorsFile,
//			stakedValue,
//			delegationScAddress,
//			initialBalanceForDelegators,
//		)
//		if err != nil {
//			return err
//		}
//	}
//
//	log.Info("started generating...")
//	for i := 0; i < totalAddressesWithBalances; i++ {
//		isValidator := i < numValidators
//
//		ia, node, err := createInitialAccount(
//			balancesKeyGenerator,
//			pubKeyConverterTxs,
//			pubKeyConverterBlocks,
//			initialNodeBalance,
//			stakedValue,
//			delegationScAddress,
//			walletKeyFile,
//			validatorKeyFile,
//			isValidator,
//			stakeTypeString,
//			uint32(initialRating),
//		)
//		if err != nil {
//			return err
//		}
//
//		genesisList = append(genesisList, ia)
//		if node != nil {
//			nodes.InitialNodes = append(nodes.InitialNodes, node)
//		}
//	}
//
//	for i := 0; i < numOfAdditionalAccounts; i++ {
//		ia, txGenAccount, shardID, err := createAdditionalNode(
//			balancesKeyGenerator,
//			pubKeyConverterTxs,
//			initialNodeBalance,
//			shardCoordinator,
//			generateTxgenFile,
//		)
//		if err != nil {
//			return err
//		}
//
//		if txGenAccount != nil {
//			txgenAccounts[shardID] = append(txgenAccounts[shardID], txGenAccount)
//		}
//
//		genesisList = append(genesisList, ia)
//	}
//
//	remainder := computeRemainder(
//		initialTotalBalance,
//		totalAddressesWithBalances+numOfAdditionalAccounts,
//		initialNodeBalance,
//	)
//
//	firstInitialAccount := genesisList[0]
//	firstInitialAccount.Supply.Add(firstInitialAccount.Supply, remainder)
//	firstInitialAccount.Balance.Add(firstInitialAccount.Balance, remainder)
//
//	err = checkValues(genesisList, totalSupplyValue)
//	if err != nil {
//		return err
//	}
//
//	err = writeDataInFile(genesisFile, genesisList)
//	if err != nil {
//		return err
//	}
//
//	err = writeDataInFile(nodesFile, nodes)
//	if err != nil {
//		return err
//	}
//
//	if generateTxgenFile {
//		err = writeDataInFile(txgenAccountsFile, txgenAccounts)
//		if err != nil {
//			return err
//		}
//	}
//
//	log.Info("elapsed time", "value", time.Since(startTime))
//	log.Info("files generated successfully!")
//	return nil
//}
//
//func createInitialAccount(
//	balancesKeyGenerator crypto.KeyGenerator,
//	pubKeyConverterTxs core.PubkeyConverter,
//	pubKeyConverterBlocks core.PubkeyConverter,
//	initialNodeBalance *big.Int,
//	stakedValue *big.Int,
//	delegationScAddress string,
//	walletKeyFile *os.File,
//	validatorKeyFile *os.File,
//	isValidator bool,
//	stakeTypeString string,
//	initialRating uint32,
//) (*data.InitialAccount, *sharding.InitialNode, error) {
//
//	pkString, skHex, err := getIdentifierAndPrivateKey(balancesKeyGenerator, pubKeyConverterTxs)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	supply := big.NewInt(0).Set(big.NewInt(0).Set(initialNodeBalance))
//	supply.Add(supply, stakedValue)
//	ia := &data.InitialAccount{
//		Address:      pkString,
//		Supply:       supply,
//		Balance:      big.NewInt(0).Set(initialNodeBalance),
//		StakingValue: stakedValue,
//		Delegation: &data.DelegationData{
//			Address: "",
//			Value:   big.NewInt(0),
//		},
//	}
//
//	err = core.SaveSkToPemFile(walletKeyFile, pkString, skHex)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	keyGen := getNodesKeyGen()
//	if keyGen == nil {
//		return nil, nil, errCreatingKeygen
//	}
//
//	pkHexForNode, skHexForNode, err := getIdentifierAndPrivateKey(keyGen, pubKeyConverterBlocks)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	err = core.SaveSkToPemFile(validatorKeyFile, pkHexForNode, skHexForNode)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	var node *sharding.InitialNode
//	if isValidator {
//		node = &sharding.InitialNode{
//			PubKey:        pkHexForNode,
//			Address:       pkString,
//			InitialRating: initialRating,
//		}
//
//		if stakeTypeString == delegatedStakeType {
//			node.Address = delegationScAddress
//		}
//	} else {
//		ia.StakingValue = big.NewInt(0)
//		ia.Supply = big.NewInt(0).Set(ia.Balance)
//		ia.Delegation.Address = ""
//		ia.Delegation.Value = big.NewInt(0)
//	}
//
//	return ia, node, nil
//}
//
//func createAdditionalNode(
//	balancesKeyGenerator crypto.KeyGenerator,
//	pubKeyConverterTxs core.PubkeyConverter,
//	initialNodeBalance *big.Int,
//	shardCoordinator sharding.Coordinator,
//	generateTxgenFile bool,
//) (*data.InitialAccount, *txgenAccount, uint32, error) {
//
//	pkString, skHex, err := getIdentifierAndPrivateKey(balancesKeyGenerator, pubKeyConverterTxs)
//	if err != nil {
//		return nil, nil, 0, err
//	}
//
//	ia := &data.InitialAccount{
//		Address:      pkString,
//		Supply:       big.NewInt(0).Set(initialNodeBalance),
//		Balance:      big.NewInt(0).Set(initialNodeBalance),
//		StakingValue: big.NewInt(0),
//		Delegation: &data.DelegationData{
//			Address: "",
//			Value:   big.NewInt(0),
//		},
//	}
//
//	var txGen *txgenAccount
//	var shId uint32
//	if generateTxgenFile {
//		pkBytes, _ := pubKeyConverterTxs.Decode(pkString)
//		shId = shardCoordinator.ComputeId(pkBytes)
//
//		txGen = &txgenAccount{
//			PubKey:        pkString,
//			PrivKey:       string(skHex),
//			LastNonce:     0,
//			Balance:       big.NewInt(0).Set(initialNodeBalance),
//			TokenBalance:  big.NewInt(0),
//			CanReuseNonce: true,
//		}
//	}
//
//	return ia, txGen, shId, nil
//}

//func manageDelegators(
//	genesisList []*data.InitialAccount,
//	numValidators int,
//	nodePrice *big.Int,
//	numDelegators uint,
//	balancesKeyGenerator crypto.KeyGenerator,
//	pubKeyConverterTxs core.PubkeyConverter,
//	delegatorsFile *os.File,
//	stakedValue *big.Int,
//	delegationScAddress string,
//	initialBalanceForDelegators *big.Int,
//) ([]*data.InitialAccount, error) {
//	totalDelegationNeeded := big.NewInt(int64(numValidators))
//	totalDelegationNeeded.Mul(totalDelegationNeeded, nodePrice)
//	delegatorValue := big.NewInt(0).Set(totalDelegationNeeded)
//	delegatorValue.Div(delegatorValue, big.NewInt(int64(numDelegators)))
//
//	remainder := computeRemainder(totalDelegationNeeded, int(numDelegators), delegatorValue)
//
//	log.Info("delegators info",
//		"total delegation needed", totalDelegationNeeded.String(),
//		"num delegators", numDelegators,
//		"delegator value", delegatorValue.String(),
//		"remainder", remainder.String(),
//	)
//
//	for i := uint(0); i < numDelegators; i++ {
//		pk, sk, err := getIdentifierAndPrivateKey(balancesKeyGenerator, pubKeyConverterTxs)
//		if err != nil {
//			return nil, err
//		}
//
//		err = core.SaveSkToPemFile(delegatorsFile, pk, sk)
//		if err != nil {
//			return nil, err
//		}
//
//		ia := &data.InitialAccount{
//			Address:      pk,
//			Supply:       big.NewInt(0).Set(delegatorValue),
//			Balance:      big.NewInt(0),
//			StakingValue: stakedValue,
//			Delegation: &data.DelegationData{
//				Address: delegationScAddress,
//				Value:   big.NewInt(0).Set(delegatorValue),
//			},
//		}
//		ia.Supply.Add(ia.Supply, initialBalanceForDelegators)
//		ia.Balance.Add(ia.Balance, initialBalanceForDelegators)
//
//		if i == 0 {
//			//treat the remainder on the first delegator
//			ia.Supply.Add(ia.Supply, remainder)
//			ia.Delegation.Value.Add(ia.Delegation.Value, remainder)
//		}
//
//		genesisList = append(genesisList, ia)
//	}
//
//	return genesisList, nil
//}
