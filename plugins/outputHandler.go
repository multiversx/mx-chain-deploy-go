package plugins

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-deploy-go/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/core"
	elrondData "github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("io")

// ArgOutputHandler represents the output handler constructor argument
type ArgOutputHandler struct {
	WalletHandler               FileHandler
	ValidatorKeyHandler         FileHandler
	GenesisHandler              FileHandler
	NodesSetupHandler           FileHandler
	TxgenAccountsHandler        FileHandler
	DelegatorsHandler           FileHandler
	ValidatorPubKeyConverter    core.PubkeyConverter
	WalletPubKeyConverter       core.PubkeyConverter
	ShardCoordinator            sharding.Coordinator
	RoundDuration               uint64
	ConsensusGroupSize          int
	NumOfNodesPerShard          int
	MetachainConsensusGroupSize int
	NumOfMetachainNodes         int
	HysteresisValue             float32
	AdaptivityValue             bool
	ChainID                     string
	TxVersion                   uint
}

type outputHandler struct {
	walletHandler               FileHandler
	validatorKeyHandler         FileHandler
	genesisHandler              FileHandler
	nodesSetupHandler           FileHandler
	txgenAccountsHandler        FileHandler
	delegatorsHandler           FileHandler
	validatorPubKeyConverter    core.PubkeyConverter
	walletPubKeyConverter       core.PubkeyConverter
	shardCoordinator            sharding.Coordinator
	roundDuration               uint64
	consensusGroupSize          int
	numOfNodesPerShard          int
	metachainConsensusGroupSize int
	numOfMetachainNodes         int
	hysteresisValue             float32
	adaptivityValue             bool
	chainID                     string
	txVersion                   uint
}

// NewOutputHandler will create a new output handler able to write data on disk
func NewOutputHandler(arg ArgOutputHandler) (*outputHandler, error) {
	if check.IfNil(arg.WalletHandler) {
		return nil, fmt.Errorf("%w for WalletHandler", ErrNilFileHandler)
	}
	if check.IfNil(arg.ValidatorKeyHandler) {
		return nil, fmt.Errorf("%w for ValidatorKeyHandler", ErrNilFileHandler)
	}
	if check.IfNil(arg.GenesisHandler) {
		return nil, fmt.Errorf("%w for GenesisHandler", ErrNilFileHandler)
	}
	if check.IfNil(arg.NodesSetupHandler) {
		return nil, fmt.Errorf("%w for NodesSetupHandler", ErrNilFileHandler)
	}
	//TxgenAccountsHandler and DelegatorsHandler can be nil

	if check.IfNil(arg.ValidatorPubKeyConverter) {
		return nil, fmt.Errorf("%w for ValidatorPubKeyConverter", ErrNilPubKeyConverter)
	}
	if check.IfNil(arg.WalletPubKeyConverter) {
		return nil, fmt.Errorf("%w for WalletPubKeyConverter", ErrNilPubKeyConverter)
	}
	if check.IfNil(arg.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}

	return &outputHandler{
		walletHandler:               arg.WalletHandler,
		validatorKeyHandler:         arg.ValidatorKeyHandler,
		genesisHandler:              arg.GenesisHandler,
		nodesSetupHandler:           arg.NodesSetupHandler,
		txgenAccountsHandler:        arg.TxgenAccountsHandler,
		delegatorsHandler:           arg.DelegatorsHandler,
		validatorPubKeyConverter:    arg.ValidatorPubKeyConverter,
		walletPubKeyConverter:       arg.WalletPubKeyConverter,
		shardCoordinator:            arg.ShardCoordinator,
		roundDuration:               arg.RoundDuration,
		consensusGroupSize:          arg.ConsensusGroupSize,
		numOfNodesPerShard:          arg.NumOfNodesPerShard,
		metachainConsensusGroupSize: arg.MetachainConsensusGroupSize,
		numOfMetachainNodes:         arg.NumOfMetachainNodes,
		hysteresisValue:             arg.HysteresisValue,
		adaptivityValue:             arg.AdaptivityValue,
		chainID:                     arg.ChainID,
		txVersion:                   arg.TxVersion,
	}, nil
}

func (oh *outputHandler) writeNodesSetup(
	initialNodes []*sharding.InitialNode,
) error {
	nodes := &sharding.NodesSetup{
		StartTime:                   0,
		RoundDuration:               oh.roundDuration,
		ConsensusGroupSize:          uint32(oh.consensusGroupSize),
		MinNodesPerShard:            uint32(oh.numOfNodesPerShard),
		MetaChainConsensusGroupSize: uint32(oh.metachainConsensusGroupSize),
		MetaChainMinNodes:           uint32(oh.numOfMetachainNodes),
		Hysteresis:                  oh.hysteresisValue,
		Adaptivity:                  oh.adaptivityValue,
		ChainID:                     oh.chainID,
		MinTransactionVersion:       uint32(oh.txVersion),
		InitialNodes:                initialNodes,
	}

	return oh.nodesSetupHandler.WriteObjectInFile(nodes)
}

// writeGenesisFile will write the provided initial accounts to the genesis file
func (oh *outputHandler) writeGenesisFile(initialAccounts []elrondData.InitialAccount) error {
	return oh.genesisHandler.WriteObjectInFile(initialAccounts)
}

// writeValidatorKeys will write the validator keys
func (oh *outputHandler) writeValidatorKeys(
	validatorKeys []*data.BlsKey,
	observerKeys []*data.BlsKey,
) error {
	keys := append(validatorKeys, observerKeys...)

	for _, key := range keys {
		pkString := oh.validatorPubKeyConverter.Encode(key.PubKeyBytes)

		err := oh.validatorKeyHandler.SaveSkToPemFile(pkString, key.PrivKeyBytes)
		if err != nil {
			return fmt.Errorf("%w for pk %s", err, pkString)
		}
	}

	return nil
}

// writeWalletKeys will write the wallet keys
func (oh *outputHandler) writeWalletKeys(walletKeys []*data.WalletKey) error {
	for _, key := range walletKeys {
		pkString := oh.walletPubKeyConverter.Encode(key.PubKeyBytes)

		err := oh.walletHandler.SaveSkToPemFile(pkString, key.PrivKeyBytes)
		if err != nil {
			return fmt.Errorf("%w for pk %s", err, pkString)
		}
	}

	return nil
}

// writeDelegatorKeys will write the delegator keys
func (oh *outputHandler) writeDelegatorKeys(delegatorKeys []*data.WalletKey) error {
	if check.IfNil(oh.delegatorsHandler) {
		log.Debug("can not write to delegator keys file as it is nil")
		return nil
	}

	for _, key := range delegatorKeys {
		pkString := oh.walletPubKeyConverter.Encode(key.PubKeyBytes)

		err := oh.delegatorsHandler.SaveSkToPemFile(pkString, key.PrivKeyBytes)
		if err != nil {
			return fmt.Errorf("%w for pk %s", err, pkString)
		}
	}

	return nil
}

// writeTxGenAccounts will write the optional txgen accounts
func (oh *outputHandler) writeTxGenAccounts(additionalKeys []*data.WalletKey) error {
	if check.IfNil(oh.txgenAccountsHandler) {
		log.Debug("can not write to tx gen accounts file as it is nil")
		return nil
	}

	txgenAccounts := make(map[uint32][]*txgenAccount)
	for _, key := range additionalKeys {
		shardID := oh.shardCoordinator.ComputeId(key.PubKeyBytes)
		pkString := oh.walletPubKeyConverter.Encode(key.PubKeyBytes)

		txGenAccount := &txgenAccount{
			PubKey:        pkString,
			PrivKey:       string(key.PrivKeyBytes),
			LastNonce:     0,
			Balance:       big.NewInt(0).Set(key.Balance),
			TokenBalance:  big.NewInt(0),
			CanReuseNonce: true,
		}
		txgenAccounts[shardID] = append(txgenAccounts[shardID], txGenAccount)
	}

	return oh.txgenAccountsHandler.WriteObjectInFile(txgenAccounts)
}

// WriteData will write the generated output in the files
func (oh *outputHandler) WriteData(generatedOutput data.GeneratorOutput) error {
	err := oh.writeNodesSetup(generatedOutput.InitialNodes)
	if err != nil {
		return err
	}

	err = oh.writeGenesisFile(generatedOutput.InitialAccounts)
	if err != nil {
		return err
	}

	err = oh.writeValidatorKeys(generatedOutput.ValidatorBlsKeys, generatedOutput.ObserverBlsKeys)
	if err != nil {
		return err
	}

	err = oh.writeWalletKeys(generatedOutput.WalletKeys)
	if err != nil {
		return err
	}

	err = oh.writeDelegatorKeys(generatedOutput.DelegatorKeys)
	if err != nil {
		return err
	}

	err = oh.writeTxGenAccounts(generatedOutput.AdditionalKeys)
	if err != nil {
		return err
	}

	return nil
}

// Close closes all inner handlers
func (oh *outputHandler) Close() {
	oh.walletHandler.Close()
	oh.validatorKeyHandler.Close()
	oh.genesisHandler.Close()
	oh.nodesSetupHandler.Close()
	if !check.IfNil(oh.txgenAccountsHandler) {
		oh.txgenAccountsHandler.Close()
	}
	if !check.IfNil(oh.delegatorsHandler) {
		oh.delegatorsHandler.Close()
	}
}

// IsInterfaceNil returns if underlying object is nil
func (oh *outputHandler) IsInterfaceNil() bool {
	return oh == nil
}
