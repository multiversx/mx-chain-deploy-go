package disabled

// NilRandomizer is a mock implementation of the IntRandomizer interface
type NilRandomizer struct {
}

// Intn returns 0
func (nr *NilRandomizer) Intn(_ int) int {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (nr *NilRandomizer) IsInterfaceNil() bool {
	return nr == nil
}
