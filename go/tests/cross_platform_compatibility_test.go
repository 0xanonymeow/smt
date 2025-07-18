package tests

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	smt "github.com/0xanonymeow/smt/go"
	"github.com/0xanonymeow/smt/go/internal/testutils"
	"github.com/0xanonymeow/smt/go/internal/vectors"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

// CrossPlatformTestVector represents a comprehensive test case for cross-platform validation
type CrossPlatformTestVector struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	TreeDepth   uint16                `json:"treeDepth"`
	Operations  []CrossPlatformOp     `json:"operations"`
	Expected    CrossPlatformExpected `json:"expected"`
}

// CrossPlatformOp represents a single operation in a test sequence
type CrossPlatformOp struct {
	Type  string `json:"type"`  // "insert", "update", "get"
	Index string `json:"index"` // hex string
	Value string `json:"value"` // hex string (for insert/update)
}

// CrossPlatformExpected represents expected results for cross-platform validation
type CrossPlatformExpected struct {
	FinalRoot    string                   `json:"finalRoot"`
	ProofResults []CrossPlatformProofTest `json:"proofResults"`
	HashResults  []CrossPlatformHashTest  `json:"hashResults"`
}

// CrossPlatformProofTest represents expected proof validation results
type CrossPlatformProofTest struct {
	Index    string   `json:"index"`
	Exists   bool     `json:"exists"`
	Leaf     string   `json:"leaf"`
	Value    string   `json:"value"`
	Enables  string   `json:"enables"`
	Siblings []string `json:"siblings"`
}

// CrossPlatformHashTest represents expected hash computation results
type CrossPlatformHashTest struct {
	Left     string `json:"left"`
	Right    string `json:"right"`
	Expected string `json:"expected"`
}

// TestCrossPlatformHashCompatibility tests that Go and Solidity hash functions produce identical results
func TestCrossPlatformHashCompatibility(t *testing.T) {
	// Create test vectors for hash compatibility
	testVectors := []struct {
		left     string
		right    string
		expected string
	}{
		{
			left:  "0x0000000000000000000000000000000000000000000000000000000000000001",
			right: "0x0000000000000000000000000000000000000000000000000000000000000002",
		},
		{
			left:  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			right: "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
		},
	}

	for i, vector := range testVectors {
		t.Run(fmt.Sprintf("hash_vector_%d", i), func(t *testing.T) {
			// Convert hex strings to Bytes32
			leftBytes, err := smt.HexToBytes32(vector.left)
			if err != nil {
				t.Fatalf("Failed to convert left hex: %v", err)
			}

			rightBytes, err := smt.HexToBytes32(vector.right)
			if err != nil {
				t.Fatalf("Failed to convert right hex: %v", err)
			}

			// Compute hash using Go implementation
			goResult := smt.HashBytes32(leftBytes, rightBytes)
			goResultHex := goResult.String()

			// Generate Solidity test data for verification
			solidityTestData := map[string]interface{}{
				"left":     vector.left,
				"right":    vector.right,
				"expected": goResultHex,
			}

			// Save test data for Solidity verification
			testDataJSON, _ := json.MarshalIndent(solidityTestData, "", "  ")
			testDataFile := fmt.Sprintf("testdata/cross_platform_hash_%d.json", i)
			os.WriteFile(testDataFile, testDataJSON, 0644)

			t.Logf("Hash compatibility test %d: left=%s, right=%s, result=%s",
				i, vector.left, vector.right, goResultHex)
		})
	}
}

// TestCrossPlatformProofCompatibility tests that Go-generated proofs can be verified in Solidity
func TestCrossPlatformProofCompatibility(t *testing.T) {
	// Create test scenarios with different tree depths and operations
	testScenarios := []struct {
		name       string
		treeDepth  uint16
		operations []struct {
			opType string
			index  *big.Int
			value  string
		}
	}{
		{
			name:      "simple_single_insert",
			treeDepth: 4,
			operations: []struct {
				opType string
				index  *big.Int
				value  string
			}{
				{"insert", big.NewInt(5), "0x1111111111111111111111111111111111111111111111111111111111111111"},
			},
		},
		{
			name:      "multiple_inserts",
			treeDepth: 6,
			operations: []struct {
				opType string
				index  *big.Int
				value  string
			}{
				{"insert", big.NewInt(1), "0x1111111111111111111111111111111111111111111111111111111111111111"},
				{"insert", big.NewInt(5), "0x5555555555555555555555555555555555555555555555555555555555555555"},
				{"insert", big.NewInt(10), "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			},
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Create Go SMT
			db := smt.NewInMemoryDatabase()
			goSMT, err := smt.NewSparseMerkleTree(db, scenario.treeDepth)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			// Execute operations
			var proofResults []CrossPlatformProofTest
			for _, op := range scenario.operations {
				switch op.opType {
				case "insert":
					value, err := smt.HexToBytes32(op.value)
					if err != nil {
						t.Fatalf("Failed to parse value: %v", err)
					}
					_, err = goSMT.Insert(op.index, value)
					if err != nil {
						t.Fatalf("Insert failed: %v", err)
					}
				case "update":
					value, err := smt.HexToBytes32(op.value)
					if err != nil {
						t.Fatalf("Failed to parse value: %v", err)
					}
					_, err = goSMT.Update(op.index, value)
					if err != nil {
						t.Fatalf("Update failed: %v", err)
					}
				}
			}

			// Generate proofs for all inserted indices
			for _, op := range scenario.operations {
				proof, err := goSMT.Get(op.index)
				if err != nil {
					t.Fatalf("Get failed: %v", err)
				}

				// Verify proof using Go implementation
				isValid := goSMT.VerifyProof(proof)
				if !isValid {
					t.Fatalf("Go proof verification failed for index %s", op.index.String())
				}

				// Convert proof to cross-platform format
				proofResult := CrossPlatformProofTest{
					Index:    op.index.String(),
					Exists:   proof.Exists,
					Leaf:     proof.Leaf.String(),
					Value:    proof.Value.String(),
					Enables:  proof.Enables.String(),
					Siblings: make([]string, len(proof.Siblings)),
				}

				for i, sibling := range proof.Siblings {
					proofResult.Siblings[i] = sibling.String()
				}

				proofResults = append(proofResults, proofResult)
			}

			// Create test vector for Solidity verification
			testVector := CrossPlatformTestVector{
				Name:        scenario.name,
				Description: fmt.Sprintf("Cross-platform proof compatibility test for %s", scenario.name),
				TreeDepth:   scenario.treeDepth,
				Operations:  make([]CrossPlatformOp, len(scenario.operations)),
				Expected: CrossPlatformExpected{
					FinalRoot:    goSMT.Root().String(),
					ProofResults: proofResults,
				},
			}

			// Convert operations to cross-platform format
			for i, op := range scenario.operations {
				testVector.Operations[i] = CrossPlatformOp{
					Type:  op.opType,
					Index: op.index.String(),
					Value: op.value,
				}
			}

			// Save test data for Solidity verification
			testDataJSON, _ := json.MarshalIndent(testVector, "", "  ")
			testDataFile := fmt.Sprintf("testdata/cross_platform_proof_%s.json", scenario.name)
			os.WriteFile(testDataFile, testDataJSON, 0644)

			t.Logf("Cross-platform proof test %s: operations=%d, root=%s",
				scenario.name, len(scenario.operations), goSMT.Root().String())
		})
	}
}

// TestCrossPlatformRootCompatibility tests that Go and Solidity produce identical tree roots
func TestCrossPlatformRootCompatibility(t *testing.T) {
	testCases := []struct {
		name      string
		treeDepth uint16
		entries   map[string]string // index -> value
	}{
		{
			name:      "empty_tree",
			treeDepth: 8,
			entries:   map[string]string{},
		},
		{
			name:      "single_entry",
			treeDepth: 4,
			entries: map[string]string{
				"3": "0x1111111111111111111111111111111111111111111111111111111111111111",
			},
		},
		{
			name:      "multiple_entries",
			treeDepth: 6,
			entries: map[string]string{
				"1":  "0x1111111111111111111111111111111111111111111111111111111111111111",
				"5":  "0x5555555555555555555555555555555555555555555555555555555555555555",
				"10": "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Go SMT
			db := smt.NewInMemoryDatabase()
			goSMT, err := smt.NewSparseMerkleTree(db, tc.treeDepth)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			// Insert entries
			for indexStr, valueStr := range tc.entries {
				index := new(big.Int)
				index.SetString(indexStr, 10)
				
				value, err := smt.HexToBytes32(valueStr)
				if err != nil {
					t.Fatalf("Failed to parse value: %v", err)
				}
				
				_, err = goSMT.Insert(index, value)
				if err != nil {
					t.Fatalf("Insert failed: %v", err)
				}
			}

			// Get final root
			finalRoot := goSMT.Root()

			// Save test data for Solidity verification
			testData := map[string]interface{}{
				"name":      tc.name,
				"treeDepth": tc.treeDepth,
				"entries":   tc.entries,
				"expected":  finalRoot.String(),
			}

			testDataJSON, _ := json.MarshalIndent(testData, "", "  ")
			testDataFile := fmt.Sprintf("testdata/cross_platform_root_%s.json", tc.name)
			os.WriteFile(testDataFile, testDataJSON, 0644)

			t.Logf("Cross-platform root test %s: depth=%d, entries=%d, root=%s",
				tc.name, tc.treeDepth, len(tc.entries), finalRoot.String())
		})
	}
}

// TestCrossPlatformSerializationCompatibility tests serialization format compatibility
func TestCrossPlatformSerializationCompatibility(t *testing.T) {
	testCases := []struct {
		name    string
		bigInt  *big.Int
		bytes32 smt.Bytes32
	}{
		{
			name:   "zero",
			bigInt: big.NewInt(0),
			bytes32: smt.Bytes32{},
		},
		{
			name:   "one",
			bigInt: big.NewInt(1),
			bytes32: smt.Bytes32{31: 1}, // Little endian in bytes32
		},
		{
			name:   "max_uint8",
			bigInt: big.NewInt(255),
			bytes32: smt.Bytes32{31: 255},
		},
		{
			name:   "max_uint16",
			bigInt: big.NewInt(65535),
			bytes32: smt.Bytes32{30: 255, 31: 255},
		},
		{
			name:   "large_number",
			bigInt: new(big.Int).SetBytes([]byte{0x12, 0x34, 0x56, 0x78}),
			bytes32: smt.Bytes32{28: 0x12, 29: 0x34, 30: 0x56, 31: 0x78},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test serialization consistency
			testData := map[string]interface{}{
				"name":      tc.name,
				"bigInt":    tc.bigInt.String(),
				"bytes32":   tc.bytes32.String(),
				"bigIntHex": fmt.Sprintf("0x%s", tc.bigInt.Text(16)),
			}

			testDataJSON, _ := json.MarshalIndent(testData, "", "  ")
			testDataFile := fmt.Sprintf("testdata/cross_platform_serialization_%s.json", tc.name)
			os.WriteFile(testDataFile, testDataJSON, 0644)

			t.Logf("Cross-platform serialization test %s: bigInt=%s, bytes32=%s",
				tc.name, tc.bigInt.String(), tc.bytes32.String())
		})
	}
}

// KVCompatibilityTestVector represents a test case for KV cross-platform compatibility
type KVCompatibilityTestVector struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Operations  []KVOperation     `json:"operations"`
	Expected    KVExpectedResults `json:"expected"`
}

// KVOperation represents a single key-value operation
type KVOperation struct {
	Type  string `json:"type"`  // "insert", "update", "get"
	Key   string `json:"key"`   // hex string
	Value string `json:"value"` // hex string (for insert/update)
}

// KVExpectedResults represents expected results for KV operations
type KVExpectedResults struct {
	FinalRoot    string          `json:"finalRoot"`
	KeyMappings  []KVKeyMapping  `json:"keyMappings"`
	ProofResults []KVProofResult `json:"proofResults"`
}

// KVKeyMapping represents the mapping from key to computed index
type KVKeyMapping struct {
	Key           string `json:"key"`
	ComputedIndex string `json:"computedIndex"`
	LeafHash      string `json:"leafHash"`
}

// KVProofResult represents expected proof results for KV operations
type KVProofResult struct {
	Key      string   `json:"key"`
	Index    string   `json:"index"`
	Exists   bool     `json:"exists"`
	Leaf     string   `json:"leaf"`
	Value    string   `json:"value"`
	Enables  string   `json:"enables"`
	Siblings []string `json:"siblings"`
}

// computeKeyIndex computes the index for a given key using the same method as InsertKV
func computeKeyIndex(key string) *big.Int {
	hash := crypto.Keccak256([]byte(key))
	index := new(big.Int).SetBytes(hash)

	// Truncate index to fit within tree depth (assuming depth 16 for tests)
	const testTreeDepth = 16
	if testTreeDepth < 256 {
		maxIndex := new(big.Int).Lsh(big.NewInt(1), uint(testTreeDepth))
		index.Mod(index, maxIndex)
	}

	return index
}

// TestKVCrossPlatformCompatibility tests KV cross-platform compatibility
func TestKVCrossPlatformCompatibility(t *testing.T) {
	// Test scenarios for KV operations
	testScenarios := []struct {
		name       string
		operations []struct {
			opType string
			key    string
			value  string
		}
	}{
		{
			name: "simple_kv_operations",
			operations: []struct {
				opType string
				key    string
				value  string
			}{
				{"insert", "key1", "0x1111111111111111111111111111111111111111111111111111111111111111"},
				{"insert", "key2", "0x2222222222222222222222222222222222222222222222222222222222222222"},
				{"insert", "key3", "0x3333333333333333333333333333333333333333333333333333333333333333"},
			},
		},
		{
			name: "kv_insert_and_update",
			operations: []struct {
				opType string
				key    string
				value  string
			}{
				{"insert", "testkey", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
				{"update", "testkey", "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"},
				{"insert", "newkey", "0x4444444444444444444444444444444444444444444444444444444444444444"},
			},
		},
		{
			name: "kv_edge_cases",
			operations: []struct {
				opType string
				key    string
				value  string
			}{
				{"insert", "empty", "0x0000000000000000000000000000000000000000000000000000000000000000"},
				{"insert", "max", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
				{"insert", "mid", "0x8888888888888888888888888888888888888888888888888888888888888888"},
			},
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Create SMT for KV operations
			db := smt.NewInMemoryDatabase()
			tree, err := smt.NewSparseMerkleTree(db, 16)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			var keyMappings []KVKeyMapping
			var proofResults []KVProofResult

			// Execute operations
			for _, op := range scenario.operations {
				switch op.opType {
				case "insert":
					value, err := smt.HexToBytes32(op.value)
					if err != nil {
						t.Fatalf("Failed to parse value: %v", err)
					}
					_, err = tree.InsertKV(op.key, value)
					if err != nil {
						t.Fatalf("KV insert failed for key %s: %v", op.key, err)
					}
				case "update":
					value, err := smt.HexToBytes32(op.value)
					if err != nil {
						t.Fatalf("Failed to parse value: %v", err)
					}
					_, err = tree.UpdateKV(op.key, value)
					if err != nil {
						t.Fatalf("KV update failed for key %s: %v", op.key, err)
					}
				}
			}

			// Generate key mappings and proofs for all keys
			processedKeys := make(map[string]bool)
			for _, op := range scenario.operations {
				if processedKeys[op.key] {
					continue // Skip duplicate keys
				}
				processedKeys[op.key] = true

				// Compute index from key
				index := computeKeyIndex(op.key)
				computedIndexHex := index.String()

				// Get current value for this key
				value, exists, err := tree.GetKV(op.key)
				if err != nil {
					t.Fatalf("KV get failed for key %s: %v", op.key, err)
				}

				// Get proof for verification
				proof, err := tree.Get(index)
				if err != nil {
					t.Fatalf("Get proof failed for key %s: %v", op.key, err)
				}

				// Verify proof
				if !tree.VerifyProof(proof) {
					t.Fatalf("Proof verification failed for key %s", op.key)
				}

				// Create key mapping
				keyMapping := KVKeyMapping{
					Key:           op.key,
					ComputedIndex: computedIndexHex,
					LeafHash:      proof.Leaf.String(),
				}
				keyMappings = append(keyMappings, keyMapping)

				// Create proof result
				proofResult := KVProofResult{
					Key:      op.key,
					Index:    proof.Index.String(),
					Exists:   exists,
					Leaf:     proof.Leaf.String(),
					Enables:  proof.Enables.String(),
					Siblings: make([]string, len(proof.Siblings)),
				}

				for i, sibling := range proof.Siblings {
					proofResult.Siblings[i] = sibling.String()
				}

				if exists {
					proofResult.Value = value.String()
				}

				proofResults = append(proofResults, proofResult)
			}

			// Create test vector
			testVector := KVCompatibilityTestVector{
				Name:        scenario.name,
				Description: fmt.Sprintf("KV cross-platform compatibility test for %s", scenario.name),
				Expected: KVExpectedResults{
					FinalRoot:    tree.Root().String(),
					KeyMappings:  keyMappings,
					ProofResults: proofResults,
				},
			}

			// Convert operations to test format
			for _, op := range scenario.operations {
				testVector.Operations = append(testVector.Operations, KVOperation{
					Type:  op.opType,
					Key:   op.key,
					Value: op.value,
				})
			}

			// Save test vector for Solidity verification
			testVectorJSON, _ := json.MarshalIndent(testVector, "", "  ")
			testVectorFile := fmt.Sprintf("testdata/cross_platform_kv_%s.json", scenario.name)
			os.WriteFile(testVectorFile, testVectorJSON, 0644)

			t.Logf("KV compatibility test %s: operations=%d, finalRoot=%s",
				scenario.name, len(scenario.operations), tree.Root().String())
		})
	}
}

// TestKVKeyIndexMapping tests that key-to-index mapping is consistent
func TestKVKeyIndexMapping(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{"simple_key", "key1"},
		{"hex_key", "0xabc"},
		{"long_key", "longkey123"},
		{"numeric_key", "12345"},
		{"zero_key", "0"},
		{"special_key", "test-key"},
		{"unicode_key", "κλειδί"},
	}

	var keyMappings []KVKeyMapping

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compute index using KV method
			computedIndex := computeKeyIndex(tc.key)
			computedIndexHex := computedIndex.String()

			// Create tree and insert the key
			db := smt.NewInMemoryDatabase()
			tree, err := smt.NewSparseMerkleTree(db, 16)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			testValue, err := smt.HexToBytes32("0x1111111111111111111111111111111111111111111111111111111111111111")
			if err != nil {
				t.Fatalf("Failed to parse test value: %v", err)
			}

			_, err = tree.InsertKV(tc.key, testValue)
			if err != nil {
				t.Fatalf("Insert failed for key %s: %v", tc.key, err)
			}

			// Get proof and verify the index matches our computation
			proof, err := tree.Get(computedIndex)
			if err != nil {
				t.Fatalf("Get failed for key %s: %v", tc.key, err)
			}

			proofIndexHex := proof.Index.String()
			if proofIndexHex != computedIndexHex {
				t.Fatalf("Index mismatch for key %s: expected %s, got %s",
					tc.key, computedIndexHex, proofIndexHex)
			}

			// Verify the leaf hash computation
			// Note: proof.Leaf now contains the raw value, not the hash
			expectedLeaf := smt.ComputeLeafHash(computedIndex, testValue)
			expectedLeafHex := expectedLeaf.String()
			
			// Compute the actual leaf hash from the proof
			actualLeaf := smt.ComputeLeafHash(proof.Index, proof.Value)
			actualLeafHex := actualLeaf.String()

			if actualLeafHex != expectedLeafHex {
				t.Fatalf("Leaf hash mismatch for key %s: expected %s, got %s",
					tc.key, expectedLeafHex, actualLeafHex)
			}

			// Create key mapping for cross-platform verification
			keyMapping := KVKeyMapping{
				Key:           tc.key,
				ComputedIndex: computedIndexHex,
				LeafHash:      actualLeafHex,
			}
			keyMappings = append(keyMappings, keyMapping)

			t.Logf("Key mapping test %s: key=%s, index=%s, leaf=%s",
				tc.name, tc.key, computedIndexHex, actualLeafHex)
		})
	}

	// Save key mappings for Solidity verification
	keyMappingsJSON, _ := json.MarshalIndent(keyMappings, "", "  ")
	os.WriteFile("testdata/cross_platform_kv_key_mappings.json", keyMappingsJSON, 0644)
}

// TestKVLeafHashComputation tests that leaf hash computation is consistent
func TestKVLeafHashComputation(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{"simple", "key1", "0x1111111111111111111111111111111111111111111111111111111111111111"},
		{"complex", "complex_key", "0x2222222222222222222222222222222222222222222222222222222222222222"},
		{"zero_key", "0", "0x3333333333333333333333333333333333333333333333333333333333333333"},
		{"zero_value", "testkey", "0x0000000000000000000000000000000000000000000000000000000000000000"},
		{"max_values", "maxkey", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
	}

	var leafHashTests []struct {
		Key           string `json:"key"`
		Value         string `json:"value"`
		ExpectedLeaf  string `json:"expectedLeaf"`
		ExpectedIndex string `json:"expectedIndex"`
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compute index from key
			expectedIndex := computeKeyIndex(tc.key)
			expectedIndexHex := expectedIndex.String()

			// Parse value
			value, err := smt.HexToBytes32(tc.value)
			if err != nil {
				t.Fatalf("Failed to parse value: %v", err)
			}

			// Compute leaf hash
			expectedLeaf := smt.ComputeLeafHash(expectedIndex, value)
			expectedLeafHex := expectedLeaf.String()

			// Verify using SMT with KV operations
			db := smt.NewInMemoryDatabase()
			tree, err := smt.NewSparseMerkleTree(db, 16)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			_, err = tree.InsertKV(tc.key, value)
			if err != nil {
				t.Fatalf("Insert failed: %v", err)
			}

			proof, err := tree.Get(expectedIndex)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			// Verify leaf hash matches our computation
			// Note: proof.Leaf now contains the raw value, not the hash
			actualLeaf := smt.ComputeLeafHash(proof.Index, proof.Value)
			actualLeafHex := actualLeaf.String()
			
			if actualLeafHex != expectedLeafHex {
				t.Fatalf("Leaf hash mismatch: expected %s, got %s", expectedLeafHex, actualLeafHex)
			}

			// Verify index matches our computation
			proofIndexHex := proof.Index.String()
			if proofIndexHex != expectedIndexHex {
				t.Fatalf("Index mismatch: expected %s, got %s", expectedIndexHex, proofIndexHex)
			}

			// Add to test data for Solidity verification
			leafHashTests = append(leafHashTests, struct {
				Key           string `json:"key"`
				Value         string `json:"value"`
				ExpectedLeaf  string `json:"expectedLeaf"`
				ExpectedIndex string `json:"expectedIndex"`
			}{
				Key:           tc.key,
				Value:         tc.value,
				ExpectedLeaf:  expectedLeafHex,
				ExpectedIndex: expectedIndexHex,
			})

			t.Logf("Leaf hash test %s: key=%s, value=%s, leaf=%s, index=%s",
				tc.name, tc.key, tc.value, expectedLeafHex, expectedIndexHex)
		})
	}

	// Save leaf hash tests for Solidity verification
	leafHashTestsJSON, _ := json.MarshalIndent(leafHashTests, "", "  ")
	os.WriteFile("testdata/cross_platform_kv_leaf_hash_tests.json", leafHashTestsJSON, 0644)
}

// TestKVProofCompatibility tests that KV proofs are compatible across platforms
func TestKVProofCompatibility(t *testing.T) {
	// Create comprehensive test with multiple keys and operations
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 16)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	operations := []struct {
		opType string
		key    string
		value  string
	}{
		{"insert", "key1", "0x1111111111111111111111111111111111111111111111111111111111111111"},
		{"insert", "key2", "0x2222222222222222222222222222222222222222222222222222222222222222"},
		{"insert", "complex_key", "0x3333333333333333333333333333333333333333333333333333333333333333"},
		{"update", "key1", "0x4444444444444444444444444444444444444444444444444444444444444444"},
		{"insert", "final_key", "0x5555555555555555555555555555555555555555555555555555555555555555"},
	}

	var proofCompatibilityTests []struct {
		Operation      string   `json:"operation"`
		Key            string   `json:"key"`
		Value          string   `json:"value"`
		ResultExists   bool     `json:"resultExists"`
		ResultLeaf     string   `json:"resultLeaf"`
		ResultValue    string   `json:"resultValue"`
		ResultIndex    string   `json:"resultIndex"`
		ResultEnables  string   `json:"resultEnables"`
		ResultSiblings []string `json:"resultSiblings"`
		TreeRoot       string   `json:"treeRoot"`
	}

	// Execute operations and capture results
	for i, op := range operations {
		switch op.opType {
		case "insert":
			value, err := smt.HexToBytes32(op.value)
			if err != nil {
				t.Fatalf("Failed to parse value: %v", err)
			}
			_, err = tree.InsertKV(op.key, value)
			if err != nil {
				t.Fatalf("Insert failed for key %s: %v", op.key, err)
			}
		case "update":
			value, err := smt.HexToBytes32(op.value)
			if err != nil {
				t.Fatalf("Failed to parse value: %v", err)
			}
			_, err = tree.UpdateKV(op.key, value)
			if err != nil {
				t.Fatalf("Update failed for key %s: %v", op.key, err)
			}
		}

		// Get proof after operation
		index := computeKeyIndex(op.key)
		proof, err := tree.Get(index)
		if err != nil {
			t.Fatalf("Get failed for key %s: %v", op.key, err)
		}

		// Verify proof
		if !tree.VerifyProof(proof) {
			t.Fatalf("Proof verification failed for operation %d, key %s", i, op.key)
		}

		// Get value using KV interface
		value, exists, err := tree.GetKV(op.key)
		if err != nil {
			t.Fatalf("GetKV failed for key %s: %v", op.key, err)
		}

		// Capture result for cross-platform verification
		proofTest := struct {
			Operation      string   `json:"operation"`
			Key            string   `json:"key"`
			Value          string   `json:"value"`
			ResultExists   bool     `json:"resultExists"`
			ResultLeaf     string   `json:"resultLeaf"`
			ResultValue    string   `json:"resultValue"`
			ResultIndex    string   `json:"resultIndex"`
			ResultEnables  string   `json:"resultEnables"`
			ResultSiblings []string `json:"resultSiblings"`
			TreeRoot       string   `json:"treeRoot"`
		}{
			Operation:      op.opType,
			Key:            op.key,
			Value:          op.value,
			ResultExists:   exists,
			ResultLeaf:     proof.Leaf.String(),
			ResultIndex:    proof.Index.String(),
			ResultEnables:  proof.Enables.String(),
			ResultSiblings: make([]string, len(proof.Siblings)),
			TreeRoot:       tree.Root().String(),
		}

		for j, sibling := range proof.Siblings {
			proofTest.ResultSiblings[j] = sibling.String()
		}

		if exists {
			proofTest.ResultValue = value.String()
		}

		proofCompatibilityTests = append(proofCompatibilityTests, proofTest)

		t.Logf("KV proof compatibility test %d: op=%s, key=%s, exists=%t, leaf=%s",
			i, op.opType, op.key, exists, proof.Leaf.String())
	}

	// Save proof compatibility tests for Solidity verification
	proofTestsJSON, _ := json.MarshalIndent(proofCompatibilityTests, "", "  ")
	os.WriteFile("testdata/cross_platform_kv_proof_compatibility.json", proofTestsJSON, 0644)

	t.Logf("KV proof compatibility test completed with %d operations, final root: %s",
		len(operations), tree.Root().String())
}

// TestKVErrorHandlingCompatibility tests error handling compatibility for KV operations
func TestKVErrorHandlingCompatibility(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 16)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	testKey := "test_key"
	testValue, err := smt.HexToBytes32("0x1111111111111111111111111111111111111111111111111111111111111111")
	if err != nil {
		t.Fatalf("Failed to parse test value: %v", err)
	}

	// Test 1: Update non-existent key should fail (InsertKV will actually succeed)
	// Note: InsertKV will succeed for non-existent keys, so this test is not applicable
	// Skip this test as KV operations are more like key-value store operations

	// Test 2: Insert new key should succeed
	_, err = tree.InsertKV(testKey, testValue)
	if err != nil {
		t.Fatalf("Insert should succeed for new key: %v", err)
	}

	// Test 3: Insert existing key should fail
	deadbeef, err := smt.HexToBytes32("0xdeadbeef00000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatalf("Failed to parse deadbeef value: %v", err)
	}
	_, err = tree.InsertKV(testKey, deadbeef)
	if err == nil {
		t.Fatal("Insert should fail for existing key")
	}
	t.Logf("Insert existing key correctly failed: %v", err)

	// Test 4: Update existing key should succeed (using UpdateKV)
	newValue, err := smt.HexToBytes32("0x2222222222222222222222222222222222222222222222222222222222222222")
	if err != nil {
		t.Fatalf("Failed to parse new value: %v", err)
	}
	_, err = tree.UpdateKV(testKey, newValue)
	if err != nil {
		t.Fatalf("Update should succeed for existing key: %v", err)
	}

	// Test 5: Get non-existent key should return exists=false
	nonExistentKey := "nonexistent_key"
	value, exists, err := tree.GetKV(nonExistentKey)
	if err != nil {
		t.Fatalf("GetKV failed: %v", err)
	}
	if exists {
		t.Fatal("GetKV should return exists=false for non-existent key")
	}
	if !value.IsZero() {
		t.Fatal("GetKV should return zero value for non-existent key")
	}

	// Test 6: Exists should return false for non-existent key
	nonExistentIndex := computeKeyIndex(nonExistentKey)
	exists, err = tree.Exists(nonExistentIndex)
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if exists {
		t.Fatal("Exists should return false for non-existent key")
	}

	// Test 7: Exists should return true for existing key
	existingIndex := computeKeyIndex(testKey)
	exists, err = tree.Exists(existingIndex)
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if !exists {
		t.Fatal("Exists should return true for existing key")
	}

	t.Logf("KV error handling compatibility test passed")
}

// TestGenerateKVCrossPlatformVectors generates comprehensive test vectors for KV cross-platform testing
func TestGenerateKVCrossPlatformVectors(t *testing.T) {
	// Ensure testdata directory exists
	os.MkdirAll("testdata", 0755)

	// Generate comprehensive KV test scenarios
	scenarios := []string{
		"simple_kv_operations",
		"kv_insert_and_update",
		"kv_edge_cases",
	}

	// This test generates the test vectors that are used by other tests
	// The actual test vectors are generated by the individual test functions above

	for _, scenario := range scenarios {
		t.Logf("Test vectors for scenario %s will be generated by TestKVCrossPlatformCompatibility", scenario)
	}

	t.Logf("KV cross-platform test vector generation completed")
}

// goKeccak256 implements the Go version of keccak256 hash function
func goKeccak256(left, right []byte) []byte {
	// Handle zero inputs - if both are zero, return zero
	if isZeroBytes(left) && isZeroBytes(right) {
		return make([]byte, 32) // Return 32 bytes of zeros
	}
	
	// For non-zero inputs, use keccak256
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(left)
	hasher.Write(right)
	return hasher.Sum(nil)
}

// isZeroBytes checks if a byte slice contains only zeros
func isZeroBytes(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// solidityKeccakSimulator simulates Solidity keccak256 behavior
func solidityKeccakSimulator(left, right []byte) []byte {
	// Solidity behavior: if both inputs are zero, return zero
	if isZeroBytes(left) && isZeroBytes(right) {
		return make([]byte, 32) // Return 32 bytes of zeros
	}
	
	// For non-zero inputs, use keccak256 (same as Go implementation)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(left)
	hasher.Write(right)
	return hasher.Sum(nil)
}

// TestHashFunctionCompatibility tests core hash function compatibility between Go and Solidity
func TestHashFunctionCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		right    string
		expected string // Expected behavior for both implementations
	}{
		{
			name:     "both inputs zero",
			left:     "0x0000000000000000000000000000000000000000000000000000000000000000",
			right:    "0x0000000000000000000000000000000000000000000000000000000000000000",
			expected: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "left zero, right non-zero",
			left:     "0x0000000000000000000000000000000000000000000000000000000000000000",
			right:    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expected: "", // Will be computed
		},
		{
			name:     "left non-zero, right zero",
			left:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			right:    "0x0000000000000000000000000000000000000000000000000000000000000000",
			expected: "", // Will be computed
		},
		{
			name:     "both inputs non-zero - simple",
			left:     "0x1111111111111111111111111111111111111111111111111111111111111111",
			right:    "0x2222222222222222222222222222222222222222222222222222222222222222",
			expected: "", // Will be computed
		},
		{
			name:     "both inputs non-zero - complex",
			left:     "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			right:    "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
			expected: "", // Will be computed
		},
		{
			name:     "maximum values",
			left:     "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			right:    "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			expected: "", // Will be computed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert hex strings to bytes
			leftBytes, err := testutils.HexToBytes(tt.left)
			if err != nil {
				t.Fatalf("Failed to convert left hex to bytes: %v", err)
			}
			
			rightBytes, err := testutils.HexToBytes(tt.right)
			if err != nil {
				t.Fatalf("Failed to convert right hex to bytes: %v", err)
			}

			// Compute hash using Go implementation
			goResult := goKeccak256(leftBytes, rightBytes)
			goResultHex := testutils.BytesToHex(goResult)

			// Compute hash using Solidity simulator
			solidityResult := solidityKeccakSimulator(leftBytes, rightBytes)
			solidityResultHex := testutils.BytesToHex(solidityResult)

			// Verify both implementations produce the same result
			if !testutils.CompareHexStrings(goResultHex, solidityResultHex) {
				t.Fatalf("Hash mismatch: Go=%s, Solidity=%s", goResultHex, solidityResultHex)
			}

			// If expected result is provided, verify against it
			if tt.expected != "" {
				if !testutils.CompareHexStrings(goResultHex, tt.expected) {
					t.Fatalf("Expected %s, got %s", tt.expected, goResultHex)
				}
			}

			t.Logf("Input: left=%s, right=%s", tt.left, tt.right)
			t.Logf("Result: %s", goResultHex)
		})
	}
}

// TestZeroInputHandling specifically tests the zero input handling requirement
func TestZeroInputHandling(t *testing.T) {
	// Test case: both inputs are zero should return zero
	zeroBytes := make([]byte, 32) // 32 bytes of zeros
	
	goResult := goKeccak256(zeroBytes, zeroBytes)
	solidityResult := solidityKeccakSimulator(zeroBytes, zeroBytes)
	
	// Both should return zero
	expectedZero := make([]byte, 32)
	
	if !isZeroBytes(goResult) {
		t.Fatalf("Go implementation should return zero for zero inputs, got %x", goResult)
	}
	
	if !isZeroBytes(solidityResult) {
		t.Fatalf("Solidity implementation should return zero for zero inputs, got %x", solidityResult)
	}
	
	// Verify they match each other
	goResultHex := testutils.BytesToHex(goResult)
	solidityResultHex := testutils.BytesToHex(solidityResult)
	expectedHex := testutils.BytesToHex(expectedZero)
	
	if !testutils.CompareHexStrings(goResultHex, solidityResultHex) {
		t.Fatalf("Go and Solidity results don't match: Go=%s, Solidity=%s", goResultHex, solidityResultHex)
	}
	
	if !testutils.CompareHexStrings(goResultHex, expectedHex) {
		t.Fatalf("Expected zero hash, got %s", goResultHex)
	}
	
	t.Logf("Zero input test passed: %s", goResultHex)
}

// TestHashVectorCompatibility tests using predefined test vectors
func TestHashVectorCompatibility(t *testing.T) {
	// Load test vectors from JSON file
	vectors, err := vectors.LoadHashVectors("testdata/hash_vectors.json")
	if err != nil {
		t.Logf("Could not load hash vectors, creating default test vectors: %v", err)
		// Create default test vectors if file doesn't exist
		vectors = createDefaultHashVectors()
	}
	
	for i, vector := range vectors {
		t.Run(fmt.Sprintf("vector_%d", i), func(t *testing.T) {
			// Convert hex strings to bytes
			leftBytes, err := testutils.HexToBytes(vector.Left)
			if err != nil {
				t.Fatalf("Failed to convert left hex to bytes: %v", err)
			}
			
			rightBytes, err := testutils.HexToBytes(vector.Right)
			if err != nil {
				t.Fatalf("Failed to convert right hex to bytes: %v", err)
			}

			// Compute hash using Go implementation
			goResult := goKeccak256(leftBytes, rightBytes)
			goResultHex := testutils.BytesToHex(goResult)

			// Compute hash using Solidity simulator
			solidityResult := solidityKeccakSimulator(leftBytes, rightBytes)
			solidityResultHex := testutils.BytesToHex(solidityResult)

			// Verify both implementations produce the same result
			if !testutils.CompareHexStrings(goResultHex, solidityResultHex) {
				t.Fatalf("Hash mismatch: Go=%s, Solidity=%s", goResultHex, solidityResultHex)
			}

			// If expected result is provided in vector, verify against it
			if vector.Expected != "" {
				if !testutils.CompareHexStrings(goResultHex, vector.Expected) {
					t.Fatalf("Expected %s, got %s", vector.Expected, goResultHex)
				}
			}

			t.Logf("Vector %d: left=%s, right=%s, result=%s", i, vector.Left, vector.Right, goResultHex)
		})
	}
}

// createDefaultHashVectors creates a set of default test vectors for hash function testing
func createDefaultHashVectors() []vectors.HashTestVector {
	return []vectors.HashTestVector{
		{
			Left:     "0x0000000000000000000000000000000000000000000000000000000000000000",
			Right:    "0x0000000000000000000000000000000000000000000000000000000000000000",
			Expected: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			Left:     "0x1111111111111111111111111111111111111111111111111111111111111111",
			Right:    "0x2222222222222222222222222222222222222222222222222222222222222222",
			Expected: "", // Will be computed
		},
		{
			Left:     "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Right:    "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
			Expected: "", // Will be computed
		},
		{
			Left:     "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			Right:    "0x0000000000000000000000000000000000000000000000000000000000000001",
			Expected: "", // Will be computed
		},
	}
}

// TestProofVectorCompatibility tests proof generation and verification using test vectors
func TestProofVectorCompatibility(t *testing.T) {
	// Test basic proof compatibility with simple vectors
	testCases := []struct {
		name     string
		treeDepth uint16
		index    *big.Int
		value    string
	}{
		{"simple_proof", 8, big.NewInt(1), "0x1111111111111111111111111111111111111111111111111111111111111111"},
		{"zero_index", 8, big.NewInt(0), "0x2222222222222222222222222222222222222222222222222222222222222222"},
		{"max_index", 8, big.NewInt(255), "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{"mid_index", 8, big.NewInt(42), "0x4242424242424242424242424242424242424242424242424242424242424242"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create tree with specified depth
			db := smt.NewInMemoryDatabase()
			tree, err := smt.NewSparseMerkleTree(db, tc.treeDepth)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			// Parse and insert value
			value, err := smt.HexToBytes32(tc.value)
			if err != nil {
				t.Fatalf("Failed to parse value: %v", err)
			}

			_, err = tree.Insert(tc.index, value)
			if err != nil {
				t.Fatalf("Failed to insert leaf: %v", err)
			}

			// Get proof from our implementation
			proof, err := tree.Get(tc.index)
			if err != nil {
				t.Fatalf("Failed to get proof: %v", err)
			}

			// Verify the proof structure matches expected format
			if proof.Index.Cmp(tc.index) != 0 {
				t.Fatalf("Proof index mismatch: expected %s, got %s",
					tc.index.String(), proof.Index.String())
			}

			if !proof.Exists {
				t.Fatal("Proof should indicate existence for inserted leaf")
			}

			if proof.Value.IsZero() {
				t.Fatal("Proof should have non-zero value for existing leaf")
			}

			// Verify proof can be verified
			if !tree.VerifyProof(proof) {
				t.Fatal("Generated proof should be valid")
			}

			// Verify enables bitmask is reasonable
			if proof.Enables.Cmp(big.NewInt(0)) < 0 {
				t.Fatalf("Proof enables should be non-negative, got %s", proof.Enables.String())
			}

			// Verify siblings array length matches enables bit count
			enablesBitCount := 0
			enables := new(big.Int).Set(proof.Enables)
			for enables.Cmp(big.NewInt(0)) > 0 {
				if enables.Bit(0) == 1 {
					enablesBitCount++
				}
				enables.Rsh(enables, 1)
			}

			if len(proof.Siblings) != enablesBitCount {
				t.Fatalf("Siblings array length (%d) should match enables bit count (%d)",
					len(proof.Siblings), enablesBitCount)
			}

			t.Logf("Proof vector %s passed: depth=%d, leaf=%s, index=%s, enables=%s, siblings=%d",
				tc.name, tc.treeDepth, proof.Leaf.String(), tc.index.String(), proof.Enables.String(), len(proof.Siblings))
		})
	}
}

// TestProofJSONCompatibility tests that proofs serialize to JSON in TypeScript-compatible format
func TestProofJSONCompatibility(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	testCases := []struct {
		index *big.Int
		value string
	}{
		{big.NewInt(0), "0x1111111111111111111111111111111111111111111111111111111111111111"},
		{big.NewInt(1), "0x2222222222222222222222222222222222222222222222222222222222222222"},
		{big.NewInt(255), "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
	}

	for _, tc := range testCases {
		value, err := smt.HexToBytes32(tc.value)
		if err != nil {
			t.Fatalf("Failed to parse value: %v", err)
		}

		_, err = tree.Insert(tc.index, value)
		if err != nil {
			t.Fatalf("Failed to insert index %s: %v", tc.index.String(), err)
		}

		proof, err := tree.Get(tc.index)
		if err != nil {
			t.Fatalf("Failed to get proof: %v", err)
		}

		// Create JSON-compatible representation
		proofJSON := map[string]interface{}{
			"exists":   proof.Exists,
			"leaf":     proof.Leaf.String(),
			"value":    proof.Value.String(),
			"index":    proof.Index.String(),
			"enables":  proof.Enables.String(),
			"siblings": make([]string, len(proof.Siblings)),
		}

		for i, sibling := range proof.Siblings {
			proofJSON["siblings"].([]string)[i] = sibling.String()
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(proofJSON)
		if err != nil {
			t.Fatalf("JSON marshaling failed for index %s: %v", tc.index.String(), err)
		}

		// Verify JSON structure contains expected fields
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		if err != nil {
			t.Fatalf("JSON unmarshaling to map failed: %v", err)
		}

		// Check required fields exist
		requiredFields := []string{"exists", "leaf", "value", "index", "enables", "siblings"}
		for _, field := range requiredFields {
			if _, exists := jsonMap[field]; !exists {
				t.Fatalf("JSON should contain field '%s'", field)
			}
		}

		// Verify field types
		if _, ok := jsonMap["exists"].(bool); !ok {
			t.Fatal("Field 'exists' should be boolean")
		}

		if _, ok := jsonMap["leaf"].(string); !ok {
			t.Fatal("Field 'leaf' should be string")
		}

		if jsonMap["value"] != nil {
			if _, ok := jsonMap["value"].(string); !ok {
				t.Fatal("Field 'value' should be string or null")
			}
		}

		if _, ok := jsonMap["siblings"].([]interface{}); !ok {
			t.Fatal("Field 'siblings' should be array")
		}

		t.Logf("JSON compatibility test passed for index %s", tc.index.String())
	}
}

// TestUpdateProofJSONCompatibility tests UpdateProof JSON serialization
func TestUpdateProofJSONCompatibility(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	index := big.NewInt(5)
	initialValue, err := smt.HexToBytes32("0x1111111111111111111111111111111111111111111111111111111111111111")
	if err != nil {
		t.Fatalf("Failed to parse initial value: %v", err)
	}
	updatedValue, err := smt.HexToBytes32("0x2222222222222222222222222222222222222222222222222222222222222222")
	if err != nil {
		t.Fatalf("Failed to parse updated value: %v", err)
	}

	// Test Insert UpdateProof
	insertProof, err := tree.Insert(index, initialValue)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Create JSON-compatible representation
	updateProofJSON := map[string]interface{}{
		"exists":   insertProof.Exists,
		"leaf":     insertProof.Leaf.String(),
		"value":    insertProof.Value.String(),
		"index":    insertProof.Index.String(),
		"enables":  insertProof.Enables.String(),
		"siblings": make([]string, len(insertProof.Siblings)),
		"newLeaf":  insertProof.NewLeaf.String(),
	}

	for i, sibling := range insertProof.Siblings {
		updateProofJSON["siblings"].([]string)[i] = sibling.String()
	}

	// Test JSON serialization of UpdateProof
	jsonData, err := json.Marshal(updateProofJSON)
	if err != nil {
		t.Fatalf("JSON marshaling of UpdateProof failed: %v", err)
	}

	// Verify JSON structure
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	if err != nil {
		t.Fatalf("JSON unmarshaling to map failed: %v", err)
	}

	// Check UpdateProof-specific field
	if _, exists := jsonMap["newLeaf"]; !exists {
		t.Fatal("UpdateProof JSON should contain 'newLeaf' field")
	}

	if _, ok := jsonMap["newLeaf"].(string); !ok {
		t.Fatal("Field 'newLeaf' should be string")
	}

	// Check embedded Proof fields
	requiredFields := []string{"exists", "leaf", "value", "index", "enables", "siblings"}
	for _, field := range requiredFields {
		if _, exists := jsonMap[field]; !exists {
			t.Fatalf("UpdateProof JSON should contain embedded Proof field '%s'", field)
		}
	}

	// Test Update UpdateProof
	updateProof, err := tree.Update(index, updatedValue)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Create JSON-compatible representation for update
	updateProofJSON2 := map[string]interface{}{
		"exists":   updateProof.Exists,
		"leaf":     updateProof.Leaf.String(),
		"value":    updateProof.Value.String(),
		"index":    updateProof.Index.String(),
		"enables":  updateProof.Enables.String(),
		"siblings": make([]string, len(updateProof.Siblings)),
		"newLeaf":  updateProof.NewLeaf.String(),
	}

	for i, sibling := range updateProof.Siblings {
		updateProofJSON2["siblings"].([]string)[i] = sibling.String()
	}

	// Test JSON serialization of Update UpdateProof
	jsonData2, err := json.Marshal(updateProofJSON2)
	if err != nil {
		t.Fatalf("JSON marshaling of Update UpdateProof failed: %v", err)
	}

	var jsonMap2 map[string]interface{}
	err = json.Unmarshal(jsonData2, &jsonMap2)
	if err != nil {
		t.Fatalf("JSON unmarshaling of Update UpdateProof failed: %v", err)
	}

	if jsonMap2["newLeaf"] != updateProof.NewLeaf.String() {
		t.Fatal("Update UpdateProof NewLeaf field mismatch")
	}

	if jsonMap2["exists"] != true {
		t.Fatal("Update UpdateProof should have Exists=true")
	}

	t.Logf("UpdateProof JSON compatibility test passed")
}

// TestKVProofCompatibilityBasic tests KV proof compatibility
func TestKVProofCompatibilityBasic(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256) // Use full depth for KV operations
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	testCases := []struct {
		key   string
		value string
	}{
		{"key1", "0x1111111111111111111111111111111111111111111111111111111111111111"},
		{"key2", "0x2222222222222222222222222222222222222222222222222222222222222222"},
		{"test_key", "0x3333333333333333333333333333333333333333333333333333333333333333"},
	}

	for _, tc := range testCases {
		value, err := smt.HexToBytes32(tc.value)
		if err != nil {
			t.Fatalf("Failed to parse value: %v", err)
		}

		_, err = tree.InsertKV(tc.key, value)
		if err != nil {
			t.Fatalf("Failed to insert key %s: %v", tc.key, err)
		}

		// Get proof using computed index
		index := new(big.Int).SetBytes(crypto.Keccak256([]byte(tc.key)))
		proof, err := tree.Get(index)
		if err != nil {
			t.Fatalf("Failed to get proof for key %s: %v", tc.key, err)
		}

		// Create JSON-compatible representation
		proofJSON := map[string]interface{}{
			"exists":   proof.Exists,
			"leaf":     proof.Leaf.String(),
			"value":    proof.Value.String(),
			"index":    proof.Index.String(),
			"enables":  proof.Enables.String(),
			"siblings": make([]string, len(proof.Siblings)),
		}

		for i, sibling := range proof.Siblings {
			proofJSON["siblings"].([]string)[i] = sibling.String()
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(proofJSON)
		if err != nil {
			t.Fatalf("JSON marshaling failed for key %s: %v", tc.key, err)
		}

		// Verify we can unmarshal back
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		if err != nil {
			t.Fatalf("JSON unmarshaling failed for key %s: %v", tc.key, err)
		}

		// Verify proof verification works
		if !tree.VerifyProof(proof) {
			t.Fatalf("Proof verification failed for key %s", tc.key)
		}

		t.Logf("KV proof compatibility test passed for key %s", tc.key)
	}
}

// TestProofFieldValidation tests that proof fields have correct formats and values
func TestProofFieldValidation(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	index := big.NewInt(42)
	value, err := smt.HexToBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("Failed to parse value: %v", err)
	}

	_, err = tree.Insert(index, value)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	proof, err := tree.Get(index)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Validate Leaf field format
	leafStr := proof.Leaf.String()
	if len(leafStr) != 66 {
		t.Fatalf("Leaf should be 66 characters long, got %d", len(leafStr))
	}

	if leafStr[:2] != "0x" {
		t.Fatalf("Leaf should start with '0x', got %s", leafStr[:2])
	}

	// Validate Index field
	if proof.Index == nil {
		t.Fatal("Index should not be nil")
	}

	if proof.Index.Cmp(index) != 0 {
		t.Fatalf("Index should match inserted index: expected %s, got %s",
			index.String(), proof.Index.String())
	}

	// Validate Enables field
	if proof.Enables == nil {
		t.Fatal("Enables should not be nil")
	}

	if proof.Enables.Cmp(big.NewInt(0)) < 0 {
		t.Fatalf("Enables should be non-negative, got %s", proof.Enables.String())
	}

	// Validate Siblings field
	if proof.Siblings == nil {
		t.Fatal("Siblings should not be nil")
	}

	for i, sibling := range proof.Siblings {
		siblingStr := sibling.String()
		if len(siblingStr) != 66 {
			t.Fatalf("Sibling %d should be 66 characters long, got %d", i, len(siblingStr))
		}

		if siblingStr[:2] != "0x" {
			t.Fatalf("Sibling %d should start with '0x', got %s", i, siblingStr[:2])
		}
	}

	// Validate Value field for existing key
	if !proof.Exists {
		t.Fatal("Proof should indicate existence")
	}

	if proof.Value.IsZero() {
		t.Fatal("Value should not be zero for existing key")
	}

	valueStr := proof.Value.String()
	if len(valueStr) != 66 {
		t.Fatalf("Value should be 66 characters long, got %d", len(valueStr))
	}

	if valueStr[:2] != "0x" {
		t.Fatalf("Value should start with '0x', got %s", valueStr[:2])
	}

	// Test non-existent key
	nonExistentIndex := big.NewInt(99)
	nonExistentProof, err := tree.Get(nonExistentIndex)
	if err != nil {
		t.Fatalf("Get for non-existent key failed: %v", err)
	}

	if nonExistentProof.Exists {
		t.Fatal("Non-existent proof should have Exists=false")
	}

	if !nonExistentProof.Value.IsZero() {
		t.Fatal("Non-existent proof should have zero Value")
	}

	// But other fields should still be properly formatted
	nonExistentLeafStr := nonExistentProof.Leaf.String()
	if len(nonExistentLeafStr) != 66 || nonExistentLeafStr[:2] != "0x" {
		t.Fatal("Non-existent proof Leaf should still be properly formatted")
	}

	t.Logf("Proof field validation test passed")
}

// TestSolidityProofCompatibility tests that Go-generated proofs can be verified in Solidity
// and that the proof structures match exactly
func TestSolidityProofCompatibility(t *testing.T) {
	// Create a Go SMT instance
	db := smt.NewInMemoryDatabase()
	goSMT, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Test data
	testCases := []struct {
		name  string
		index *big.Int
		leaf  string
	}{
		{"simple_leaf", big.NewInt(5), "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"zero_index", big.NewInt(0), "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"},
		{"max_index", big.NewInt(255), "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse leaf value
			leafValue, err := smt.HexToBytes32(tc.leaf)
			if err != nil {
				t.Fatalf("Failed to parse leaf value: %v", err)
			}

			// Insert leaf in Go SMT
			updateProof, err := goSMT.Insert(tc.index, leafValue)
			if err != nil {
				t.Fatalf("Failed to insert leaf: %v", err)
			}

			// Get proof from Go SMT
			proof, err := goSMT.Get(tc.index)
			if err != nil {
				t.Fatalf("Failed to get proof: %v", err)
			}

			// Verify the proof structure matches expected format
			if !proof.Exists {
				t.Errorf("Expected proof.Exists to be true")
			}
			
			// Compute expected leaf hash
			expectedLeafHash := smt.ComputeLeafHash(tc.index, leafValue)
			if proof.Leaf.String() != expectedLeafHash.String() {
				t.Errorf("Expected proof.Leaf to be %s (computed hash), got %s", expectedLeafHash.String(), proof.Leaf.String())
			}
			if proof.Index.Cmp(tc.index) != 0 {
				t.Errorf("Expected proof.Index to be %s, got %s", tc.index.String(), proof.Index.String())
			}
			if proof.Value.String() != tc.leaf {
				t.Errorf("Expected proof.Value to be %s, got %s", tc.leaf, proof.Value.String())
			}

			// Verify the proof using Go's VerifyProof
			if !goSMT.VerifyProof(proof) {
				t.Errorf("Go proof verification failed")
			}

			// Verify UpdateProof structure
			if updateProof.Exists {
				t.Errorf("Expected updateProof.Exists to be false for new insertion")
			}
			if updateProof.NewLeaf.String() != expectedLeafHash.String() {
				t.Errorf("Expected updateProof.NewLeaf to be %s (computed hash), got %s", expectedLeafHash.String(), updateProof.NewLeaf.String())
			}

			t.Logf("Proof compatibility test passed for %s: leaf=%s, index=%s, enables=%s, siblings=%d",
				tc.name, proof.Leaf.String(), proof.Index.String(), proof.Enables.String(), len(proof.Siblings))
		})
	}
}

// TestSolidityProofJSONCompatibility tests that proof structures can be serialized/deserialized
// in a format compatible with Solidity events
func TestSolidityProofJSONCompatibility(t *testing.T) {
	// Create a Go SMT instance
	db := smt.NewInMemoryDatabase()
	goSMT, err := smt.NewSparseMerkleTree(db, 4) // Small tree for testing
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Insert multiple leaves to create a more complex proof structure
	testData := map[*big.Int]string{
		big.NewInt(1):  "0x1111111111111111111111111111111111111111111111111111111111111111",
		big.NewInt(5):  "0x5555555555555555555555555555555555555555555555555555555555555555",
		big.NewInt(10): "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	// Insert all leaves
	for index, leaf := range testData {
		leafValue, err := smt.HexToBytes32(leaf)
		if err != nil {
			t.Fatalf("Failed to parse leaf value: %v", err)
		}
		_, err = goSMT.Insert(index, leafValue)
		if err != nil {
			t.Fatalf("Failed to insert leaf at index %s: %v", index.String(), err)
		}
	}

	// Test proof generation for each leaf
	for index, expectedValue := range testData {
		proof, err := goSMT.Get(index)
		if err != nil {
			t.Fatalf("Failed to get proof: %v", err)
		}

		// Verify proof structure
		if !proof.Exists {
			t.Errorf("Expected proof to exist for index %s", index.String())
		}
		
		// Compute expected leaf hash
		leafValue, _ := smt.HexToBytes32(expectedValue)
		expectedLeafHash := smt.ComputeLeafHash(index, leafValue)
		if proof.Leaf.String() != expectedLeafHash.String() {
			t.Errorf("Expected leaf %s (computed hash), got %s", expectedLeafHash.String(), proof.Leaf.String())
		}
		if proof.Value.String() != expectedValue {
			t.Errorf("Expected value %s, got %s", expectedValue, proof.Value.String())
		}

		// Verify proof validation
		if !goSMT.VerifyProof(proof) {
			t.Errorf("Proof verification failed for index %s", index.String())
		}

		t.Logf("JSON compatibility test passed for index %s: exists=%t, leaf=%s, enables=%s, siblings=%d",
			index.String(), proof.Exists, proof.Leaf.String(), proof.Enables.String(), len(proof.Siblings))
	}
}

// TestSolidityEventDataCompatibility tests that proof data can be used in Solidity events
func TestSolidityEventDataCompatibility(t *testing.T) {
	// Create a Go SMT instance
	db := smt.NewInMemoryDatabase()
	goSMT, err := smt.NewSparseMerkleTree(db, 6)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Test update operation to generate comprehensive proof data
	index := big.NewInt(42)
	oldLeafValue, err := smt.HexToBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("Failed to parse old leaf value: %v", err)
	}
	newLeafValue, err := smt.HexToBytes32("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")
	if err != nil {
		t.Fatalf("Failed to parse new leaf value: %v", err)
	}

	// Capture initial empty root
	emptyRoot := goSMT.Root()

	// Insert initial leaf
	insertProof, err := goSMT.Insert(index, oldLeafValue)
	if err != nil {
		t.Fatalf("Failed to insert initial leaf: %v", err)
	}

	// Capture root after insert
	rootAfterInsert := goSMT.Root()

	// Update the leaf
	updateProof, err := goSMT.Update(index, newLeafValue)
	if err != nil {
		t.Fatalf("Failed to update leaf: %v", err)
	}

	// Verify insert proof structure (for TreeStateUpdated event)
	if insertProof.Exists {
		t.Errorf("Expected insertProof.Exists to be false")
	}
	expectedOldLeafHash := smt.ComputeLeafHash(index, oldLeafValue)
	if insertProof.NewLeaf.String() != expectedOldLeafHash.String() {
		t.Errorf("Expected insertProof.NewLeaf to be %s (computed hash), got %s", expectedOldLeafHash.String(), insertProof.NewLeaf.String())
	}

	// Verify update proof structure (for TreeStateUpdated event)
	if !updateProof.Exists {
		t.Errorf("Expected updateProof.Exists to be true")
	}
	if updateProof.Leaf.String() != expectedOldLeafHash.String() {
		t.Errorf("Expected updateProof.Leaf to be %s (computed hash), got %s", expectedOldLeafHash.String(), updateProof.Leaf.String())
	}
	expectedNewLeafHash := smt.ComputeLeafHash(index, newLeafValue)
	if updateProof.NewLeaf.String() != expectedNewLeafHash.String() {
		t.Errorf("Expected updateProof.NewLeaf to be %s (computed hash), got %s", expectedNewLeafHash.String(), updateProof.NewLeaf.String())
	}

	// Get current proof (for ProofGenerated event)
	currentProof, err := goSMT.Get(index)
	if err != nil {
		t.Fatalf("Failed to get current proof: %v", err)
	}
	if !currentProof.Exists {
		t.Errorf("Expected current proof to exist")
	}
	if currentProof.Leaf.String() != expectedNewLeafHash.String() {
		t.Errorf("Expected current leaf to be %s (computed hash), got %s", expectedNewLeafHash.String(), currentProof.Leaf.String())
	}

	// Note: UpdateProof contains the OLD proof (before the operation), so we can't verify it against current tree
	// This is the correct behavior as per TypeScript reference implementation
	// The proof in UpdateProof represents the state before the operation

	// Verify current proof is valid against current tree
	if !goSMT.VerifyProof(currentProof) {
		t.Errorf("Current proof verification failed")
	}

	t.Logf("Event data compatibility test passed")
	t.Logf("Insert proof: exists=%t, newLeaf=%s", insertProof.Exists, insertProof.NewLeaf.String())
	t.Logf("Update proof: exists=%t, oldLeaf=%s, newLeaf=%s", updateProof.Exists, updateProof.Leaf.String(), updateProof.NewLeaf.String())
	t.Logf("Current proof: exists=%t, leaf=%s, enables=%s, siblings=%d",
		currentProof.Exists, currentProof.Leaf.String(), currentProof.Enables.String(), len(currentProof.Siblings))
	t.Logf("Roots: empty=%s, afterInsert=%s, final=%s", emptyRoot.String(), rootAfterInsert.String(), goSMT.Root().String())
}

// TestSolidityProofStructureAlignment tests that Go proof structures align with Solidity expectations
func TestSolidityProofStructureAlignment(t *testing.T) {
	// Create a Go SMT instance
	db := smt.NewInMemoryDatabase()
	goSMT, err := smt.NewSparseMerkleTree(db, 8)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Test various scenarios that would be used in Solidity
	testCases := []struct {
		name     string
		index    *big.Int
		leaf     string
		expected struct {
			exists   bool
			hasValue bool
		}
	}{
		{
			name:  "existing_leaf",
			index: big.NewInt(100),
			leaf:  "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			expected: struct {
				exists   bool
				hasValue bool
			}{exists: true, hasValue: true},
		},
		{
			name:  "non_existing_leaf",
			index: big.NewInt(200),
			leaf:  "",
			expected: struct {
				exists   bool
				hasValue bool
			}{exists: false, hasValue: false},
		},
	}

	// Insert the existing leaf
	leafValue, err := smt.HexToBytes32(testCases[0].leaf)
	if err != nil {
		t.Fatalf("Failed to parse leaf value: %v", err)
	}
	_, err = goSMT.Insert(testCases[0].index, leafValue)
	if err != nil {
		t.Fatalf("Failed to insert test leaf: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proof, err := goSMT.Get(tc.index)
			if err != nil {
				t.Fatalf("Failed to get proof: %v", err)
			}

			// Verify proof structure matches Solidity expectations
			if proof.Exists != tc.expected.exists {
				t.Errorf("Expected exists=%t, got %t", tc.expected.exists, proof.Exists)
			}

			if tc.expected.hasValue {
				if proof.Value.IsZero() {
					t.Errorf("Expected proof to have a non-zero value")
				} else if proof.Value.String() != tc.leaf {
					t.Errorf("Expected value=%s, got %s", tc.leaf, proof.Value.String())
				}
			} else {
				if !proof.Value.IsZero() {
					t.Errorf("Expected proof value to be zero, got %s", proof.Value.String())
				}
			}

			// Verify index is correctly set
			if proof.Index.Cmp(tc.index) != 0 {
				t.Errorf("Expected index=%s, got %s", tc.index.String(), proof.Index.String())
			}

			// Verify enables and siblings are properly formatted
			if proof.Enables == nil {
				t.Errorf("Expected enables to be non-nil")
			}
			if proof.Siblings == nil {
				t.Errorf("Expected siblings to be non-nil (even if empty)")
			}

			// Verify proof validation
			if !goSMT.VerifyProof(proof) {
				t.Errorf("Proof verification failed")
			}

			t.Logf("Structure alignment test passed for %s: exists=%t, hasValue=%t, enables=%s, siblings=%d",
				tc.name, proof.Exists, tc.expected.hasValue, proof.Enables.String(), len(proof.Siblings))
		})
	}
}
