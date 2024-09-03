package core

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToPositiveBigInt_NotANumber(t *testing.T) {
	t.Parallel()

	val, err := ConvertToPositiveBigInt("not a number")
	assert.Nil(t, val)
	assert.True(t, errors.Is(err, ErrStringIsNotANumber))
}

func TestConvertToPositiveBigInt_NegativeNumber(t *testing.T) {
	t.Parallel()

	val, err := ConvertToPositiveBigInt("-1")
	assert.Nil(t, val)
	assert.True(t, errors.Is(err, ErrNegativeValue))
}

func TestConvertToPositiveBigInt_ShouldWork(t *testing.T) {
	t.Parallel()

	val, err := ConvertToPositiveBigInt("1")
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(1), val)
}

func TestGenerateSCAddress_ShouldWork(t *testing.T) {
	t.Parallel()

	pkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	scAddress, err := GenerateSCAddress(
		"erd1ulhw20j7jvgfgak5p05kv667k5k9f320sgef5ayxkt9784ql0zssrzyhjp",
		0,
		"0500",
		pkConv,
	)

	require.Nil(t, err)
	assert.Equal(t, scAddress, "erd1qqqqqqqqqqqqqpgqvyvaeu6mnr9fq25kt0gyaymtn6zgjmp80zssuqmp6l")
}
