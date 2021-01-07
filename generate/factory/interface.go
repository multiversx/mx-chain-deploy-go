package factory

import "github.com/ElrondNetwork/elrond-deploy-go/data"

// DataGenerator represents a structure that can generate genesis data
type DataGenerator interface {
	Generate() (*data.GeneratorOutput, error)
	IsInterfaceNil() bool
}
