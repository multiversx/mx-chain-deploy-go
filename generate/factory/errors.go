package factory

import "errors"

// ErrUnknownGenerationType signals that an unknown data generation type was provided
var ErrUnknownGenerationType = errors.New("unknown data generation type")
