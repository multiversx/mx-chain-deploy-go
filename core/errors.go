package core

import "errors"

// ErrStringIsNotANumber signals that the provided string is not a number
var ErrStringIsNotANumber = errors.New("string is not a number")

// ErrNegativeValue signals that the provided value is negative
var ErrNegativeValue = errors.New("negative value")
