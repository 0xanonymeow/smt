package smt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
)

// LeafData represents the data stored for a leaf
type LeafData struct {
	Index *big.Int
	Value Bytes32
}

// Database key prefixes
const (
	NodePrefix = "n:"
	LeafPrefix = "l:"
	LeafIndexPrefix = "i:"
)

// getNode retrieves a node from the database
func (smt *SparseMerkleTree) getNode(hash Bytes32) (*Node, error) {
	key := []byte(NodePrefix + hex.EncodeToString(hash[:]))
	data, err := smt.db.Get(key)
	if err != nil {// coverage-ignore
		return nil, err
	}
	
	if len(data) == 0 {
		return &Node{}, nil
	}
	
	if len(data) != 64 { // coverage-ignore
		return nil, fmt.Errorf("invalid node data length: expected 64, got %d", len(data))
	}
	
	node := &Node{}
	copy(node.Left[:], data[0:32])
	copy(node.Right[:], data[32:64])
	
	return node, nil
}

// setNode stores a node in the database
func (smt *SparseMerkleTree) setNode(hash Bytes32, node *Node) error {
	key := []byte(NodePrefix + hex.EncodeToString(hash[:]))
	data := append(node.Left[:], node.Right[:]...)
	return smt.db.Set(key, data)
}

// deleteNode removes a node from the database
func (smt *SparseMerkleTree) deleteNode(hash Bytes32) error {
	key := []byte(NodePrefix + hex.EncodeToString(hash[:]))
	return smt.db.Delete(key)
}

// getLeaf retrieves a leaf from the database
func (smt *SparseMerkleTree) getLeaf(hash Bytes32) (*LeafData, error) {
	key := []byte(LeafPrefix + hex.EncodeToString(hash[:]))
	data, err := smt.db.Get(key)
	if err != nil {// coverage-ignore
		return nil, err
	}
	
	if len(data) == 0 {
		return nil, nil
	}
	
	if len(data) < 32 { // coverage-ignore
		return nil, fmt.Errorf("invalid leaf data length: expected at least 32, got %d", len(data))
	}
	
	// Extract value (first 32 bytes)
	var value Bytes32
	copy(value[:], data[0:32])
	
	// Extract index (remaining bytes)
	index := new(big.Int).SetBytes(data[32:])
	
	return &LeafData{
		Index: index,
		Value: value,
	}, nil
}

// setLeaf stores a leaf in the database
func (smt *SparseMerkleTree) setLeaf(hash Bytes32, leaf *LeafData) error {
	// Store leaf data
	key := []byte(LeafPrefix + hex.EncodeToString(hash[:]))
	indexBytes := leaf.Index.Bytes()
	data := append(leaf.Value[:], indexBytes...)
	
	if err := smt.db.Set(key, data); err != nil { // coverage-ignore
		return err
	}
	
	// Store index mapping
	indexKey := []byte(LeafIndexPrefix + hex.EncodeToString(indexBytes))
	return smt.db.Set(indexKey, hash[:])
}

// deleteLeaf removes a leaf from the database
func (smt *SparseMerkleTree) deleteLeaf(hash Bytes32) error {
	// Get leaf data to find index
	leaf, err := smt.getLeaf(hash)
	if err != nil {// coverage-ignore
		return err
	}
	
	if leaf != nil {
		// Delete index mapping
		indexBytes := leaf.Index.Bytes()
		indexKey := []byte(LeafIndexPrefix + hex.EncodeToString(indexBytes))
		if err := smt.db.Delete(indexKey); err != nil { // coverage-ignore
			return err
		}
	}
	
	// Delete leaf data
	key := []byte(LeafPrefix + hex.EncodeToString(hash[:]))
	return smt.db.Delete(key)
}

// getLeafByIndex retrieves a leaf hash by its index
func (smt *SparseMerkleTree) getLeafByIndex(index *big.Int) (Bytes32, error) {
	indexBytes := index.Bytes()
	indexKey := []byte(LeafIndexPrefix + hex.EncodeToString(indexBytes))
	
	data, err := smt.db.Get(indexKey)
	if err != nil { // coverage-ignore
		return Bytes32{}, err
	}
	
	if len(data) != 32 { // coverage-ignore
		return Bytes32{}, nil
	}
	
	var hash Bytes32
	copy(hash[:], data)
	return hash, nil
}

// InMemoryDatabase is a simple in-memory database implementation
type InMemoryDatabase struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewInMemoryDatabase creates a new in-memory database
func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{
		data: make(map[string][]byte),
	}
}

// Get retrieves a value by key
func (db *InMemoryDatabase) Get(key []byte) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	value, exists := db.data[string(key)]
	if !exists {
		return nil, nil
	}
	
	// Return a copy to prevent external modifications
	result := make([]byte, len(value))
	copy(result, value)
	return result, nil
}

// Set stores a key-value pair
func (db *InMemoryDatabase) Set(key []byte, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	// Store a copy to prevent external modifications
	storedValue := make([]byte, len(value))
	copy(storedValue, value)
	db.data[string(key)] = storedValue
	return nil
}

// Delete removes a key-value pair
func (db *InMemoryDatabase) Delete(key []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	delete(db.data, string(key))
	return nil
}

// Has checks if a key exists
func (db *InMemoryDatabase) Has(key []byte) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	_, exists := db.data[string(key)]
	return exists, nil
}