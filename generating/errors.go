package generating

import "errors"

// ErrNilKeyGenerator signals that a nil key generator was provided
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrNilRandomizer signals that a nil randomizer was provided
var ErrNilRandomizer = errors.New("nil int randomizer")

// ErrInvalidValue signals that an improper value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilNodePrice signals that a nil node price was provided
var ErrNilNodePrice = errors.New("nil node price")

// ErrTotalSupplyTooSmall signals that the provided total supply is too small
var ErrTotalSupplyTooSmall = errors.New("total supply is too small")

// ErrInvalidNumberOfWalletKeys signals that the provided number of wallet keys is invalid
var ErrInvalidNumberOfWalletKeys = errors.New("invalid number of wallet keys")

// ErrNilPubKeyConverter signals that a nil pub key converter was provided
var ErrNilPubKeyConverter = errors.New("nil pub key converter")
