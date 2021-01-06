package io

// FileHandler describes the file handling capabilities
type FileHandler interface {
	WriteObjectInFile(data interface{}) error
	SaveSkToPemFile(identifier string, skBytes []byte) error
	Close()
	IsInterfaceNil() bool
}
