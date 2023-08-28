package core

import (
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("core")

type fileHandler struct {
	*os.File
}

// NewFileHandler will try to open a new file in the provided output directory with the provided filename
func NewFileHandler(outputDirectory string, fileName string) (*fileHandler, error) {
	filePath := filepath.Join(outputDirectory, fileName)
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	return &fileHandler{
		File: f,
	}, nil
}

// WriteObjectInFile will try to write the provided object in the file after it has been marshaled
// in json format
func (fh *fileHandler) WriteObjectInFile(data interface{}) error {
	buff, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	_, err = fh.Write(buff)

	return err
}

// SaveSkToPemFile saves secret key bytes in the file
func (fh *fileHandler) SaveSkToPemFile(identifier string, skBytes []byte) error {
	blk := pem.Block{
		Type:  "PRIVATE KEY for " + identifier,
		Bytes: []byte(hex.EncodeToString(skBytes)),
	}

	return pem.Encode(fh, &blk)
}

// Close will try to close the file
func (fh *fileHandler) Close() {
	err := fh.File.Close()
	log.LogIfError(err)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fh *fileHandler) IsInterfaceNil() bool {
	return fh == nil
}
