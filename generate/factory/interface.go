package factory

import "github.com/multiversx/mx-chain-deploy-go/data"

// DataGenerator represents a structure that can generate genesis data
type DataGenerator interface {
	Generate() (*data.GeneratorOutput, error)
	IsInterfaceNil() bool
}
