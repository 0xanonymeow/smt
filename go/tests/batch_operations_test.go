package tests

import (
	"math/big"
	"testing"

	smt "github.com/0xanonymeow/smt/go"
)

func TestBatchInsert(t *testing.T) {
	tree := CreateTestTree(t, 8)

	// Prepare batch data
	indices := []*big.Int{
		big.NewInt(1),
		big.NewInt(2),
		big.NewInt(3),
	}
	
	values := []smt.Bytes32{
		GenerateRandomBytes32(1),
		GenerateRandomBytes32(2),
		GenerateRandomBytes32(3),
	}

	// Test batch insert
	proofs, err := tree.BatchInsert(indices, values)
	if err != nil {
		t.Fatalf("BatchInsert failed: %v", err)
	}

	if len(proofs) != len(indices) {
		t.Fatalf("Expected %d proofs, got %d", len(indices), len(proofs))
	}

	// Verify all insertions
	for i, index := range indices {
		// Check proof structure
		if proofs[i].Exists {
			t.Errorf("Expected proof[%d].Exists to be false for new insertion", i)
		}

		expectedLeafHash := smt.ComputeLeafHash(index, values[i])
		if proofs[i].NewLeaf != expectedLeafHash {
			t.Errorf("Expected proof[%d].NewLeaf to be computed hash %s, got %s", 
				i, expectedLeafHash.String(), proofs[i].NewLeaf.String())
		}

		// Verify the item exists in tree
		exists, err := tree.Exists(index)
		if err != nil {
			t.Fatalf("Exists check failed for index %d: %v", i, err)
		}
		if !exists {
			t.Errorf("Expected index %d to exist after batch insert", i)
		}

		// Verify proof validation
		proof, err := tree.Get(index)
		if err != nil {
			t.Fatalf("Get failed for index %d: %v", i, err)
		}
		if !tree.VerifyProof(proof) {
			t.Errorf("Proof verification failed for index %d", i)
		}
	}
}

func TestBatchUpdate(t *testing.T) {
	tree := CreateTestTree(t, 8)

	// First, insert some initial data
	indices := []*big.Int{
		big.NewInt(5),
		big.NewInt(10),
		big.NewInt(15),
	}
	
	initialValues := []smt.Bytes32{
		GenerateRandomBytes32(5),
		GenerateRandomBytes32(10),
		GenerateRandomBytes32(15),
	}

	// Insert initial values
	for i, index := range indices {
		_, err := tree.Insert(index, initialValues[i])
		if err != nil {
			t.Fatalf("Initial insert failed for index %d: %v", i, err)
		}
	}

	// Prepare update data
	newValues := []smt.Bytes32{
		GenerateRandomBytes32(50),
		GenerateRandomBytes32(100),
		GenerateRandomBytes32(150),
	}

	// Test batch update
	proofs, err := tree.BatchUpdate(indices, newValues)
	if err != nil {
		t.Fatalf("BatchUpdate failed: %v", err)
	}

	if len(proofs) != len(indices) {
		t.Fatalf("Expected %d proofs, got %d", len(indices), len(proofs))
	}

	// Verify all updates
	for i, index := range indices {
		// Check proof structure
		if !proofs[i].Exists {
			t.Errorf("Expected proof[%d].Exists to be true for update", i)
		}

		expectedOldLeafHash := smt.ComputeLeafHash(index, initialValues[i])
		if proofs[i].Leaf != expectedOldLeafHash {
			t.Errorf("Expected proof[%d].Leaf to be old computed hash %s, got %s", 
				i, expectedOldLeafHash.String(), proofs[i].Leaf.String())
		}

		expectedNewLeafHash := smt.ComputeLeafHash(index, newValues[i])
		if proofs[i].NewLeaf != expectedNewLeafHash {
			t.Errorf("Expected proof[%d].NewLeaf to be new computed hash %s, got %s", 
				i, expectedNewLeafHash.String(), proofs[i].NewLeaf.String())
		}

		// Verify the updated value in tree
		proof, err := tree.Get(index)
		if err != nil {
			t.Fatalf("Get failed for index %d: %v", i, err)
		}
		if proof.Value != newValues[i] {
			t.Errorf("Expected updated value %s, got %s", newValues[i].String(), proof.Value.String())
		}

		// Verify proof validation
		if !tree.VerifyProof(proof) {
			t.Errorf("Proof verification failed for updated index %d", i)
		}
	}
}

func TestBatchGet(t *testing.T) {
	tree := CreateTestTree(t, 8)

	// Insert test data
	indices := []*big.Int{
		big.NewInt(7),
		big.NewInt(14),
		big.NewInt(21),
	}
	
	values := []smt.Bytes32{
		GenerateRandomBytes32(7),
		GenerateRandomBytes32(14),
		GenerateRandomBytes32(21),
	}

	for i, index := range indices {
		_, err := tree.Insert(index, values[i])
		if err != nil {
			t.Fatalf("Insert failed for index %d: %v", i, err)
		}
	}

	// Test batch get
	proofs, err := tree.BatchGet(indices)
	if err != nil {
		t.Fatalf("BatchGet failed: %v", err)
	}

	if len(proofs) != len(indices) {
		t.Fatalf("Expected %d proofs, got %d", len(indices), len(proofs))
	}

	// Verify all proofs
	for i, index := range indices {
		if !proofs[i].Exists {
			t.Errorf("Expected proof[%d] to exist", i)
		}

		if proofs[i].Index.Cmp(index) != 0 {
			t.Errorf("Expected proof[%d].Index to be %s, got %s", 
				i, index.String(), proofs[i].Index.String())
		}

		if proofs[i].Value != values[i] {
			t.Errorf("Expected proof[%d].Value to be %s, got %s", 
				i, values[i].String(), proofs[i].Value.String())
		}

		expectedLeafHash := smt.ComputeLeafHash(index, values[i])
		if proofs[i].Leaf != expectedLeafHash {
			t.Errorf("Expected proof[%d].Leaf to be computed hash %s, got %s", 
				i, expectedLeafHash.String(), proofs[i].Leaf.String())
		}

		// Verify proof validation
		if !tree.VerifyProof(proofs[i]) {
			t.Errorf("Proof verification failed for index %d", i)
		}
	}
}

func TestBatchExists(t *testing.T) {
	tree := CreateTestTree(t, 8)

	// Insert some data
	existingIndices := []*big.Int{
		big.NewInt(3),
		big.NewInt(9),
	}
	
	for _, index := range existingIndices {
		value := GenerateRandomBytes32(int(index.Int64()))
		_, err := tree.Insert(index, value)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	// Test batch exists with mix of existing and non-existing
	testIndices := []*big.Int{
		big.NewInt(3),   // exists
		big.NewInt(6),   // doesn't exist
		big.NewInt(9),   // exists
		big.NewInt(12),  // doesn't exist
	}

	results, err := tree.BatchExists(testIndices)
	if err != nil {
		t.Fatalf("BatchExists failed: %v", err)
	}

	expectedResults := []bool{true, false, true, false}
	
	if len(results) != len(expectedResults) {
		t.Fatalf("Expected %d results, got %d", len(expectedResults), len(results))
	}

	for i, expected := range expectedResults {
		if results[i] != expected {
			t.Errorf("Expected results[%d] to be %v, got %v", i, expected, results[i])
		}
	}
}

func TestBatchInsertKV(t *testing.T) {
	tree := CreateTestTree(t, 16)

	// BatchInsertKV takes a map, not slices
	kvPairs := map[string]smt.Bytes32{
		"key1": GenerateRandomBytes32(1),
		"key2": GenerateRandomBytes32(2),
		"key3": GenerateRandomBytes32(3),
	}

	// Test batch KV insert
	proofs, err := tree.BatchInsertKV(kvPairs)
	if err != nil {
		t.Fatalf("BatchInsertKV failed: %v", err)
	}

	if len(proofs) != len(kvPairs) {
		t.Fatalf("Expected %d proofs, got %d", len(kvPairs), len(proofs))
	}

	// Verify all KV insertions
	for key, expectedValue := range kvPairs {
		// Verify the KV item exists in tree
		value, exists, err := tree.GetKV(key)
		if err != nil {
			t.Fatalf("GetKV failed for key %s: %v", key, err)
		}
		if !exists {
			t.Errorf("Expected key %s to exist after batch insert", key)
		}
		if value != expectedValue {
			t.Errorf("Expected KV value %s, got %s", expectedValue.String(), value.String())
		}
	}
}

func TestBatchGetKV(t *testing.T) {
	tree := CreateTestTree(t, 16)

	keys := []string{"test1", "test2", "test3"}
	values := []smt.Bytes32{
		GenerateRandomBytes32(10),
		GenerateRandomBytes32(20),
		GenerateRandomBytes32(30),
	}

	// Insert KV data
	for i, key := range keys {
		_, err := tree.InsertKV(key, values[i])
		if err != nil {
			t.Fatalf("InsertKV failed for key %s: %v", key, err)
		}
	}

	// Test batch KV get - returns map[string]Bytes32
	results, err := tree.BatchGetKV(keys)
	if err != nil {
		t.Fatalf("BatchGetKV failed: %v", err)
	}

	if len(results) != len(keys) {
		t.Fatalf("Expected %d results, got %d", len(keys), len(results))
	}

	// Verify all results
	for i, key := range keys {
		result, exists := results[key]
		if !exists {
			t.Errorf("Expected result for key %s to exist", key)
		}
		if result != values[i] {
			t.Errorf("Expected result for key %s to be %s, got %s", 
				key, values[i].String(), result.String())
		}
	}
}

// Note: ExecuteBatch and NewBatch methods don't exist in the current implementation
// This test is commented out as the functionality is not available
/*
func TestExecuteBatch(t *testing.T) {
	// This functionality is not implemented yet
	t.Skip("ExecuteBatch functionality not implemented")
}
*/

func TestBatchOperationsErrorHandling(t *testing.T) {
	tree := CreateTestTree(t, 8)

	// Test batch insert with empty input
	emptyIndices := []*big.Int{}
	emptyValues := []smt.Bytes32{}

	proofs, err := tree.BatchInsert(emptyIndices, emptyValues)
	if err != nil {
		t.Errorf("BatchInsert with empty arrays should not error: %v", err)
	}
	if len(proofs) != 0 {
		t.Errorf("Expected 0 proofs for empty insert, got %d", len(proofs))
	}

	// Test mismatched array lengths
	mismatchedIndices := []*big.Int{big.NewInt(1), big.NewInt(2)}
	mismatchedValues := []smt.Bytes32{GenerateRandomBytes32(1)} // one less

	_, err = tree.BatchInsert(mismatchedIndices, mismatchedValues)
	if err == nil {
		t.Error("Expected error for mismatched array lengths")
	}
}