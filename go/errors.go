package smt

import (
	"fmt"
	"math/big"
)

// Custom errors for SMT operations
var (
	// ErrInvalidTreeDepth is returned when tree depth is invalid
	ErrInvalidTreeDepth = fmt.Errorf("tree depth must be between 1 and 256")

	// ErrOutOfRange is returned when index is out of range for the tree depth
	ErrOutOfRange = fmt.Errorf("index out of range for tree depth")

	// ErrKeyNotFound is returned when trying to update a non-existent key
	ErrKeyNotFound = fmt.Errorf("key not found")

	// ErrKeyExists is returned when trying to insert an existing key
	ErrKeyExists = fmt.Errorf("key already exists")

	// ErrInvalidProof is returned when proof verification fails
	ErrInvalidProof = fmt.Errorf("invalid proof")

	// ErrNilDatabase is returned when database is nil
	ErrNilDatabase = fmt.Errorf("database cannot be nil")
)

// InvalidTreeDepthError represents an error for invalid tree depth
type InvalidTreeDepthError struct {
	Depth uint16
}

func (e InvalidTreeDepthError) Error() string {
	return fmt.Sprintf("invalid tree depth: %d (must be between 1 and 256)", e.Depth)
}

// OutOfRangeError represents an error for out of range index
type OutOfRangeError struct {
	Index     *big.Int
	TreeDepth uint16
}

func (e OutOfRangeError) Error() string {
	maxIndex := new(big.Int).Lsh(big.NewInt(1), uint(e.TreeDepth))
	return fmt.Sprintf("index %s out of range for tree depth %d (max: %s)",
		e.Index.String(), e.TreeDepth, maxIndex.String())
}

// KeyNotFoundError represents an error when key is not found
type KeyNotFoundError struct {
	Index *big.Int
}

func (e KeyNotFoundError) Error() string {
	return fmt.Sprintf("key not found at index: %s", e.Index.String())
}

// KeyExistsError represents an error when key already exists
type KeyExistsError struct {
	Index *big.Int
}

func (e KeyExistsError) Error() string {
	return fmt.Sprintf("key already exists at index: %s", e.Index.String())
}

// IsKeyExistsError checks if an error is a KeyExistsError
func IsKeyExistsError(err error) bool {
	_, ok := err.(*KeyExistsError)
	return ok
}
