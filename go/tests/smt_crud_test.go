package tests

import (
	"math/big"
	"testing"

	smt "github.com/0xanonymeow/smt/go"
)

// TestSparseMerkleTreeCRUD tests the CRUD operations
func TestSparseMerkleTreeCRUD(t *testing.T) {
	// Create a new SparseMerkleTree with depth 8 for testing
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Test initial state
	zeroBytes32 := smt.Bytes32{}
	if tree.Root() != zeroBytes32 {
		t.Fatalf("Expected empty tree root to be zero, got %s", tree.Root().String())
	}

	// Test Insert operation
	index := big.NewInt(5)
	leafValue, err := smt.HexToBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("Failed to parse leaf value: %v", err)
	}

	// Insert should succeed for new key
	updateProof, err := tree.Insert(index, leafValue)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if updateProof == nil {
		t.Fatal("Insert should return UpdateProof")
	}

	// NewLeaf should be the computed leaf hash, not the raw value
	expectedLeafHash := smt.ComputeLeafHash(index, leafValue)
	if updateProof.NewLeaf != expectedLeafHash {
		t.Fatalf("Expected newLeaf to be %s (computed hash), got %s", expectedLeafHash.String(), updateProof.NewLeaf.String())
	}

	// Test Exists operation
	exists, err := tree.Exists(index)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("Key should exist after insert")
	}

	// Test Get operation
	proof, err := tree.Get(index)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if proof == nil {
		t.Fatal("Get should return proof")
	}

	if !proof.Exists {
		t.Fatal("Proof should indicate key exists")
	}

	if proof.Value.IsZero() {
		t.Fatal("Proof should have value")
	}

	if proof.Index.Cmp(index) != 0 {
		t.Fatalf("Expected index %s, got %s", index.String(), proof.Index.String())
	}

	// Test Insert duplicate key (should fail)
	duplicateValue, _ := smt.HexToBytes32("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	_, err = tree.Insert(index, duplicateValue)
	if err == nil {
		t.Fatal("Insert should fail for existing key")
	}

	// Test Update operation
	newLeafValue, err := smt.HexToBytes32("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")
	if err != nil {
		t.Fatalf("Failed to parse new leaf value: %v", err)
	}
	
	updateProof, err = tree.Update(index, newLeafValue)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// NewLeaf should be the computed leaf hash for updates
	expectedLeafHashUpdate := smt.ComputeLeafHash(index, newLeafValue)
	if updateProof.NewLeaf != expectedLeafHashUpdate {
		t.Fatalf("Expected newLeaf to be %s (computed hash), got %s", expectedLeafHashUpdate.String(), updateProof.NewLeaf.String())
	}

	// Verify the value was updated
	proof, err = tree.Get(index)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if proof.Value.IsZero() {
		t.Fatal("Updated proof should have value")
	}

	// Test Update non-existent key (should fail)
	nonExistentIndex := big.NewInt(99)
	_, err = tree.Update(nonExistentIndex, duplicateValue)
	if err == nil {
		t.Fatal("Update should fail for non-existent key")
	}

	// Test VerifyProof
	isValid := tree.VerifyProof(proof)
	if !isValid {
		t.Fatal("Proof should be valid")
	}

	t.Logf("All CRUD operations completed successfully")
	t.Logf("Final tree root: %s", tree.Root().String())
}

// TestSparseMerkleTreeMultipleOperations tests multiple insert/update operations
func TestSparseMerkleTreeMultipleOperations(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Insert multiple keys
	keys := []*big.Int{
		big.NewInt(1),
		big.NewInt(2),
		big.NewInt(3),
		big.NewInt(10),
		big.NewInt(255), // Max value for depth 8
	}

	values := []smt.Bytes32{}
	valueStrings := []string{
		"0x1111111111111111111111111111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222222222222222222222222222",
		"0x3333333333333333333333333333333333333333333333333333333333333333",
		"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	}

	// Parse values
	for _, vs := range valueStrings {
		v, err := smt.HexToBytes32(vs)
		if err != nil {
			t.Fatalf("Failed to parse value: %v", err)
		}
		values = append(values, v)
	}

	// Insert all keys
	for i, key := range keys {
		_, err := tree.Insert(key, values[i])
		if err != nil {
			t.Fatalf("Failed to insert key %s: %v", key.String(), err)
		}

		// Verify key exists
		exists, err := tree.Exists(key)
		if err != nil {
			t.Fatalf("Exists check failed: %v", err)
		}
		if !exists {
			t.Fatalf("Key %s should exist after insert", key.String())
		}
	}

	// Verify all keys still exist
	for _, key := range keys {
		exists, err := tree.Exists(key)
		if err != nil {
			t.Fatalf("Exists check failed: %v", err)
		}
		if !exists {
			t.Fatalf("Key %s should still exist", key.String())
		}

		proof, err := tree.Get(key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !proof.Exists {
			t.Fatalf("Proof for key %s should indicate existence", key.String())
		}

		// Verify proof
		isValid := tree.VerifyProof(proof)
		if !isValid {
			t.Fatalf("Proof for key %s should be valid", key.String())
		}
	}

	// Update some keys
	updatedValues := []smt.Bytes32{}
	updatedValueStrings := []string{
		"0xdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
		"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	}

	for _, vs := range updatedValueStrings {
		v, err := smt.HexToBytes32(vs)
		if err != nil {
			t.Fatalf("Failed to parse updated value: %v", err)
		}
		updatedValues = append(updatedValues, v)
	}

	for i := 0; i < 2; i++ {
		_, err := tree.Update(keys[i], updatedValues[i])
		if err != nil {
			t.Fatalf("Failed to update key %s: %v", keys[i].String(), err)
		}
	}

	// Verify updates
	for i := 0; i < 2; i++ {
		proof, err := tree.Get(keys[i])
		if err != nil {
			t.Fatalf("Get failed after update: %v", err)
		}
		if !proof.Exists {
			t.Fatalf("Updated key %s should still exist", keys[i].String())
		}

		// Verify proof is still valid after update
		isValid := tree.VerifyProof(proof)
		if !isValid {
			t.Fatalf("Proof for updated key %s should be valid", keys[i].String())
		}
	}

	t.Logf("Multiple operations test completed successfully")
	t.Logf("Final tree root: %s", tree.Root().String())
}

// TestSparseMerkleTreeErrorHandling tests error conditions
func TestSparseMerkleTreeErrorHandling(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	index := big.NewInt(5)
	leafValue, err := smt.HexToBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("Failed to parse leaf value: %v", err)
	}

	// Test Update on non-existent key
	_, err = tree.Update(index, leafValue)
	if err == nil {
		t.Fatal("Update should fail for non-existent key")
	}

	// Insert the key
	_, err = tree.Insert(index, leafValue)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test Insert on existing key
	deadbeef, _ := smt.HexToBytes32("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	_, err = tree.Insert(index, deadbeef)
	if err == nil {
		t.Fatal("Insert should fail for existing key")
	}

	// Test Get on non-existent key
	nonExistentIndex := big.NewInt(99)
	proof, err := tree.Get(nonExistentIndex)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if proof.Exists {
		t.Fatal("Proof should indicate key does not exist")
	}

	if !proof.Value.IsZero() {
		t.Fatal("Non-existent key should have zero value")
	}

	// Test Exists on non-existent key
	exists, err := tree.Exists(nonExistentIndex)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("Non-existent key should return false for Exists")
	}

	t.Logf("Error handling test completed successfully")
}