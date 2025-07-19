package tests

import (
	"math/big"
	"testing"

	smt "github.com/0xanonymeow/smt/go"
)

func TestUpsertComprehensiveCoverage(t *testing.T) {
	// Test comprehensive upsert scenarios to improve coverage from 47% to 100%
	
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 16) // Use larger depth for more complex scenarios
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Test case 1: Simple upsert on empty tree (insert path)
	index1 := big.NewInt(100)
	value1 := smt.Bytes32{1, 2, 3}
	
	proof, err := tree.Insert(index1, value1)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}
	if proof.Exists {
		t.Error("Proof should show old leaf didn't exist for insert (Exists should be false)")
	}

	// Test case 2: Update existing leaf (normal update path)
	newValue1 := smt.Bytes32{4, 5, 6}
	updateProof, err := tree.Update(index1, newValue1)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}
	if !updateProof.Exists {
		t.Error("Update proof should show old leaf existed")
	}

	// Test case 3: Complex divergence scenario - create two leaves that diverge at different levels
	// This triggers the path divergence logic in upsert
	
	index2 := big.NewInt(356)  // Choose indices that will diverge at specific levels
	value2 := smt.Bytes32{10, 11, 12}
	
	_, err = tree.Insert(index2, value2)
	if err != nil {
		t.Fatalf("Failed to insert divergent leaf: %v", err)
	}

	// Test case 4: Insert multiple leaves that create complex tree structures
	testIndices := []*big.Int{
		big.NewInt(200),
		big.NewInt(300), 
		big.NewInt(400),
		big.NewInt(500),
		big.NewInt(600),
		big.NewInt(1000),
		big.NewInt(2000),
		big.NewInt(4000),
		big.NewInt(8000),
	}
	
	for i, idx := range testIndices {
		value := smt.Bytes32{byte(i + 20)}
		_, err = tree.Insert(idx, value)
		if err != nil {
			t.Fatalf("Failed to insert index %s: %v", idx.String(), err)
		}
	}

	// Test case 5: Update leaves in a tree with complex structure
	for i, idx := range testIndices[:5] { // Update first 5
		newValue := smt.Bytes32{byte(i + 100)}
		_, err = tree.Update(idx, newValue)
		if err != nil {
			t.Fatalf("Failed to update index %s: %v", idx.String(), err)
		}
	}

	// Test case 6: Test scenario where old and new leaf have same index (normal update)
	// This should trigger the !oldProof.Exists || oldIndex == index path
	sameIndexValue := smt.Bytes32{99, 98, 97}
	_, err = tree.Update(index1, sameIndexValue)
	if err != nil {
		t.Fatalf("Failed to update same index: %v", err)
	}

	// Test case 7: Insert and immediately update to test various upsert paths
	rapidIndex := big.NewInt(9999)
	rapidValue1 := smt.Bytes32{50, 51, 52}
	
	_, err = tree.Insert(rapidIndex, rapidValue1)
	if err != nil {
		t.Fatalf("Failed rapid insert: %v", err)
	}
	
	rapidValue2 := smt.Bytes32{60, 61, 62}
	_, err = tree.Update(rapidIndex, rapidValue2)
	if err != nil {
		t.Fatalf("Failed rapid update: %v", err)
	}

	// Test case 8: Test edge cases that trigger different bit patterns in divergence logic
	// Create patterns that diverge at various levels to exercise all paths
	bitTestIndices := []*big.Int{
		big.NewInt(1),       // ...00001
		big.NewInt(2),       // ...00010
		big.NewInt(4),       // ...00100
		big.NewInt(8),       // ...01000
		big.NewInt(16),      // ...10000
		big.NewInt(32),      // ...100000
		big.NewInt(65),      // ...1000001 (diverges at bit 6 from 1)
		big.NewInt(129),     // ...10000001 (diverges at bit 7 from 1)
	}
	
	for i, idx := range bitTestIndices {
		value := smt.Bytes32{byte(i + 200)}
		_, err = tree.Insert(idx, value)
		if err != nil {
			t.Fatalf("Failed to insert bit test index %s: %v", idx.String(), err)
		}
	}

	// Test case 9: Verify all insertions were successful
	for _, idx := range append(testIndices, bitTestIndices...) {
		exists, err := tree.Exists(idx)
		if err != nil {
			t.Fatalf("Failed to check exists for %s: %v", idx.String(), err)
		}
		if !exists {
			t.Errorf("Index %s should exist but doesn't", idx.String())
		}
	}

	// Test case 10: Test zero index scenarios
	zeroIndex := big.NewInt(0)
	zeroValue := smt.Bytes32{255, 254, 253}
	
	_, err = tree.Insert(zeroIndex, zeroValue)
	if err != nil {
		t.Fatalf("Failed to insert zero index: %v", err)
	}
	
	// Update zero index
	newZeroValue := smt.Bytes32{250, 249, 248}
	_, err = tree.Update(zeroIndex, newZeroValue)
	if err != nil {
		t.Fatalf("Failed to update zero index: %v", err)
	}
}

func TestUpsertDivergenceLogicCoverage(t *testing.T) {
	// Test specific divergence scenarios to cover all paths in upsert
	
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Test case 1: Create scenario where oldProof.Exists && oldProof.Index != index
	// This should trigger the path divergence logic
	
	// First insert a leaf
	oldIndex := big.NewInt(5)  // Binary: 00000101
	oldValue := smt.Bytes32{1, 2, 3}
	
	_, err = tree.Insert(oldIndex, oldValue)
	if err != nil {
		t.Fatalf("Failed to insert old leaf: %v", err)
	}

	// Now try to insert at a different index that will cause path divergence
	// This should be handled by the upsert function's special case logic
	newIndex := big.NewInt(7)  // Binary: 00000111 (differs from 5 at bit 1)
	newValue := smt.Bytes32{10, 20, 30}
	
	_, err = tree.Insert(newIndex, newValue)
	if err != nil {
		t.Fatalf("Failed to insert new leaf: %v", err)
	}

	// Verify both exist
	exists1, err := tree.Exists(oldIndex)
	if err != nil || !exists1 {
		t.Errorf("Old index should exist: %v", err)
	}
	
	exists2, err := tree.Exists(newIndex)
	if err != nil || !exists2 {
		t.Errorf("New index should exist: %v", err)
	}

	// Test case 2: Test various divergence levels
	divergenceTests := []struct {
		index1, index2 *big.Int
		description    string
	}{
		{big.NewInt(0), big.NewInt(1), "diverge at bit 0"},
		{big.NewInt(0), big.NewInt(2), "diverge at bit 1"},
		{big.NewInt(0), big.NewInt(4), "diverge at bit 2"},
		{big.NewInt(0), big.NewInt(8), "diverge at bit 3"},
		{big.NewInt(1), big.NewInt(3), "diverge at bit 1 (both have bit 0 set)"},
		{big.NewInt(5), big.NewInt(13), "diverge at bit 3"},
	}

	for i, test := range divergenceTests {
		// Create fresh tree for each test
		freshDB := smt.NewInMemoryDatabase()
		freshTree, err := smt.NewSparseMerkleTree(freshDB, 8)
		if err != nil {
			t.Fatalf("Test %d: Failed to create tree: %v", i, err)
		}

		// Insert first leaf
		value1 := smt.Bytes32{byte(i * 10)}
		_, err = freshTree.Insert(test.index1, value1)
		if err != nil {
			t.Fatalf("Test %d (%s): Failed to insert first: %v", i, test.description, err)
		}

		// Insert second leaf (should trigger divergence logic)
		value2 := smt.Bytes32{byte(i*10 + 5)}
		_, err = freshTree.Insert(test.index2, value2)
		if err != nil {
			t.Fatalf("Test %d (%s): Failed to insert second: %v", i, test.description, err)
		}

		// Verify both exist
		exists1, err := freshTree.Exists(test.index1)
		if err != nil || !exists1 {
			t.Errorf("Test %d (%s): First index should exist", i, test.description)
		}
		
		exists2, err := freshTree.Exists(test.index2)
		if err != nil || !exists2 {
			t.Errorf("Test %d (%s): Second index should exist", i, test.description)
		}

		// Try to update each one to trigger normal upsert paths
		newValue1 := smt.Bytes32{byte(i*10 + 1)}
		_, err = freshTree.Update(test.index1, newValue1)
		if err != nil {
			t.Fatalf("Test %d (%s): Failed to update first: %v", i, test.description, err)
		}

		newValue2 := smt.Bytes32{byte(i*10 + 6)}
		_, err = freshTree.Update(test.index2, newValue2)
		if err != nil {
			t.Fatalf("Test %d (%s): Failed to update second: %v", i, test.description, err)
		}
	}
}

func TestUpsertSiblingConstructionCoverage(t *testing.T) {
	// Test the sibling construction logic in upsert to get full coverage
	
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 10) // Medium depth for complex sibling scenarios
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Test case 1: Create a tree structure with multiple levels to test sibling logic
	baseIndices := []*big.Int{
		big.NewInt(1),
		big.NewInt(2),
		big.NewInt(4),
		big.NewInt(8),
		big.NewInt(16),
	}

	// Insert base indices
	for i, idx := range baseIndices {
		value := smt.Bytes32{byte(i + 1)}
		_, err = tree.Insert(idx, value)
		if err != nil {
			t.Fatalf("Failed to insert base index %s: %v", idx.String(), err)
		}
	}

	// Test case 2: Update each leaf to exercise the sibling reconstruction logic
	for i, idx := range baseIndices {
		newValue := smt.Bytes32{byte(i + 50)}
		_, err = tree.Update(idx, newValue)
		if err != nil {
			t.Fatalf("Failed to update index %s: %v", idx.String(), err)
		}

		// Verify the update worked by checking the value
		proof, err := tree.Get(idx)
		if err != nil {
			t.Fatalf("Failed to get proof for updated index %s: %v", idx.String(), err)
		}
		
		if proof.Value != newValue {
			t.Errorf("Updated value mismatch for index %s: expected %v, got %v", 
				idx.String(), newValue, proof.Value)
		}
	}

	// Test case 3: Test scenarios with different enables bit patterns
	// This exercises the GetBit(oldProof.Enables, i) logic in upsert
	
	// Insert leaves at positions that will create different enables patterns
	enablesTestIndices := []*big.Int{
		big.NewInt(17),   // Binary pattern that creates specific enables
		big.NewInt(33),   // Different pattern
		big.NewInt(65),   // Another pattern
		big.NewInt(129),  // Yet another
	}

	for i, idx := range enablesTestIndices {
		value := smt.Bytes32{byte(i + 100)}
		_, err = tree.Insert(idx, value)
		if err != nil {
			t.Fatalf("Failed to insert enables test index %s: %v", idx.String(), err)
		}
	}

	// Update these leaves to test the sibling indexing logic
	for i, idx := range enablesTestIndices {
		newValue := smt.Bytes32{byte(i + 150)}
		updateProof, err := tree.Update(idx, newValue)
		if err != nil {
			t.Fatalf("Failed to update enables test index %s: %v", idx.String(), err)
		}

		// Verify the update proof has correct structure
		if !updateProof.Exists {
			t.Errorf("Update proof should show old leaf existed for index %s", idx.String())
		}

		// Verify the new leaf hash is different from old leaf hash
		if updateProof.NewLeaf == updateProof.Leaf {
			t.Errorf("New leaf hash should be different from old leaf hash for index %s", idx.String())
		}
	}

	// Test case 4: Edge case where sibling array is empty
	// Create a scenario where enables is zero (no siblings)
	emptyTree := smt.NewInMemoryDatabase()
	singleTree, err := smt.NewSparseMerkleTree(emptyTree, 8)
	if err != nil {
		t.Fatalf("Failed to create single tree: %v", err)
	}

	singleIndex := big.NewInt(42)
	singleValue := smt.Bytes32{200, 201, 202}
	
	_, err = singleTree.Insert(singleIndex, singleValue)
	if err != nil {
		t.Fatalf("Failed to insert single leaf: %v", err)
	}

	// Update it (should have no siblings, enables = 0)
	newSingleValue := smt.Bytes32{210, 211, 212}
	_, err = singleTree.Update(singleIndex, newSingleValue)
	if err != nil {
		t.Fatalf("Failed to update single leaf: %v", err)
	}
}

func TestUpsertErrorPaths(t *testing.T) {
	// Test all error paths in upsert function to achieve 100% coverage

	// Test case 1: validateIndex error - index out of range
	t.Run("validateIndex error", func(t *testing.T) {
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, 8)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Test negative index
		negativeIndex := big.NewInt(-1)
		_, err = tree.Insert(negativeIndex, smt.Bytes32{1})
		if err == nil {
			t.Error("Expected error for negative index")
		}

		// Test index >= 2^depth
		tooLargeIndex := big.NewInt(256) // >= 2^8
		_, err = tree.Insert(tooLargeIndex, smt.Bytes32{1})
		if err == nil {
			t.Error("Expected error for index >= 2^depth")
		}

		// Also test update with invalid index
		_, err = tree.Update(negativeIndex, smt.Bytes32{2})
		if err == nil {
			t.Error("Expected error for update with negative index")
		}
	})

	// Test case 2: Database Get error during proof retrieval
	t.Run("database Get error during proof retrieval", func(t *testing.T) {
		mockDB := &UpsertMockDatabase{
			data:          make(map[string][]byte),
			shouldFailGet: false,
			shouldFailSet: false,
			shouldFailDelete: false,
		}
		
		tree, err := smt.NewSparseMerkleTree(mockDB, 8)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Insert something first to establish state
		_, err = tree.Insert(big.NewInt(1), smt.Bytes32{1})
		if err != nil {
			t.Fatalf("Failed initial insert: %v", err)
		}

		// Now make Get operations fail and try update
		mockDB.shouldFailGet = true
		_, err = tree.Update(big.NewInt(1), smt.Bytes32{2})
		if err == nil {
			t.Error("Expected error when database Get fails during proof retrieval")
		}
	})

	// Test case 3: Database Set error during setLeaf
	t.Run("database Set error during setLeaf", func(t *testing.T) {
		mockDB := &UpsertMockDatabase{
			data:          make(map[string][]byte),
			shouldFailGet: false,
			shouldFailSet: false,
			shouldFailDelete: false,
		}
		
		tree, err := smt.NewSparseMerkleTree(mockDB, 8)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Make Set operations fail from the start
		mockDB.shouldFailSet = true
		_, err = tree.Insert(big.NewInt(1), smt.Bytes32{1})
		if err == nil {
			t.Error("Expected error when setLeaf fails")
		}
	})

	// Test case 4: Database Set error during setNode in divergence path  
	t.Run("database Set error during divergence setNode", func(t *testing.T) {
		mockDB := &UpsertMockDatabase{
			data:          make(map[string][]byte),
			shouldFailGet: false,
			shouldFailSet: false,
			shouldFailDelete: false,
		}
		
		tree, err := smt.NewSparseMerkleTree(mockDB, 8)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Insert first leaf successfully
		_, err = tree.Insert(big.NewInt(1), smt.Bytes32{1})
		if err != nil {
			t.Fatalf("Failed first insert: %v", err)
		}

		// Now make Set fail and try to insert a divergent leaf
		mockDB.shouldFailSet = true
		_, err = tree.Insert(big.NewInt(2), smt.Bytes32{2}) // Should trigger divergence logic
		if err == nil {
			t.Error("Expected error when setNode fails during divergence reconstruction")
		}
	})

	// Test case 5: Database Set error during normal sibling reconstruction
	t.Run("database Set error during normal reconstruction", func(t *testing.T) {
		mockDB := &UpsertMockDatabase{
			data:          make(map[string][]byte),
			shouldFailGet: false,
			shouldFailSet: false,
			shouldFailDelete: false,
		}
		
		tree, err := smt.NewSparseMerkleTree(mockDB, 8)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Create a tree structure with multiple nodes
		_, err = tree.Insert(big.NewInt(1), smt.Bytes32{1})
		if err != nil {
			t.Fatalf("Failed first insert: %v", err)
		}
		_, err = tree.Insert(big.NewInt(2), smt.Bytes32{2})
		if err != nil {
			t.Fatalf("Failed second insert: %v", err)
		}
		_, err = tree.Insert(big.NewInt(4), smt.Bytes32{4})
		if err != nil {
			t.Fatalf("Failed third insert: %v", err)
		}

		// Now make Set fail and try to update (should trigger normal reconstruction)
		mockDB.shouldFailSet = true
		_, err = tree.Update(big.NewInt(1), smt.Bytes32{10})
		if err == nil {
			t.Error("Expected error when setNode fails during normal reconstruction")
		}
	})

	// Test case 6: Database Delete error during old leaf cleanup
	t.Run("database Delete error during old leaf cleanup", func(t *testing.T) {
		mockDB := &UpsertMockDatabase{
			data:          make(map[string][]byte),
			shouldFailGet: false,
			shouldFailSet: false,
			shouldFailDelete: false,
		}
		
		tree, err := smt.NewSparseMerkleTree(mockDB, 8)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Insert a leaf
		_, err = tree.Insert(big.NewInt(1), smt.Bytes32{1})
		if err != nil {
			t.Fatalf("Failed insert: %v", err)
		}

		// Make Delete operations fail and try update
		mockDB.shouldFailDelete = true
		_, err = tree.Update(big.NewInt(1), smt.Bytes32{2})
		if err == nil {
			t.Error("Expected error when deleteLeaf fails during old leaf cleanup")
		}
	})

	// Test case 7: Test zero node condition - both sides zero
	t.Run("zero node storage condition", func(t *testing.T) {
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, 2) // Very small depth
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Insert and update a single leaf to test the zero node condition
		_, err = tree.Insert(big.NewInt(0), smt.Bytes32{1})
		if err != nil {
			t.Fatalf("Failed insert: %v", err)
		}

		_, err = tree.Update(big.NewInt(0), smt.Bytes32{2})
		if err != nil {
			t.Fatalf("Failed update: %v", err)
		}

		// This should exercise the condition where both node sides might be zero
		// in the normal reconstruction path
	})

	// Test case 8: Complex old leaf deletion conditions
	t.Run("old leaf deletion edge cases", func(t *testing.T) {
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, 8)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Test scenario where oldProof.Exists is false
		_, err = tree.Insert(big.NewInt(100), smt.Bytes32{1})
		if err != nil {
			t.Fatalf("Failed insert: %v", err)
		}

		// Test scenario where old leaf hash equals new leaf hash
		sameValue := smt.Bytes32{5, 5, 5}
		_, err = tree.Insert(big.NewInt(50), sameValue)
		if err != nil {
			t.Fatalf("Failed insert same value first time: %v", err)
		}

		// Update with the same value - should create same hash
		_, err = tree.Update(big.NewInt(50), sameValue)
		if err != nil {
			t.Fatalf("Failed update with same value: %v", err)
		}

		// This tests the condition: oldProof.Leaf != leafHash
	})
}

func TestUpsertDivergenceEdgeCases(t *testing.T) {
	// Test edge cases in the divergence logic

	t.Run("divergence at maximum depth", func(t *testing.T) {
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, 4) // Small depth to test edge cases
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Test indices that differ at the very last bit
		index1 := big.NewInt(0)  // 0000
		index2 := big.NewInt(8)  // 1000 (differs at bit 3 for depth 4)

		_, err = tree.Insert(index1, smt.Bytes32{1})
		if err != nil {
			t.Fatalf("Failed to insert index1: %v", err)
		}

		_, err = tree.Insert(index2, smt.Bytes32{2})
		if err != nil {
			t.Fatalf("Failed to insert index2: %v", err)
		}

		// Both should exist
		exists1, err := tree.Exists(index1)
		if err != nil || !exists1 {
			t.Error("Index1 should exist")
		}

		exists2, err := tree.Exists(index2)
		if err != nil || !exists2 {
			t.Error("Index2 should exist")
		}
	})

	t.Run("divergence loop all levels", func(t *testing.T) {
		// Test patterns that diverge at each possible level
		testPairs := []struct {
			index1, index2 int64
			description    string
		}{
			{0, 1, "diverge at level 0"},
			{0, 2, "diverge at level 1"}, 
			{0, 4, "diverge at level 2"},
			{0, 8, "diverge at level 3"},
			{0, 16, "diverge at level 4"},
			{0, 32, "diverge at level 5"},
		}

		for _, pair := range testPairs {
			// Create fresh tree for each test
			freshDB := smt.NewInMemoryDatabase()
			freshTree, err := smt.NewSparseMerkleTree(freshDB, 6)
			if err != nil {
				t.Fatalf("Failed to create fresh tree: %v", err)
			}

			idx1 := big.NewInt(pair.index1)
			idx2 := big.NewInt(pair.index2)

			_, err = freshTree.Insert(idx1, smt.Bytes32{byte(pair.index1)})
			if err != nil {
				t.Fatalf("Failed to insert %d (%s): %v", pair.index1, pair.description, err)
			}

			_, err = freshTree.Insert(idx2, smt.Bytes32{byte(pair.index2)})
			if err != nil {
				t.Fatalf("Failed to insert %d (%s): %v", pair.index2, pair.description, err)
			}

			// Verify both exist
			exists1, _ := freshTree.Exists(idx1)
			exists2, _ := freshTree.Exists(idx2)
			if !exists1 || !exists2 {
				t.Errorf("Both indices should exist for %s", pair.description)
			}
		}
	})
}

// UpsertMockDatabase for testing upsert error paths
type UpsertMockDatabase struct {
	data             map[string][]byte
	shouldFailGet    bool
	shouldFailSet    bool
	shouldFailDelete bool
}

func (m *UpsertMockDatabase) Get(key []byte) ([]byte, error) {
	if m.shouldFailGet {
		return nil, &UpsertDatabaseError{Op: "get", Key: string(key)}
	}
	
	data, exists := m.data[string(key)]
	if !exists {
		return nil, &UpsertDatabaseError{Op: "get", Key: string(key)}
	}
	
	return data, nil
}

func (m *UpsertMockDatabase) Set(key []byte, value []byte) error {
	if m.shouldFailSet {
		return &UpsertDatabaseError{Op: "set", Key: string(key)}
	}
	
	m.data[string(key)] = value
	return nil
}

func (m *UpsertMockDatabase) Delete(key []byte) error {
	if m.shouldFailDelete {
		return &UpsertDatabaseError{Op: "delete", Key: string(key)}
	}
	
	delete(m.data, string(key))
	return nil
}

func (m *UpsertMockDatabase) Has(key []byte) (bool, error) {
	_, exists := m.data[string(key)]
	return exists, nil
}

// UpsertDatabaseError for testing error paths
type UpsertDatabaseError struct {
	Op  string
	Key string
}

func (e *UpsertDatabaseError) Error() string {
	return "upsert database error: " + e.Op + " failed for key " + e.Key
}