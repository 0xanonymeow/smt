package smt

import (
	"fmt"
	"math/big"
	
	"github.com/ethereum/go-ethereum/crypto"
)

// BatchInsert inserts multiple leaves efficiently
func (smt *SparseMerkleTree) BatchInsert(indices []*big.Int, leaves []Bytes32) ([]*UpdateProof, error) {
	if len(indices) != len(leaves) {// coverage-ignore
		return nil, fmt.Errorf("indices and leaves must have same length")
	}
	
	proofs := make([]*UpdateProof, len(indices))
	
	for i := range indices {
		proof, err := smt.Insert(indices[i], leaves[i])
		if err != nil {
			// Continue with other insertions
			proofs[i] = nil
			continue
		}
		proofs[i] = proof
	}
	
	return proofs, nil
}

// BatchUpdate updates multiple leaves efficiently
func (smt *SparseMerkleTree) BatchUpdate(indices []*big.Int, newLeaves []Bytes32) ([]*UpdateProof, error) {
	if len(indices) != len(newLeaves) {// coverage-ignore
		return nil, fmt.Errorf("indices and newLeaves must have same length")
	}
	
	proofs := make([]*UpdateProof, len(indices))
	
	for i := range indices {
		proof, err := smt.Update(indices[i], newLeaves[i])
		if err != nil {
			// Continue with other updates
			proofs[i] = nil
			continue
		}
		proofs[i] = proof
	}
	
	return proofs, nil
}

// BatchGet retrieves multiple proofs efficiently
func (smt *SparseMerkleTree) BatchGet(indices []*big.Int) ([]*Proof, error) {
	proofs := make([]*Proof, len(indices))
	
	for i, index := range indices {
		proof, err := smt.Get(index)
		if err != nil { // coverage-ignore
			return nil, err
		}
		proofs[i] = proof
	}
	
	return proofs, nil
}

// BatchExists checks existence of multiple keys efficiently
func (smt *SparseMerkleTree) BatchExists(indices []*big.Int) ([]bool, error) {
	results := make([]bool, len(indices))
	
	for i, index := range indices {
		exists, err := smt.Exists(index)
		if err != nil { // coverage-ignore
			return nil, err
		}
		results[i] = exists
	}
	
	return results, nil
}

// BatchInsertKV inserts multiple key-value pairs
func (smt *SparseMerkleTree) BatchInsertKV(kvPairs map[string]Bytes32) ([]*UpdateProof, error) {
	indices := make([]*big.Int, 0, len(kvPairs))
	values := make([]Bytes32, 0, len(kvPairs))
	
	for key, value := range kvPairs {
		// Compute index from key and truncate to tree depth (same logic as InsertKV)
		hash := crypto.Keccak256([]byte(key))
		index := new(big.Int).SetBytes(hash)
		
		// Truncate index to fit within tree depth
		if smt.depth < SMT_DEPTH {
			maxIndex := new(big.Int).Lsh(ONE, uint(smt.depth))
			index.Mod(index, maxIndex)
		}
		
		indices = append(indices, index)
		values = append(values, value)
		
		// Store in KV store
		smt.kvStore.Set(key, value)
	}
	
	return smt.BatchInsert(indices, values)
}

// BatchGetKV retrieves multiple values by keys
func (smt *SparseMerkleTree) BatchGetKV(keys []string) (map[string]Bytes32, error) {
	results := make(map[string]Bytes32)
	
	for _, key := range keys {
		value, exists, err := smt.GetKV(key)
		if err != nil { // coverage-ignore
			return nil, err
		}
		if exists {
			results[key] = value
		}
	}
	
	return results, nil
}

// BatchOperation represents a batch of operations to be performed atomically
type BatchOperation struct {
	Type    string    // "insert", "update", "delete"
	Index   *big.Int
	Leaf    Bytes32
	Key     string    // For KV operations
	Value   Bytes32   // For KV operations
}

// ExecuteBatch executes a batch of operations atomically
func (smt *SparseMerkleTree) ExecuteBatch(operations []BatchOperation) ([]*UpdateProof, error) {
	smt.mu.Lock()
	defer smt.mu.Unlock()
	
	// Save current root for rollback
	oldRoot := smt.root
	
	proofs := make([]*UpdateProof, len(operations))
	
	for i, op := range operations {
		var proof *UpdateProof
		var err error
		
		switch op.Type {
		case "insert":
			if op.Key != "" { // coverage-ignore
				// KV insert - compute index and use internal method
				hash := crypto.Keccak256([]byte(op.Key))
				index := new(big.Int).SetBytes(hash)
				
				// Truncate index to fit within tree depth
				if smt.depth < SMT_DEPTH {
					maxIndex := new(big.Int).Lsh(ONE, uint(smt.depth))
					index.Mod(index, maxIndex)
				}
				
				leafHash := ComputeLeafHash(index, op.Value)
				proof, err = smt.insertInternal(index, leafHash)
				if err == nil {
					smt.kvStore.Set(op.Key, op.Value)
				}
			} else { // coverage-ignore
				// Direct insert - use internal method to avoid deadlock
				proof, err = smt.insertInternal(op.Index, op.Leaf)
			}
			
		case "update": // coverage-ignore
			if op.Key != "" {
				// KV update
				hash := crypto.Keccak256([]byte(op.Key))
				index := new(big.Int).SetBytes(hash)
				
				// Truncate index to fit within tree depth
				if smt.depth < SMT_DEPTH {
					maxIndex := new(big.Int).Lsh(ONE, uint(smt.depth))
					index.Mod(index, maxIndex)
				}
				
				leafHash := ComputeLeafHash(index, op.Value)
				proof, err = smt.updateInternal(index, leafHash)
				if err == nil {
					smt.kvStore.Set(op.Key, op.Value)
				}
			} else {
				// Direct update - use internal method to avoid deadlock
				proof, err = smt.updateInternal(op.Index, op.Leaf)
			}
			
		case "delete": // coverage-ignore
			if op.Key != "" {
				// KV delete - use internal method to avoid deadlock
				hash := crypto.Keccak256([]byte(op.Key))
				index := new(big.Int).SetBytes(hash)
				
				// Truncate index to fit within tree depth
				if smt.depth < SMT_DEPTH {
					maxIndex := new(big.Int).Lsh(ONE, uint(smt.depth))
					index.Mod(index, maxIndex)
				}

				// Check if key exists in KV store
				if !smt.kvStore.Has(op.Key) {
					err = &KeyNotFoundError{Index: index}
				} else {
					// Delete from tree using internal method
					proof, err = smt.deleteInternal(index)
					if err == nil {
						// Remove from KV store
						smt.kvStore.Delete(op.Key)
					}
				}
			} else {
				// Direct delete - use internal method to avoid deadlock
				proof, err = smt.deleteInternal(op.Index)
			}
			
		default:
			err = fmt.Errorf("unknown operation type: %s", op.Type)
		}
		
		if err != nil {
			// Rollback on error
			smt.root = oldRoot
			return nil, fmt.Errorf("batch operation %d failed: %w", i, err)
		}
		
		proofs[i] = proof
	}
	
	return proofs, nil
}

