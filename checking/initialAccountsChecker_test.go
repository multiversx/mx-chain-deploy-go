package checking

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/stretchr/testify/assert"
)

func TestInitialAccountsChecker_CheckInitialAccountsSupplyMismatch(t *testing.T) {
	t.Parallel()

	iac, _ := NewInitialAccountsChecker(big.NewInt(2500), big.NewInt(20000000))

	initialAccounts := []data.InitialAccount{
		{
			Address:      "a",
			Supply:       big.NewInt(20000000),
			Balance:      big.NewInt(8),
			StakingValue: big.NewInt(1),
			Delegation: &data.DelegationData{
				Address: "b",
				Value:   big.NewInt(19999990),
			},
		},
	}

	err := iac.CheckInitialAccounts(initialAccounts)
	assert.True(t, errors.Is(err, ErrSupplyMismatch))
}

func TestInitialAccountsChecker_CheckInitialAccountsTotalSupplyMismatch(t *testing.T) {
	t.Parallel()

	iac, _ := NewInitialAccountsChecker(big.NewInt(1), big.NewInt(20000000))

	initialAccounts := []data.InitialAccount{
		{
			Address:      "a",
			Supply:       big.NewInt(19999999),
			Balance:      big.NewInt(8),
			StakingValue: big.NewInt(1),
			Delegation: &data.DelegationData{
				Address: "b",
				Value:   big.NewInt(19999990),
			},
		},
	}

	err := iac.CheckInitialAccounts(initialAccounts)
	assert.True(t, errors.Is(err, ErrTotalSupplyMismatch))
}

func TestInitialAccountsChecker_CheckInitialAccountsStakingValueNotAMultiple(t *testing.T) {
	t.Parallel()

	iac, _ := NewInitialAccountsChecker(big.NewInt(2), big.NewInt(20000000))

	initialAccounts := []data.InitialAccount{
		{
			Address:      "a",
			Supply:       big.NewInt(20000000),
			Balance:      big.NewInt(7),
			StakingValue: big.NewInt(3),
			Delegation: &data.DelegationData{
				Address: "b",
				Value:   big.NewInt(19999990),
			},
		},
	}

	err := iac.CheckInitialAccounts(initialAccounts)
	assert.True(t, errors.Is(err, ErrStakingValueError))
}

func TestInitialAccountsChecker_CheckInitialAccountsDelegationError(t *testing.T) {
	t.Parallel()

	iac, _ := NewInitialAccountsChecker(big.NewInt(1), big.NewInt(20000000))

	initialAccounts := []data.InitialAccount{
		{
			Address:      "a",
			Supply:       big.NewInt(20000000),
			Balance:      big.NewInt(7),
			StakingValue: big.NewInt(3),
			Delegation: &data.DelegationData{
				Address: "",
				Value:   big.NewInt(19999990),
			},
		},
	}

	err := iac.CheckInitialAccounts(initialAccounts)
	assert.True(t, errors.Is(err, ErrDelegationValues))
}

func TestInitialAccountsChecker_CheckInitialAccountsNegativeSupply(t *testing.T) {
	t.Parallel()

	iac, _ := NewInitialAccountsChecker(big.NewInt(1), big.NewInt(20000000))

	initialAccounts := []data.InitialAccount{
		{
			Address:      "a",
			Supply:       big.NewInt(-20000000),
			Balance:      big.NewInt(7),
			StakingValue: big.NewInt(3),
			Delegation: &data.DelegationData{
				Address: "",
				Value:   big.NewInt(19999990),
			},
		},
	}

	err := iac.CheckInitialAccounts(initialAccounts)
	assert.True(t, errors.Is(err, ErrNegativeValue))
	assert.True(t, strings.Contains(err.Error(), "Supply"))
}

func TestInitialAccountsChecker_CheckInitialAccountsNegativeBalance(t *testing.T) {
	t.Parallel()

	iac, _ := NewInitialAccountsChecker(big.NewInt(1), big.NewInt(20000000))

	initialAccounts := []data.InitialAccount{
		{
			Address:      "a",
			Supply:       big.NewInt(20000000),
			Balance:      big.NewInt(-7),
			StakingValue: big.NewInt(3),
			Delegation: &data.DelegationData{
				Address: "",
				Value:   big.NewInt(19999990),
			},
		},
	}

	err := iac.CheckInitialAccounts(initialAccounts)
	assert.True(t, errors.Is(err, ErrNegativeValue))
	assert.True(t, strings.Contains(err.Error(), "Balance"))
}

func TestInitialAccountsChecker_CheckInitialAccountsNegativeStakingValue(t *testing.T) {
	t.Parallel()

	iac, _ := NewInitialAccountsChecker(big.NewInt(1), big.NewInt(20000000))

	initialAccounts := []data.InitialAccount{
		{
			Address:      "a",
			Supply:       big.NewInt(20000000),
			Balance:      big.NewInt(7),
			StakingValue: big.NewInt(-3),
			Delegation: &data.DelegationData{
				Address: "",
				Value:   big.NewInt(19999990),
			},
		},
	}

	err := iac.CheckInitialAccounts(initialAccounts)
	assert.True(t, errors.Is(err, ErrNegativeValue))
	assert.True(t, strings.Contains(err.Error(), "StakingValue"))
}

func TestInitialAccountsChecker_CheckInitialAccountsNegativeDelegationValue(t *testing.T) {
	t.Parallel()

	iac, _ := NewInitialAccountsChecker(big.NewInt(1), big.NewInt(20000000))

	initialAccounts := []data.InitialAccount{
		{
			Address:      "a",
			Supply:       big.NewInt(20000000),
			Balance:      big.NewInt(7),
			StakingValue: big.NewInt(3),
			Delegation: &data.DelegationData{
				Address: "",
				Value:   big.NewInt(-19999990),
			},
		},
	}

	err := iac.CheckInitialAccounts(initialAccounts)
	assert.True(t, errors.Is(err, ErrNegativeValue))
	assert.True(t, strings.Contains(err.Error(), "Delegation.Value"))
}

func TestInitialAccountsChecker_CheckInitialAccountsShouldWork(t *testing.T) {
	t.Parallel()

	iac, _ := NewInitialAccountsChecker(big.NewInt(1), big.NewInt(20000000))

	initialAccounts := []data.InitialAccount{
		{
			Address:      "a",
			Supply:       big.NewInt(10000000),
			Balance:      big.NewInt(10),
			StakingValue: big.NewInt(9),
			Delegation: &data.DelegationData{
				Address: "b",
				Value:   big.NewInt(9999981),
			},
		},
		{
			Address:      "b",
			Supply:       big.NewInt(10000000),
			Balance:      big.NewInt(9999981),
			StakingValue: big.NewInt(19),
			Delegation: &data.DelegationData{
				Address: "",
				Value:   big.NewInt(0),
			},
		},
	}

	err := iac.CheckInitialAccounts(initialAccounts)
	assert.Nil(t, err)
}
