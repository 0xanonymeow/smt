package smt

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

// HashFunction defines the interface for hash functions used in SMT
type HashFunction func(left, right []byte) []byte

// Database defines the interface for key-value storage
type Database interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Delete(key []byte) error
	Has(key []byte) (bool, error)
}

// Bytes32 represents a 32-byte hash value
type Bytes32 [32]byte

// String returns the hex string representation of Bytes32
func (b Bytes32) String() string {
	return "0x" + hex.EncodeToString(b[:])
}

// Hex returns the hex string representation without 0x prefix
func (b Bytes32) Hex() string {
	return hex.EncodeToString(b[:])
}

// IsZero checks if the Bytes32 is all zeros
func (b Bytes32) IsZero() bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

// NewBytes32FromHex creates a Bytes32 from hex string
func NewBytes32FromHex(s string) (Bytes32, error) {
	// Remove 0x prefix if present
	if len(s) >= 2 && s[0:2] == "0x" {
		s = s[2:]
	}
	
	// Ensure correct length
	if len(s) != 64 {
		return Bytes32{}, fmt.Errorf("hex string must be 64 characters, got %d", len(s))
	}
	
	bytes, err := hex.DecodeString(s)
	if err != nil { // coverage-ignore
		return Bytes32{}, err
	}
	
	var b32 Bytes32
	copy(b32[:], bytes)
	return b32, nil
}

// Proof represents a membership/non-membership proof
// 
// Field semantics:
//   - Leaf:     The computed leaf hash (Keccak256(index || value || 1))
//   - Value:    The original raw value stored at the index
//   - Index:    The tree index where the value is stored
//   - Exists:   Whether the leaf exists in the tree
//   - Enables:  Bitmask indicating which siblings are non-zero
//   - Siblings: Array of non-zero sibling hashes for proof verification
type Proof struct {
	Exists   bool     `json:"exists"`   // Whether the leaf exists
	Leaf     Bytes32  `json:"leaf"`     // Computed leaf hash (Keccak256(index || value || 1))
	Value    Bytes32  `json:"value"`    // Raw value stored at the index
	Index    *big.Int `json:"index"`    // Tree index
	Enables  *big.Int `json:"enables"`  // Sibling enable bitmask
	Siblings []Bytes32 `json:"siblings"` // Non-zero sibling hashes
}

// UpdateProof represents the proof data for an update operation
type UpdateProof struct {
	Exists   bool      `json:"exists"`
	Leaf     Bytes32   `json:"leaf"`
	Value    Bytes32   `json:"value"`
	Index    *big.Int  `json:"index"`
	Enables  *big.Int  `json:"enables"`
	Siblings []Bytes32 `json:"siblings"`
	NewLeaf  Bytes32   `json:"newLeaf"`
}

// Node represents an internal node in the tree
type Node struct {
	Left  Bytes32
	Right Bytes32
}

// IsEmpty checks if the node has no children
func (n Node) IsEmpty() bool {
	return n.Left.IsZero() && n.Right.IsZero()
}

// SerializedProof represents a proof in serialized format
type SerializedProof struct {
	Exists   uint8    `json:"exists"`
	Index    *big.Int `json:"index"`
	Leaf     string   `json:"leaf"`
	Value    string   `json:"value"`
	Enables  string   `json:"enables"`
	Siblings []string `json:"siblings"`
}

// SerializedUpdateProof represents an update proof in serialized format
type SerializedUpdateProof struct {
	Exists   uint8    `json:"exists"`
	Index    *big.Int `json:"index"`
	Leaf     string   `json:"leaf"`
	Value    string   `json:"value"`
	Enables  string   `json:"enables"`
	Siblings []string `json:"siblings"`
	NewLeaf  string   `json:"newLeaf"`
}

// KVStore represents a key-value mapping for the tree
type KVStore struct {
	kv map[string]Bytes32
}

// NewKVStore creates a new key-value store
func NewKVStore() *KVStore {
	return &KVStore{
		kv: make(map[string]Bytes32),
	}
}

// Get retrieves a value by key
func (kv *KVStore) Get(key string) (Bytes32, bool) {
	val, exists := kv.kv[key]
	return val, exists
}

// Set stores a key-value pair
func (kv *KVStore) Set(key string, value Bytes32) {
	kv.kv[key] = value
}

// Delete removes a key-value pair
func (kv *KVStore) Delete(key string) {
	delete(kv.kv, key)
}

// Has checks if a key exists
func (kv *KVStore) Has(key string) bool {
	_, exists := kv.kv[key]
	return exists
}

// All returns all key-value pairs
func (kv *KVStore) All() map[string]Bytes32 {
	result := make(map[string]Bytes32)
	for k, v := range kv.kv {
		result[k] = v
	}
	return result
}