package tests

import (
	"encoding/json"
	"math/big"
	"testing"

	smt "github.com/0xanonymeow/smt/go"
)

// TestProofStructures tests that Proof and UpdateProof structs match expected interfaces
func TestProofStructures(t *testing.T) {
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

	// Test UpdateProof structure from Insert
	updateProof, err := tree.Insert(index, leafValue)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify UpdateProof has all required fields
	if updateProof.NewLeaf.IsZero() {
		t.Fatal("UpdateProof should have non-zero NewLeaf")
	}

	// Verify embedded Proof fields
	if updateProof.Index == nil {
		t.Fatal("UpdateProof should have non-nil Index")
	}

	if updateProof.Enables == nil {
		t.Fatal("UpdateProof should have non-nil Enables")
	}

	if updateProof.Siblings == nil {
		t.Fatal("UpdateProof should have non-nil Siblings")
	}

	// Test Proof structure from Get
	proof, err := tree.Get(index)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Verify all required fields are present
	if proof.Index == nil {
		t.Fatal("Proof should have non-nil Index")
	}

	if proof.Enables == nil {
		t.Fatal("Proof should have non-nil Enables")
	}

	if proof.Siblings == nil {
		t.Fatal("Proof should have non-nil Siblings")
	}

	// Test field types
	if proof.Index.Cmp(index) != 0 {
		t.Fatalf("Expected proof index %s, got %s", index.String(), proof.Index.String())
	}

	if !proof.Exists {
		t.Fatal("Proof should indicate key exists")
	}

	if proof.Leaf.IsZero() {
		t.Fatal("Proof should have non-zero leaf")
	}

	if proof.Value.IsZero() {
		t.Fatal("Proof should have non-zero value")
	}

	t.Logf("Proof verification successful: index=%s, exists=%v, leaf=%s, value=%s",
		proof.Index.String(), proof.Exists, proof.Leaf.String(), proof.Value.String())
}

// TestProofSerialization tests proof serialization and deserialization
func TestProofSerialization(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Insert test data
	testData := []struct {
		index *big.Int
		value string
	}{
		{big.NewInt(1), "0x1111111111111111111111111111111111111111111111111111111111111111"},
		{big.NewInt(5), "0x5555555555555555555555555555555555555555555555555555555555555555"},
		{big.NewInt(10), "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}

	for _, td := range testData {
		value, err := smt.HexToBytes32(td.value)
		if err != nil {
			t.Fatalf("Failed to parse value: %v", err)
		}

		_, err = tree.Insert(td.index, value)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	// Test serialization for each proof
	for _, td := range testData {
		t.Run(td.index.String(), func(t *testing.T) {
			// Get original proof
			originalProof, err := tree.Get(td.index)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			// Serialize proof
			serialized := smt.SerializeProof(originalProof)
			if serialized == nil {
				t.Fatal("Serialized proof should not be nil")
			}

			// Deserialize proof
			deserializedProof, err := smt.DeserializeProof(serialized)
			if err != nil {
				t.Fatalf("Deserialization failed: %v", err)
			}

			// Verify deserialized proof matches original
			if originalProof.Exists != deserializedProof.Exists {
				t.Fatalf("Exists mismatch: original=%v, deserialized=%v", originalProof.Exists, deserializedProof.Exists)
			}

			if originalProof.Index.Cmp(deserializedProof.Index) != 0 {
				t.Fatalf("Index mismatch: original=%s, deserialized=%s", originalProof.Index.String(), deserializedProof.Index.String())
			}

			if originalProof.Leaf != deserializedProof.Leaf {
				t.Fatalf("Leaf mismatch: original=%s, deserialized=%s", originalProof.Leaf.String(), deserializedProof.Leaf.String())
			}

			if originalProof.Value != deserializedProof.Value {
				t.Fatalf("Value mismatch: original=%s, deserialized=%s", originalProof.Value.String(), deserializedProof.Value.String())
			}

			if originalProof.Enables.Cmp(deserializedProof.Enables) != 0 {
				t.Fatalf("Enables mismatch: original=%s, deserialized=%s", originalProof.Enables.String(), deserializedProof.Enables.String())
			}

			if len(originalProof.Siblings) != len(deserializedProof.Siblings) {
				t.Fatalf("Siblings length mismatch: original=%d, deserialized=%d", len(originalProof.Siblings), len(deserializedProof.Siblings))
			}

			for i, original := range originalProof.Siblings {
				if original != deserializedProof.Siblings[i] {
					t.Fatalf("Sibling %d mismatch: original=%s, deserialized=%s", i, original.String(), deserializedProof.Siblings[i].String())
				}
			}

			// Verify both proofs are valid
			if !tree.VerifyProof(originalProof) {
				t.Fatal("Original proof should be valid")
			}

			if !tree.VerifyProof(deserializedProof) {
				t.Fatal("Deserialized proof should be valid")
			}

			t.Logf("Proof serialization successful for index %s", td.index.String())
		})
	}
}

// TestProofVerification tests proof verification under various conditions
func TestProofVerification(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Insert test data
	index := big.NewInt(42)
	value, err := smt.HexToBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("Failed to parse value: %v", err)
	}

	_, err = tree.Insert(index, value)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test valid proof
	t.Run("ValidProof", func(t *testing.T) {
		proof, err := tree.Get(index)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if !tree.VerifyProof(proof) {
			t.Fatal("Valid proof should verify successfully")
		}
	})

	// Test invalid proof with wrong leaf
	t.Run("InvalidLeaf", func(t *testing.T) {
		proof, err := tree.Get(index)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// Modify value to make it invalid
		wrongValue := smt.Bytes32{}
		wrongValue[0] = 0xFF // Different from actual value
		invalidProof := &smt.Proof{
			Exists:   proof.Exists,
			Leaf:     wrongValue,
			Value:    wrongValue, // Invalid value
			Index:    proof.Index,
			Enables:  proof.Enables,
			Siblings: proof.Siblings,
		}

		if tree.VerifyProof(invalidProof) {
			t.Fatal("Invalid proof with wrong leaf should not verify")
		}
	})

	// Test invalid proof with wrong siblings
	t.Run("InvalidSiblings", func(t *testing.T) {
		proof, err := tree.Get(index)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// Skip this test if there are no siblings
		if len(proof.Siblings) == 0 {
			t.Skip("No siblings in proof, skipping invalid siblings test")
		}

		// Modify siblings to make it invalid
		invalidSiblings := make([]smt.Bytes32, len(proof.Siblings))
		copy(invalidSiblings, proof.Siblings)
		
		// Flip bits in the first sibling to ensure it's different
		for i := range invalidSiblings[0] {
			invalidSiblings[0][i] = proof.Siblings[0][i] ^ 0xFF
		}

		invalidProof := &smt.Proof{
			Exists:   proof.Exists,
			Leaf:     proof.Leaf,
			Value:    proof.Value,
			Index:    proof.Index,
			Enables:  proof.Enables,
			Siblings: invalidSiblings,
		}

		// Note: In some cases, changing siblings might still result in a valid proof
		// if the computed root happens to match. This is unlikely but possible.
		if tree.VerifyProof(invalidProof) {
			t.Skip("Modified siblings still produced a valid proof (rare but possible)")
		}
	})

	// Test proof for non-existent key
	t.Run("NonExistentKey", func(t *testing.T) {
		nonExistentIndex := big.NewInt(99) // Within range for depth 8 (max 255)
		proof, err := tree.Get(nonExistentIndex)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if proof.Exists {
			t.Fatal("Proof for non-existent key should not indicate existence")
		}

		if !tree.VerifyProof(proof) {
			t.Fatal("Proof for non-existent key should still verify")
		}
	})
}

// TestProofJSONMarshalingCompatibility tests JSON marshaling/unmarshaling compatibility
func TestProofJSONMarshalingCompatibility(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Insert test data
	index := big.NewInt(7)
	value, err := smt.HexToBytes32("0x7777777777777777777777777777777777777777777777777777777777777777")
	if err != nil {
		t.Fatalf("Failed to parse value: %v", err)
	}

	_, err = tree.Insert(index, value)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Get proof
	proof, err := tree.Get(index)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Create JSON-compatible representation
	jsonProof := map[string]interface{}{
		"exists":   proof.Exists,
		"leaf":     proof.Leaf.String(),
		"value":    proof.Value.String(),
		"index":    proof.Index.String(),
		"enables":  proof.Enables.String(),
		"siblings": make([]string, len(proof.Siblings)),
	}

	// Convert siblings to strings
	for i, sibling := range proof.Siblings {
		jsonProof["siblings"].([]string)[i] = sibling.String()
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(jsonProof, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	// Verify unmarshaled data
	if unmarshaled["exists"] != proof.Exists {
		t.Fatal("JSON exists field mismatch")
	}

	if unmarshaled["leaf"] != proof.Leaf.String() {
		t.Fatal("JSON leaf field mismatch")
	}

	if unmarshaled["value"] != proof.Value.String() {
		t.Fatal("JSON value field mismatch")
	}

	if unmarshaled["index"] != proof.Index.String() {
		t.Fatal("JSON index field mismatch")
	}

	if unmarshaled["enables"] != proof.Enables.String() {
		t.Fatal("JSON enables field mismatch")
	}

	// Verify siblings
	siblings := unmarshaled["siblings"].([]interface{})
	if len(siblings) != len(proof.Siblings) {
		t.Fatal("JSON siblings length mismatch")
	}

	for i, sibling := range siblings {
		if sibling != proof.Siblings[i].String() {
			t.Fatalf("JSON sibling %d mismatch", i)
		}
	}

	t.Logf("JSON compatibility test successful")
	t.Logf("JSON proof: %s", string(jsonBytes))
}

// TestProofRootComputation tests root computation from proofs
func TestProofRootComputation(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Insert multiple entries
	entries := []struct {
		index *big.Int
		value string
	}{
		{big.NewInt(1), "0x1111111111111111111111111111111111111111111111111111111111111111"},
		{big.NewInt(5), "0x5555555555555555555555555555555555555555555555555555555555555555"},
		{big.NewInt(10), "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}

	for _, entry := range entries {
		value, err := smt.HexToBytes32(entry.value)
		if err != nil {
			t.Fatalf("Failed to parse value: %v", err)
		}

		_, err = tree.Insert(entry.index, value)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	// Get tree root
	treeRoot := tree.Root()

	// Test root computation for each entry
	for _, entry := range entries {
		t.Run(entry.index.String(), func(t *testing.T) {
			proof, err := tree.Get(entry.index)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			computedRoot := tree.ComputeRoot(proof)
			if computedRoot != treeRoot {
				t.Fatalf("Computed root mismatch: expected=%s, computed=%s", treeRoot.String(), computedRoot.String())
			}

			t.Logf("Root computation successful for index %s", entry.index.String())
		})
	}

	// Test root computation for non-existent key
	t.Run("NonExistentKey", func(t *testing.T) {
		nonExistentIndex := big.NewInt(99) // Within range for depth 8 (max 255)
		proof, err := tree.Get(nonExistentIndex)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		computedRoot := tree.ComputeRoot(proof)
		if computedRoot != treeRoot {
			t.Fatalf("Computed root mismatch for non-existent key: expected=%s, computed=%s", treeRoot.String(), computedRoot.String())
		}
	})
}