package tests

import (
	"math/big"
	"testing"

	smt "github.com/0xanonymeow/smt/go"
)

// TestDebugger provides debugging utilities for test failures
type TestDebugger struct {
	tree       *smt.SparseMerkleTree
	operations []Operation
}

// Operation represents a tree operation for debugging
type Operation struct {
	Type   string
	Key    *big.Int
	Value  smt.Bytes32
	Error  error
	Result interface{}
}

// NewTestDebugger creates a new test debugger
func NewTestDebugger() *TestDebugger {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 16)
	if err != nil {
		panic(err)
	}
	return &TestDebugger{
		tree:       tree,
		operations: make([]Operation, 0),
	}
}

// RecordOperation records an operation for debugging
func (d *TestDebugger) RecordOperation(opType string, key *big.Int, value smt.Bytes32, err error, result interface{}) {
	d.operations = append(d.operations, Operation{
		Type:   opType,
		Key:    key,
		Value:  value,
		Error:  err,
		Result: result,
	})
}

// DumpOperations prints all recorded operations
func (d *TestDebugger) DumpOperations(t *testing.T) {
	t.Logf("=== Operation History ===")
	for i, op := range d.operations {
		t.Logf("[%d] %s: key=%s value=%s err=%v", i, op.Type, op.Key.String(), op.Value.String(), op.Error)
	}
	t.Logf("=== End Operation History ===")
}

// CreateTestTree creates a test tree with given depth
func CreateTestTree(t *testing.T, depth uint16) *smt.SparseMerkleTree {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, depth)
	if err != nil {
		t.Fatalf("Failed to create test tree: %v", err)
	}
	return tree
}

// InsertTestData inserts test data into a tree
func InsertTestData(t *testing.T, tree *smt.SparseMerkleTree, count int) {
	for i := 0; i < count; i++ {
		index := big.NewInt(int64(i))
		value := smt.Bytes32{}
		// Fill value with pattern
		for j := 0; j < 32; j++ {
			value[j] = byte(i % 256)
		}
		
		_, err := tree.Insert(index, value)
		if err != nil {
			t.Fatalf("Failed to insert test data at index %d: %v", i, err)
		}
	}
}

// VerifyTreeConsistency verifies tree consistency
func VerifyTreeConsistency(t *testing.T, tree *smt.SparseMerkleTree, expectedKeys []*big.Int) {
	for _, key := range expectedKeys {
		exists, err := tree.Exists(key)
		if err != nil {
			t.Fatalf("Exists check failed for key %s: %v", key.String(), err)
		}
		if !exists {
			t.Fatalf("Expected key %s to exist", key.String())
		}
		
		proof, err := tree.Get(key)
		if err != nil {
			t.Fatalf("Get failed for key %s: %v", key.String(), err)
		}
		if !proof.Exists {
			t.Fatalf("Proof indicates key %s doesn't exist", key.String())
		}
		
		if !tree.VerifyProof(proof) {
			t.Fatalf("Proof verification failed for key %s", key.String())
		}
	}
}

// CompareTreeRoots compares roots of two trees
func CompareTreeRoots(t *testing.T, tree1, tree2 *smt.SparseMerkleTree) {
	root1 := tree1.Root()
	root2 := tree2.Root()
	
	if root1 != root2 {
		t.Fatalf("Tree roots don't match: %s vs %s", root1.String(), root2.String())
	}
}

// GenerateRandomBytes32 generates a random Bytes32 value for testing
func GenerateRandomBytes32(seed int) smt.Bytes32 {
	var b smt.Bytes32
	for i := 0; i < 32; i++ {
		b[i] = byte((seed * (i + 1)) % 256)
	}
	return b
}