package core

import (
	"encoding/hex"
	"math/big"

	elrondCore "github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
)

// ConvertToBigInt will try to convert the provided string to its big int corresponding value
func ConvertToBigInt(value string) (*big.Int, error) {
	valueNumber, isNumber := big.NewInt(0).SetString(value, 10)
	if !isNumber {
		return nil, ErrStringIsNotANumber
	}

	if valueNumber.Cmp(big.NewInt(0)) < 0 {
		return nil, ErrNegativeValue
	}

	return valueNumber, nil
}

// GenerateSCAddress will generate the resulting SC address from the provided public key string and nonce
func GenerateSCAddress(
	pkString string,
	nonce uint64,
	vmType string,
	converter elrondCore.PubkeyConverter,
) (string, error) {

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

	scResultingAddressBytes, err := blockchainHook.NewAddress(pk, nonce, vmTypeBytes)
	if err != nil {
		return "", err
	}

	return converter.Encode(scResultingAddressBytes), nil
}

func generateBlockchainHook(converter elrondCore.PubkeyConverter) (process.BlockChainHookHandler, error) {
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
