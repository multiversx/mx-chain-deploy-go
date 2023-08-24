package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/ElrondNetwork/elrond-deploy-go/check"
	"github.com/ElrondNetwork/elrond-deploy-go/core"
	"github.com/ElrondNetwork/elrond-deploy-go/generate/factory"
	"github.com/ElrondNetwork/elrond-deploy-go/plugins"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondCommonFactory "github.com/ElrondNetwork/elrond-go/common/factory"
	"github.com/ElrondNetwork/elrond-go/config"
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
		Usage: "number of delegators if the stake-type is of type `delegated` or `mixed`",
		Value: 100,
	}
	numDelegatedNodes = cli.UintFlag{
		Name:  "num-delegated-nodes",
		Usage: "number of delegated nodes if the stake-type is of type `mixed`",
		Value: 4,
	}
	maxNumValidatorsPerOwner = cli.UintFlag{
		Name:  "max-num-validators-per-node",
		Usage: "maximum number of validators held by an owner. The value will vary between [1-max] randomly.",
		Value: 1,
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
		txgenFile,
		stakeType,
		delegationOwnerPublicKey,
		numDelegators,
		richestAccount,
		numDelegatedNodes,
		maxNumValidatorsPerOwner,
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
	numOfShardsValue := ctx.GlobalInt(numOfShards.Name)
	numOfNodesPerShardValue := ctx.GlobalInt(numOfNodesPerShard.Name)
	consensusGroupSizeValue := ctx.GlobalInt(consensusGroupSize.Name)
	numOfObserversPerShardValue := ctx.GlobalInt(numOfObserversPerShard.Name)
	numOfMetachainNodesValue := ctx.GlobalInt(numOfMetachainNodes.Name)
	metachainConsensusGroupSizeValue := ctx.GlobalInt(metachainConsensusGroupSize.Name)
	numOfMetachainObserversValue := ctx.GlobalInt(numOfMetachainObservers.Name)
	numOfAdditionalAccountsValue := ctx.GlobalInt(numAdditionalAccountsInGenesis.Name)
	initialRatingValue := ctx.GlobalUint64(initialRating.Name)
	hysteresisValue := ctx.GlobalFloat64(hysteresis.Name)
	adaptivityValue := ctx.GlobalBool(adaptivity.Name)
	generateTxgenFile := ctx.IsSet(txgenFile.Name)
	numDelegatorsValue := ctx.GlobalUint(numDelegators.Name)
	withRichestAccount := ctx.GlobalBool(richestAccount.Name)
	stakeTypeString := ctx.GlobalString(stakeType.Name)
	delegationOwnerPkString := ctx.GlobalString(delegationOwnerPublicKey.Name)
	numDelegatedNodesValue := ctx.GlobalUint(numDelegatedNodes.Name)
	maxNumValidatorsPerOwnerValue := ctx.GlobalUint(maxNumValidatorsPerOwner.Name)

	err = prepareOutputDirectory(outputDirectory)
	if err != nil {
		return err
	}

	numValidatorsOnAShard := int(math.Ceil(float64(numOfNodesPerShardValue) * (1 + hysteresisValue)))
	numShardValidators := numOfShardsValue * numValidatorsOnAShard
	numValidatorsOnMeta := int(math.Ceil(float64(metachainConsensusGroupSizeValue) * (1 + hysteresisValue)))
	numValidators := numShardValidators + numValidatorsOnMeta
	numObservers := numOfShardsValue*numOfObserversPerShardValue + numOfMetachainObserversValue

	invalidNumPrivPubKey := numValidators < 1 ||
		numOfShardsValue < 1 ||
		numOfNodesPerShardValue < 1
	if invalidNumPrivPubKey {
		return errInvalidNumPrivPubKeys
	}

	invalidNumOfNodes := consensusGroupSizeValue < 1 ||
		consensusGroupSizeValue > numOfNodesPerShardValue ||
		numOfObserversPerShardValue < 0

	if invalidNumOfNodes {
		return errInvalidNumOfNodes
	}

	totalSupplyString := ctx.GlobalString(totalSupply.Name)
	totalSupplyValue, err := core.ConvertToPositiveBigInt(totalSupplyString)
	if err != nil {
		return err
	}

	nodePriceString := ctx.GlobalString(nodePrice.Name)
	nodePriceValue, err := core.ConvertToPositiveBigInt(nodePriceString)
	if err != nil {
		return err
	}

	validatorPubKeyConverter, walletPubKeyConverter, err := createPubKeyConverters()
	validatorKeyGenerator, walletKeyGenerator := createKeyGenerators()

	shardCoordinator, err := sharding.NewMultiShardCoordinator(uint32(numOfShardsValue), 0)
	if err != nil {
		return err
	}

	argOutputHandler, err := plugins.CreateOutputHandlerArgument(
		outputDirectory,
		validatorPubKeyConverter,
		walletPubKeyConverter,
		shardCoordinator,
		generateTxgenFile,
		stakeTypeString == core.DelegatedStakeType || stakeTypeString == core.MixedType,
	)
	if err != nil {
		return err
	}
	argOutputHandler.RoundDuration = defaultRoundDuration
	argOutputHandler.ConsensusGroupSize = consensusGroupSizeValue
	argOutputHandler.NumOfNodesPerShard = numOfNodesPerShardValue
	argOutputHandler.MetachainConsensusGroupSize = metachainConsensusGroupSizeValue
	argOutputHandler.NumOfMetachainNodes = numOfMetachainNodesValue
	argOutputHandler.HysteresisValue = float32(hysteresisValue)
	argOutputHandler.AdaptivityValue = adaptivityValue

	outputHandler, err := plugins.NewOutputHandler(argOutputHandler)
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
		MaxNumNodesOnOwner:        maxNumValidatorsPerOwnerValue,
		NumAdditionalWalletKeys:   uint(numOfAdditionalAccountsValue),
		IntRandomizer:             &random.ConcurrentSafeIntRandomizer{},
		NodePrice:                 nodePriceValue,
		TotalSupply:               totalSupplyValue,
		InitialRating:             initialRatingValue,
		GenerationType:            stakeTypeString,
		DelegationOwnerPkString:   delegationOwnerPkString,
		DelegationOwnerNonce:      delegationOwnerNonce,
		VmType:                    vmType,
		NumDelegators:             numDelegatorsValue,
		NumDelegatedNodes:         numDelegatedNodesValue,
	}

	dataGenerator, err := factory.CreateDataGenerator(argDataGenerator)
	if err != nil {
		return err
	}

	generatedOutput, err := dataGenerator.Generate()
	if err != nil {
		return err
	}

	initialAccountChecker, err := check.NewInitialAccountsChecker(nodePriceValue, totalSupplyValue)
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
	walletPubKeyConverter, err := elrondCommonFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length: 32,
		Type:   walletPubKeyFormat,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("%w for walletPubKeyConverter", err)
	}

	validatorPubKeyConverter, err := elrondCommonFactory.NewPubkeyConverter(config.PubkeyConfig{
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
