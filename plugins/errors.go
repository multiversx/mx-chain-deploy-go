package plugins

import "errors"

// ErrNilFileHandler signals that the provided file handler is nil
var ErrNilFileHandler = errors.New("nil file handler")

// ErrNilPubKeyConverter signals that a nil pub key converter was provided
var ErrNilPubKeyConverter = errors.New("nil pub key converter")

// ErrNilShardCoordinator signals that a nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")
