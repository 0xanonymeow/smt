package tests

import (
	"fmt"
	"testing"

	"github.com/0xanonymeow/smt/go/internal/simulator"
	"github.com/0xanonymeow/smt/go/internal/testutils"
	"github.com/0xanonymeow/smt/go/internal/vectors"
)

// TestSolidityRootSimulator tests the basic functionality of the Solidity root simulator
func TestSolidityRootSimulator(t *testing.T) {
	sim := simulator.NewSolidityRootSimulator()

	tests := []struct {
		name      string
		treeDepth uint16
		leaf      string
		index     string
		enables   string
		siblings  []string
		expected  string
		shouldErr bool
	}{
		{
			name:      "simple single level tree",
			treeDepth: 1,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x0",
			enables:   "0x1", // Only first level enabled
			siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"},
			expected:  "", // Will be computed
			shouldErr: false,
		},
		{
			name:      "zero leaf with zero sibling",
			treeDepth: 1,
			leaf:      "0x0000000000000000000000000000000000000000000000000000000000000000",
			index:     "0x0",
			enables:   "0x1",
			siblings:  []string{"0x0000000000000000000000000000000000000000000000000000000000000000"},
			expected:  "0x0000000000000000000000000000000000000000000000000000000000000000", // Should return zero
			shouldErr: false,
		},
		{
			name:      "multi-level tree",
			treeDepth: 3,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x5", // Binary: 101, so path is right-left-right
			enables:   "0x7", // Binary: 111, all three levels enabled
			siblings: []string{
				"0x2222222222222222222222222222222222222222222222222222222222222222", // Level 0
				"0x3333333333333333333333333333333333333333333333333333333333333333", // Level 1
				"0x4444444444444444444444444444444444444444444444444444444444444444", // Level 2
			},
			expected:  "", // Will be computed
			shouldErr: false,
		},
		{
			name:      "no levels enabled",
			treeDepth: 3,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x0",
			enables:   "0x0", // No levels enabled
			siblings:  []string{},
			expected:  "0x1111111111111111111111111111111111111111111111111111111111111111", // Should return original leaf
			shouldErr: false,
		},
		{
			name:      "partial levels enabled",
			treeDepth: 4,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x5", // Binary: 0101
			enables:   "0x5", // Binary: 0101, levels 0 and 2 enabled
			siblings: []string{
				"0x2222222222222222222222222222222222222222222222222222222222222222", // Level 0
				"0x3333333333333333333333333333333333333333333333333333333333333333", // Level 1 (not used)
				"0x4444444444444444444444444444444444444444444444444444444444444444", // Level 2
				"0x5555555555555555555555555555555555555555555555555555555555555555", // Level 3 (not used)
			},
			expected:  "", // Will be computed
			shouldErr: false,
		},
		{
			name:      "invalid tree depth",
			treeDepth: 257,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x0",
			enables:   "0x1",
			siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"},
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "index out of range",
			treeDepth: 2,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x4", // 4 is out of range for depth 2 (max index is 3)
			enables:   "0x1",
			siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"},
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "insufficient siblings",
			treeDepth: 3,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x0",
			enables:   "0x7", // All 3 levels enabled (binary: 111)
			siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"}, // Only 1 sibling, but need 3
			expected:  "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sim.ComputeRoot(tt.treeDepth, tt.leaf, tt.index, tt.enables, tt.siblings)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				t.Logf("Expected error occurred: %v", err)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expected != "" {
				if !testutils.CompareHexStrings(result, tt.expected) {
					t.Fatalf("Expected %s, got %s", tt.expected, result)
				}
			}

			t.Logf("Test %s: result=%s", tt.name, result)
		})
	}
}

// TestSolidityHashFunction tests the hash function behavior specifically
func TestSolidityHashFunction(t *testing.T) {
	sim := simulator.NewSolidityRootSimulator()

	tests := []struct {
		name     string
		left     string
		right    string
		expected string
	}{
		{
			name:     "both zero should return zero",
			left:     "0x0000000000000000000000000000000000000000000000000000000000000000",
			right:    "0x0000000000000000000000000000000000000000000000000000000000000000",
			expected: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "left zero, right non-zero",
			left:     "0x0000000000000000000000000000000000000000000000000000000000000000",
			right:    "0x1111111111111111111111111111111111111111111111111111111111111111",
			expected: "", // Will be computed using keccak256
		},
		{
			name:     "both non-zero",
			left:     "0x1111111111111111111111111111111111111111111111111111111111111111",
			right:    "0x2222222222222222222222222222222222222222222222222222222222222222",
			expected: "", // Will be computed using keccak256
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to test the hash function indirectly through ComputeRoot
			// since solidityHash is not exported. We'll use a simple single-level tree.
			result, err := sim.ComputeRoot(1, tt.left, "0x0", "0x1", []string{tt.right})
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expected != "" {
				if !testutils.CompareHexStrings(result, tt.expected) {
					t.Fatalf("Expected %s, got %s", tt.expected, result)
				}
			}

			t.Logf("Hash test %s: result=%s", tt.name, result)
		})
	}
}

// TestSolidityRootSimulatorWithVectors tests the simulator using test vectors
func TestSolidityRootSimulatorWithVectors(t *testing.T) {
	sim := simulator.NewSolidityRootSimulator()

	// Load test vectors from JSON file
	testVectors, err := vectors.LoadRootComputationVectors("testdata/root_computation_vectors.json")
	if err != nil {
		t.Logf("Could not load root computation vectors, creating default test vectors: %v", err)
		// Create and save default test vectors if file doesn't exist
		testVectors = createDefaultRootComputationVectors()
		if saveErr := vectors.SaveRootComputationVectors("testdata/root_computation_vectors.json", testVectors); saveErr != nil {
			t.Logf("Could not save default vectors: %v", saveErr)
		}
	}

	for i, vector := range testVectors {
		t.Run(fmt.Sprintf("vector_%d", i), func(t *testing.T) {
			result, err := sim.ComputeRoot(vector.TreeDepth, vector.Leaf, vector.Index, vector.Enables, vector.Siblings)
			if err != nil {
				t.Fatalf("Unexpected error for vector %d: %v", i, err)
			}

			if vector.Expected != "" {
				if !testutils.CompareHexStrings(result, vector.Expected) {
					t.Fatalf("Vector %d: Expected %s, got %s", i, vector.Expected, result)
				}
			}

			t.Logf("Vector %d: treeDepth=%d, leaf=%s, index=%s, enables=%s, result=%s",
				i, vector.TreeDepth, vector.Leaf, vector.Index, vector.Enables, result)
		})
	}
}

// TestInputValidation tests the input validation functionality
func TestInputValidation(t *testing.T) {
	sim := simulator.NewSolidityRootSimulator()

	tests := []struct {
		name      string
		treeDepth uint16
		leaf      string
		index     string
		enables   string
		siblings  []string
		shouldErr bool
	}{
		{
			name:      "valid inputs",
			treeDepth: 8,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0xff",
			enables:   "0xff",
			siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"},
			shouldErr: false,
		},
		{
			name:      "invalid tree depth",
			treeDepth: 257,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x0",
			enables:   "0x1",
			siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"},
			shouldErr: true,
		},
		{
			name:      "invalid leaf hex",
			treeDepth: 8,
			leaf:      "0xzzzz",
			index:     "0x0",
			enables:   "0x1",
			siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"},
			shouldErr: true,
		},
		{
			name:      "invalid index hex",
			treeDepth: 8,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0xzzzz",
			enables:   "0x1",
			siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"},
			shouldErr: true,
		},
		{
			name:      "invalid enables hex",
			treeDepth: 8,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x0",
			enables:   "0xzzzz",
			siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"},
			shouldErr: true,
		},
		{
			name:      "invalid sibling hex",
			treeDepth: 8,
			leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			index:     "0x0",
			enables:   "0x1",
			siblings:  []string{"0xzzzz"},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sim.ValidateInputs(tt.treeDepth, tt.leaf, tt.index, tt.enables, tt.siblings)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected validation error but got none")
				}
				t.Logf("Expected validation error occurred: %v", err)
			} else {
				if err != nil {
					t.Fatalf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

// createDefaultRootComputationVectors creates a set of default test vectors for root computation testing
func createDefaultRootComputationVectors() []vectors.RootComputationTestVector {
	return []vectors.RootComputationTestVector{
		{
			TreeDepth: 1,
			Leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			Index:     "0x0",
			Enables:   "0x1",
			Siblings:  []string{"0x2222222222222222222222222222222222222222222222222222222222222222"},
			Expected:  "", // Will be computed
		},
		{
			TreeDepth: 1,
			Leaf:      "0x0000000000000000000000000000000000000000000000000000000000000000",
			Index:     "0x0",
			Enables:   "0x1",
			Siblings:  []string{"0x0000000000000000000000000000000000000000000000000000000000000000"},
			Expected:  "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			TreeDepth: 3,
			Leaf:      "0x1111111111111111111111111111111111111111111111111111111111111111",
			Index:     "0x5", // Binary: 101
			Enables:   "0x7", // Binary: 111
			Siblings: []string{
				"0x2222222222222222222222222222222222222222222222222222222222222222",
				"0x3333333333333333333333333333333333333333333333333333333333333333",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
			},
			Expected: "", // Will be computed
		},
		{
			TreeDepth: 2,
			Leaf:      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Index:     "0x2", // Binary: 10
			Enables:   "0x3", // Binary: 11
			Siblings: []string{
				"0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
				"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			Expected: "", // Will be computed
		},
	}
}