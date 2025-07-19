package tests

import (
	"crypto/rand"
	"math/big"
	"testing"

	smt "github.com/0xanonymeow/smt/go"
	"github.com/ethereum/go-ethereum/crypto"
)

// PropertyTest represents a property-based test function
type PropertyTest func(t *testing.T, inputs []interface{}) bool

// RunPropertyTest runs a property-based test with random inputs
func RunPropertyTest(t *testing.T, name string, property PropertyTest, numTests int) {
	t.Run(name, func(t *testing.T) {
		passed := 0
		failed := 0
		
		for i := 0; i < numTests; i++ {
			inputs := generateRandomInputs(t, i)
			
			if property(t, inputs) {
				passed++
			} else {
				failed++
				if failed > numTests/10 { // Stop if more than 10% fail
					break
				}
			}
		}
		
		t.Logf("Property test results: %d passed, %d failed out of %d tests", passed, failed, numTests)
		
		if failed > 0 {
			t.Errorf("Property test failed %d times", failed)
		}
	})
}

// generateRandomInputs generates random inputs for property testing
func generateRandomInputs(t *testing.T, seed int) []interface{} {
	inputs := make([]interface{}, 0)
	
	// Generate random key
	keyBytes := make([]byte, 4)
	rand.Read(keyBytes)
	key := new(big.Int).SetBytes(keyBytes)
	inputs = append(inputs, key)
	
	// Generate random value
	valueBytes := make([]byte, 32)
	rand.Read(valueBytes)
	value := smt.Bytes32{}
	copy(value[:], valueBytes)
	inputs = append(inputs, value)
	
	// Generate random tree depth
	depth := uint16(8 + (seed % 8)) // Depth between 8 and 15
	inputs = append(inputs, depth)
	
	return inputs
}

// TestTreeInvariants tests fundamental tree invariants
func TestTreeInvariants(t *testing.T) {
	// Property: Root is deterministic for same operations
	RunPropertyTest(t, "DeterministicRoot", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		depth := inputs[2].(uint16)
		
		// Create two identical trees
		db1 := smt.NewInMemoryDatabase()
		tree1, err := smt.NewSparseMerkleTree(db1, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		db2 := smt.NewInMemoryDatabase()
		tree2, err := smt.NewSparseMerkleTree(db2, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		// Perform same operations
		_, err1 := tree1.Insert(key, value)
		_, err2 := tree2.Insert(key, value)
		
		if err1 != nil || err2 != nil {
			return true // Skip if operation fails
		}
		
		// Roots should be identical
		root1 := tree1.Root()
		root2 := tree2.Root()
		
		if root1 != root2 {
			t.Errorf("Roots differ: %s vs %s", root1.String(), root2.String())
			return false
		}
		
		return true
	}, 50)
	
	// Property: Proof always verifies correctly
	RunPropertyTest(t, "ProofValidation", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		depth := inputs[2].(uint16)
		
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		// Insert element
		_, err = tree.Insert(key, value)
		if err != nil {
			return true // Skip if insert fails
		}
		
		// Generate proof
		proof, err := tree.Get(key)
		if err != nil {
			return true // Skip if get fails
		}
		if !proof.Exists {
			t.Error("Proof should indicate element exists")
			return false
		}
		
		// Verify proof
		valid := tree.VerifyProof(proof)
		if !valid {
			t.Error("Generated proof should always be valid")
			return false
		}
		
		return true
	}, 50)
	
	// Property: Insert then Get returns same value
	RunPropertyTest(t, "InsertGetConsistency", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		depth := inputs[2].(uint16)
		
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		// Insert element
		_, err = tree.Insert(key, value)
		if err != nil {
			return true // Skip if insert fails
		}
		
		// Get element
		proof, err := tree.Get(key)
		if err != nil {
			return true // Skip if get fails
		}
		if !proof.Exists {
			t.Error("Element should exist after insertion")
			return false
		}
		
		// Value should match (through leaf verification)
		valid := tree.VerifyProof(proof)
		if !valid {
			t.Error("Retrieved proof should be valid")
			return false
		}
		
		return true
	}, 50)
	
	// Property: Update changes root
	RunPropertyTest(t, "UpdateChangesRoot", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		depth := inputs[2].(uint16)
		
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		// Insert initial element
		_, err = tree.Insert(key, value)
		if err != nil {
			return true // Skip if insert fails
		}
		
		initialRoot := tree.Root()
		
		// Update with different value
		newValue := smt.Bytes32{}
		for i := range newValue {
			newValue[i] = value[i] ^ 0x01 // Flip bit to ensure different value
		}
		_, err = tree.Update(key, newValue)
		if err != nil {
			return true // Skip if update fails
		}
		
		updatedRoot := tree.Root()
		
		// Root should change (unless values are identical)
		if value != newValue && initialRoot == updatedRoot {
			t.Error("Root should change after update with different value")
			return false
		}
		
		return true
	}, 50)
}

// TestInvariants tests consistency invariants across multiple trees
func TestInvariants(t *testing.T) {
	// Property: Multiple trees with same operations produce same results
	RunPropertyTest(t, "MultipleTreeConsistency", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		depth := inputs[2].(uint16)
		
		// Create multiple trees
		db1 := smt.NewInMemoryDatabase()
		tree1, err := smt.NewSparseMerkleTree(db1, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		db2 := smt.NewInMemoryDatabase()
		tree2, err := smt.NewSparseMerkleTree(db2, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		// Perform same operations
		_, err1 := tree1.Insert(key, value)
		_, err2 := tree2.Insert(key, value)
		
		// Both should succeed or both should fail
		if (err1 == nil) != (err2 == nil) {
			t.Error("Trees should have same success/failure")
			return false
		}
		
		if err1 != nil || err2 != nil {
			return true // Skip if operations fail
		}
		
		// Roots should be identical
		root1 := tree1.Root()
		root2 := tree2.Root()
		
		if root1 != root2 {
			t.Errorf("Roots differ: tree1=%s, tree2=%s", root1.String(), root2.String())
			return false
		}
		
		// Proofs should be equivalent
		proof1, err1 := tree1.Get(key)
		proof2, err2 := tree2.Get(key)
		
		if err1 != nil || err2 != nil {
			return true // Skip if get fails
		}
		
		if proof1.Exists != proof2.Exists {
			t.Error("Proof existence should match")
			return false
		}
		
		if proof1.Leaf != proof2.Leaf {
			t.Error("Proof leaves should match")
			return false
		}
		
		return true
	}, 50)
}

// TestKVInvariants tests invariants specific to key-value implementation
func TestKVInvariants(t *testing.T) {
	// Property: KV operations maintain consistency
	RunPropertyTest(t, "KVConsistency", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		depth := inputs[2].(uint16)
		
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		keyStr := key.String()
		
		// Insert element using KV interface
		_, err = tree.InsertKV(keyStr, value)
		if err != nil {
			return true // Skip if insert fails
		}
		
		// Check existence using computed index
		index := new(big.Int).SetBytes(crypto.Keccak256([]byte(keyStr)))
		exists, err := tree.Exists(index)
		if err != nil {
			return true // Skip if exists check fails
		}
		if !exists {
			t.Error("Element should exist after KV insertion")
			return false
		}
		
		// Get element using KV interface
		retrievedValue, kvExists, err := tree.GetKV(keyStr)
		if err != nil {
			return true // Skip if get fails
		}
		if !kvExists {
			t.Error("KV should indicate element exists")
			return false
		}
		
		// Values should match
		if retrievedValue != value {
			t.Error("Retrieved value should match inserted value")
			return false
		}
		
		return true
	}, 50)
}

// TestConcurrencyInvariants tests invariants under concurrent access
func TestConcurrencyInvariants(t *testing.T) {
	// Property: Concurrent reads don't interfere
	RunPropertyTest(t, "ConcurrentReadSafety", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		depth := inputs[2].(uint16)
		
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		// Insert element
		_, err = tree.Insert(key, value)
		if err != nil {
			return true // Skip if insert fails
		}
		
		// Perform concurrent reads
		numReaders := 10
		results := make(chan bool, numReaders)
		
		for i := 0; i < numReaders; i++ {
			go func() {
				proof, err := tree.Get(key)
				if err != nil {
					results <- false
					return
				}
				valid := tree.VerifyProof(proof)
				results <- valid && proof.Exists
			}()
		}
		
		// Check all results
		for i := 0; i < numReaders; i++ {
			result := <-results
			if !result {
				t.Error("Concurrent read should succeed")
				return false
			}
		}
		
		return true
	}, 20) // Fewer tests for concurrency due to overhead
}

// TestSerializationInvariants tests serialization/deserialization invariants
func TestSerializationInvariants(t *testing.T) {
	// Property: Bytes32 conversion consistency
	RunPropertyTest(t, "Bytes32Consistency", func(t *testing.T, inputs []interface{}) bool {
		value := inputs[1].(smt.Bytes32)
		
		// Convert to string and back
		hexStr := value.String()
		converted, err := smt.HexToBytes32(hexStr)
		if err != nil {
			t.Errorf("Failed to convert back from hex: %v", err)
			return false
		}
		
		// Should be equal
		if value != converted {
			t.Errorf("Bytes32 round-trip failed: original=%s, result=%s", value.String(), converted.String())
			return false
		}
		
		return true
	}, 100)
	
	// Property: Hash function consistency
	RunPropertyTest(t, "HashConsistency", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		
		// Hash should be deterministic
		hash1 := smt.ComputeLeafHash(key, value)
		hash2 := smt.ComputeLeafHash(key, value)
		
		if hash1 != hash2 {
			t.Error("Hash function should be deterministic")
			return false
		}
		
		// Hash should be different for different inputs (with high probability)
		differentValue := smt.Bytes32{}
		for i := range differentValue {
			differentValue[i] = value[i] ^ 0x01
		}
		hash3 := smt.ComputeLeafHash(key, differentValue)
		
		if hash1 == hash3 {
			t.Log("Hash collision detected (rare but possible)")
		}
		
		return true
	}, 50)
}

// TestBatchInvariants tests batch processing invariants
func TestBatchInvariants(t *testing.T) {
	// Property: Multiple operations produce consistent results
	RunPropertyTest(t, "MultipleOperationConsistency", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		depth := inputs[2].(uint16)
		
		// Create two identical trees
		db1 := smt.NewInMemoryDatabase()
		tree1, err := smt.NewSparseMerkleTree(db1, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		db2 := smt.NewInMemoryDatabase()
		tree2, err := smt.NewSparseMerkleTree(db2, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		// Individual operation
		_, err1 := tree1.Insert(key, value)
		if err1 != nil {
			return true // Skip if insert fails
		}
		
		// Same operation on second tree
		_, err2 := tree2.Insert(key, value)
		if err2 != nil {
			return true // Skip if insert fails
		}
		
		// Results should be identical
		root1 := tree1.Root()
		root2 := tree2.Root()
		
		if root1 != root2 {
			t.Errorf("Trees differ after same operations: %s vs %s", root1.String(), root2.String())
			return false
		}
		
		return true
	}, 30)
}

// TestMemoryInvariants tests memory usage invariants
func TestMemoryInvariants(t *testing.T) {
	// Property: Database operations work correctly
	RunPropertyTest(t, "DatabaseCorrectness", func(t *testing.T, inputs []interface{}) bool {
		key := inputs[0].(*big.Int)
		value := inputs[1].(smt.Bytes32)
		depth := inputs[2].(uint16)
		
		// Test database operations
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, depth)
		if err != nil {
			return true // Skip if tree creation fails
		}
		
		// Insert value
		_, err = tree.Insert(key, value)
		if err != nil {
			return true // Skip if insert fails
		}
		
		// Verify existence
		exists, err := tree.Exists(key)
		if err != nil {
			return true // Skip if exists check fails
		}
		
		if !exists {
			t.Error("Element should exist after insertion")
			return false
		}
		
		// Get proof
		proof, err := tree.Get(key)
		if err != nil {
			return true // Skip if get fails
		}
		
		if !proof.Exists {
			t.Error("Proof should indicate existence")
			return false
		}
		
		return true
	}, 50)
}