package checking

import "errors"

// ErrNilValue signals that a nil value was provided
var ErrNilValue = errors.New("nil value")

// ErrZeroOrNegative signals that a zero or negative value was provided
var ErrZeroOrNegative = errors.New("zero or negative value")

// ErrEmptyInitialAccounts signals that an empty initial accounts list
var ErrEmptyInitialAccounts = errors.New("empty initial accounts list")

// ErrSupplyMismatch signals that the initial account's supply mismatches
var ErrSupplyMismatch = errors.New("supply mismatch")

// ErrStakingValueError signals the provided staking value is not a multiple of the node's price
var ErrStakingValueError = errors.New("staking value error")

// ErrNegativeValue signals that the provided value is negative
var ErrNegativeValue = errors.New("negative value")

// ErrDelegationValues signals that the provided delegation fields were improperly set
var ErrDelegationValues = errors.New("delegation values error")

// ErrTotalSupplyMismatch signals that the total supply mismatches
var ErrTotalSupplyMismatch = errors.New("total supply mismatch")
