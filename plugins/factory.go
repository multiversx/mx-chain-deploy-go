package plugins

import (
	"fmt"

	mxCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-deploy-go/core"
	"github.com/multiversx/mx-chain-go/sharding"
)

const walletKeyFileName = "walletKey.pem"
const validatorKeyFileName = "validatorKey.pem"
const genesisFilename = "genesis.json"
const nodesSetupFilename = "nodesSetup.json"
const txgenAccountsFileName = "accounts.json"
const delegatorsFileName = "delegators.pem"

// CreateOutputHandlerArgument will create an output handler argument
func CreateOutputHandlerArgument(
	outputDirectory string,
	validatorPubKeyConverter mxCore.PubkeyConverter,
	walletPubKeyConverter mxCore.PubkeyConverter,
	shardCoordinator sharding.Coordinator,
	shouldOutputTxgenAccountsFile bool,
	shouldOutputDelegatorsFile bool,
) (ArgOutputHandler, error) {
	aoh := ArgOutputHandler{
		ValidatorPubKeyConverter: validatorPubKeyConverter,
		WalletPubKeyConverter:    walletPubKeyConverter,
		ShardCoordinator:         shardCoordinator,
	}

	var err error
	aoh.WalletHandler, err = core.NewFileHandler(outputDirectory, walletKeyFileName)
	if err != nil {
		return ArgOutputHandler{}, fmt.Errorf("%w for WalletHandler", err)
	}
	aoh.NodesSetupHandler, err = core.NewFileHandler(outputDirectory, nodesSetupFilename)
	if err != nil {
		return ArgOutputHandler{}, fmt.Errorf("%w for NodesSetupHandler", err)
	}
	aoh.GenesisHandler, err = core.NewFileHandler(outputDirectory, genesisFilename)
	if err != nil {
		return ArgOutputHandler{}, fmt.Errorf("%w for GenesisHandler", err)
	}
	aoh.ValidatorKeyHandler, err = core.NewFileHandler(outputDirectory, validatorKeyFileName)
	if err != nil {
		return ArgOutputHandler{}, fmt.Errorf("%w for ValidatorKeyHandler", err)
	}

	if shouldOutputTxgenAccountsFile {
		aoh.TxgenAccountsHandler, err = core.NewFileHandler(outputDirectory, txgenAccountsFileName)
		if err != nil {
			return ArgOutputHandler{}, fmt.Errorf("%w for TxgenAccountsHandler", err)
		}
	}
	if shouldOutputDelegatorsFile {
		aoh.DelegatorsHandler, err = core.NewFileHandler(outputDirectory, delegatorsFileName)
		if err != nil {
			return ArgOutputHandler{}, fmt.Errorf("%w for DelegatorsHandler", err)
		}
	}

	return aoh, nil
}
