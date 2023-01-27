package check

import (
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-go/genesis/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("checking")
var zero = big.NewInt(0)

type initialAccountsChecker struct {
	nodePrice   *big.Int
	totalSupply *big.Int
}

// NewInitialAccountsChecker creates a new initial accounts checker
func NewInitialAccountsChecker(nodePrice *big.Int, totalSupply *big.Int) (*initialAccountsChecker, error) {
	if nodePrice == nil {
		return nil, fmt.Errorf("%w for nodePrice", ErrNilValue)
	}
	if totalSupply == nil {
		return nil, fmt.Errorf("%w for totalSupply", ErrNilValue)
	}
	if nodePrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w for nodePrice", ErrZeroOrNegative)
	}
	if totalSupply.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w for totalSupply", ErrZeroOrNegative)
	}

	return &initialAccountsChecker{
		nodePrice:   nodePrice,
		totalSupply: totalSupply,
	}, nil
}

// CheckInitialAccounts will check the provided initial accounts
func (iac *initialAccountsChecker) CheckInitialAccounts(initialAccounts []data.InitialAccount) error {
	if len(initialAccounts) == 0 {
		return ErrEmptyInitialAccounts
	}

	totalSupply := big.NewInt(0)
	totalStaked := big.NewInt(0)
	totalBalance := big.NewInt(0)
	totalDelegated := big.NewInt(0)
	for _, ia := range initialAccounts {
		if ia.StakingValue.Cmp(zero) < 0 {
			return fmt.Errorf("%w for address %s, field StakingValue", ErrNegativeValue, ia.Address)
		}
		if ia.Balance.Cmp(zero) < 0 {
			return fmt.Errorf("%w for address %s, field Balance", ErrNegativeValue, ia.Address)
		}
		if ia.Supply.Cmp(zero) < 0 {
			return fmt.Errorf("%w for address %s, field Supply", ErrNegativeValue, ia.Address)
		}
		if ia.Delegation.Value.Cmp(zero) < 0 {
			return fmt.Errorf("%w for address %s, field Delegation.Value", ErrNegativeValue, ia.Address)
		}

		supply := big.NewInt(0)
		supply.Add(supply, ia.Balance)
		supply.Add(supply, ia.StakingValue)
		supply.Add(supply, ia.Delegation.Value)
		if supply.Cmp(ia.Supply) != 0 {
			return fmt.Errorf("%w for address %s", ErrSupplyMismatch, ia.Address)
		}

		remainder := big.NewInt(0).Set(ia.StakingValue)
		remainder.Mod(remainder, iac.nodePrice)
		if remainder.Cmp(zero) != 0 {
			return ErrStakingValueError
		}

		if ia.Delegation.Value.Cmp(zero) > 0 {
			if len(ia.Delegation.Address) == 0 {
				return fmt.Errorf("%w for address %s", ErrDelegationValues, ia.Address)
			}
		}

		totalSupply.Add(totalSupply, supply)
		totalBalance.Add(totalBalance, ia.Balance)
		totalStaked.Add(totalStaked, ia.StakingValue)
		totalDelegated.Add(totalDelegated, ia.Delegation.Value)
	}

	if totalSupply.Cmp(iac.totalSupply) != 0 {
		return fmt.Errorf("%w computed: %s, provided: %d", ErrTotalSupplyMismatch, totalSupply, iac.totalSupply)
	}

	log.Info("checked values",
		"total supply", totalSupply.String(),
		"total staked", totalStaked.String(),
		"total balance", totalBalance.String(),
		"total delegated", totalDelegated.String(),
	)

	return nil
}
