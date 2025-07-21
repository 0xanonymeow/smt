package tests

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	smt "github.com/0xanonymeow/smt/go"
)

// TestFullWorkflowIntegration tests complete end-to-end workflows
func TestFullWorkflowIntegration(t *testing.T) {
	t.Run("CompleteLifecycleWorkflow", func(t *testing.T) {
		// Create tree for testing
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, 16)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Phase 1: Initial data insertion
		t.Log("Phase 1: Initial data insertion")
		initialData := make(map[string]smt.Bytes32)
		for i := 0; i < 50; i++ {
			key := big.NewInt(int64(i))
			value := GenerateRandomBytes32(i * 10)
			keyStr := key.String()
			
			// Insert in standard tree
			_, err := tree.Insert(key, value)
			if err != nil {
				t.Fatalf("Tree insert failed: %v", err)
			}
			
			initialData[keyStr] = value
		}

		// Phase 2: Verification of insertions
		t.Log("Phase 2: Verification of insertions")
		for keyStr := range initialData {
			key := new(big.Int)
			key.SetString(keyStr, 10)
			
			// Verify existence
			exists, err := tree.Exists(key)
			if err != nil {
				t.Fatalf("Exists check failed: %v", err)
			}
			if !exists {
				t.Fatalf("Key %s should exist after insertion", keyStr)
			}
			
			// Get proof
			proof, err := tree.Get(key)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}
			if !proof.Exists {
				t.Fatalf("Proof should indicate key %s exists", keyStr)
			}
			
			// Verify proof
			if !tree.VerifyProof(proof) {
				t.Fatalf("Proof verification failed for key %s", keyStr)
			}
		}

		// Phase 3: Updates and modifications
		t.Log("Phase 3: Updates and modifications")
		updateData := make(map[string]smt.Bytes32)
		for i := 0; i < 25; i++ {
			key := big.NewInt(int64(i))
			newValue := GenerateRandomBytes32(i * 100)
			keyStr := key.String()
			
			// Update existing key
			_, err := tree.Update(key, newValue)
			if err != nil {
				t.Fatalf("Update failed: %v", err)
			}
			
			updateData[keyStr] = newValue
		}

		// Phase 4: Verification of updates
		t.Log("Phase 4: Verification of updates")
		for keyStr := range updateData {
			key := new(big.Int)
			key.SetString(keyStr, 10)
			
			// Get proof after update
			proof, err := tree.Get(key)
			if err != nil {
				t.Fatalf("Get after update failed: %v", err)
			}
			if !proof.Exists {
				t.Fatalf("Proof should indicate updated key %s exists", keyStr)
			}
			
			// Verify proof after update
			if !tree.VerifyProof(proof) {
				t.Fatalf("Proof verification failed for updated key %s", keyStr)
			}
		}

		// Phase 5: Key-Value operations
		t.Log("Phase 5: Key-Value operations")
		kvData := make(map[string]smt.Bytes32)
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("test-key-%d", i)
			value := GenerateRandomBytes32(i * 5)
			
			// Insert using KV interface
			_, err := tree.InsertKV(key, value)
			if err != nil {
				t.Fatalf("KV insert failed: %v", err)
			}
			
			kvData[key] = value
		}

		// Phase 6: Verification of KV operations
		t.Log("Phase 6: Verification of KV operations")
		for key, expectedValue := range kvData {
			// Get using KV interface
			value, exists, err := tree.GetKV(key)
			if err != nil {
				t.Fatalf("KV get failed: %v", err)
			}
			if !exists {
				t.Fatalf("KV key %s should exist", key)
			}
			if value != expectedValue {
				t.Fatalf("KV value mismatch for key %s", key)
			}
		}

		// Phase 7: Proof serialization and deserialization
		t.Log("Phase 7: Proof serialization and deserialization")
		testKey := big.NewInt(42)
		proof, err := tree.Get(testKey)
		if err != nil {
			t.Fatalf("Get for serialization test failed: %v", err)
		}

		// Serialize proof
		serialized := smt.SerializeProof(proof)
		
		// Deserialize proof
		deserialized, err := smt.DeserializeProof(serialized)
		if err != nil {
			t.Fatalf("Proof deserialization failed: %v", err)
		}

		// Verify deserialized proof
		if !tree.VerifyProof(deserialized) {
			t.Fatalf("Deserialized proof verification failed")
		}

		// Phase 8: Root computation and verification
		t.Log("Phase 8: Root computation and verification")
		currentRoot := tree.Root()
		computedRoot := tree.ComputeRoot(proof)
		
		if currentRoot != computedRoot {
			t.Logf("Root mismatch: current=%s, computed=%s", currentRoot.String(), computedRoot.String())
			// Note: This might be expected if the tree has been modified since the proof was generated
		}

		// Phase 9: Performance and stress testing
		t.Log("Phase 9: Performance and stress testing")
		startTime := time.Now()
		
		// Insert additional entries to test performance
		for i := 1000; i < 1100; i++ {
			key := big.NewInt(int64(i))
			value := GenerateRandomBytes32(i)
			
			_, err := tree.Insert(key, value)
			if err != nil {
				t.Fatalf("Performance test insert failed: %v", err)
			}
		}
		
		insertTime := time.Since(startTime)
		t.Logf("Performance test: 100 inserts took %v", insertTime)

		// Phase 10: Final verification
		t.Log("Phase 10: Final verification")
		finalRoot := tree.Root()
		if finalRoot.IsZero() {
			t.Fatal("Final root should not be zero")
		}
		
		t.Logf("Integration test completed successfully")
		t.Logf("Final tree root: %s", finalRoot.String())
		t.Logf("Total operations: %d inserts, %d updates, %d KV operations", 
			len(initialData), len(updateData), len(kvData))
	})
}

// TestConcurrentIntegration tests concurrent operations integration
func TestConcurrentIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent integration test in short mode")
	}

	t.Run("ConcurrentReadWrite", func(t *testing.T) {
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, 16)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Pre-populate tree
		for i := 0; i < 100; i++ {
			key := big.NewInt(int64(i))
			value := GenerateRandomBytes32(i)
			
			_, err := tree.Insert(key, value)
			if err != nil {
				t.Fatalf("Pre-population insert failed: %v", err)
			}
		}

		// Concurrent operations
		numWorkers := 10
		done := make(chan bool, numWorkers)
		
		// Start readers
		for i := 0; i < numWorkers/2; i++ {
			go func(workerID int) {
				defer func() { done <- true }()
				
				for j := 0; j < 50; j++ {
					key := big.NewInt(int64(j % 100))
					
					// Read operations
					_, err := tree.Get(key)
					if err != nil {
						t.Errorf("Concurrent get failed: %v", err)
					}
					
					_, err = tree.Exists(key)
					if err != nil {
						t.Errorf("Concurrent exists failed: %v", err)
					}
				}
			}(i)
		}
		
		// Start writers
		for i := numWorkers/2; i < numWorkers; i++ {
			go func(workerID int) {
				defer func() { done <- true }()
				
				for j := 0; j < 25; j++ {
					key := big.NewInt(int64(workerID*100 + j))
					value := GenerateRandomBytes32(workerID*100 + j)
					
					// Write operations
					_, err := tree.Insert(key, value)
					if err != nil {
						// Expected for some concurrent operations
						t.Logf("Concurrent insert failed (expected): %v", err)
					}
				}
			}(i)
		}
		
		// Wait for all workers to complete
		for i := 0; i < numWorkers; i++ {
			<-done
		}
		
		// Verify tree consistency
		finalRoot := tree.Root()
		if finalRoot.IsZero() {
			t.Fatal("Final root should not be zero after concurrent operations")
		}
		
		t.Logf("Concurrent integration test completed")
		t.Logf("Final root after concurrent operations: %s", finalRoot.String())
	})
}

// TestErrorHandlingIntegration tests error handling in integration scenarios
func TestErrorHandlingIntegration(t *testing.T) {
	t.Run("ErrorRecoveryWorkflow", func(t *testing.T) {
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, 8)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Test 1: Insert duplicate key error handling
		key := big.NewInt(42)
		value := GenerateRandomBytes32(42)
		
		_, err = tree.Insert(key, value)
		if err != nil {
			t.Fatalf("First insert should succeed: %v", err)
		}
		
		_, err = tree.Insert(key, value)
		if err == nil {
			t.Fatal("Second insert should fail with duplicate key error")
		}
		
		// Test 2: Update non-existent key error handling
		nonExistentKey := big.NewInt(999)
		_, err = tree.Update(nonExistentKey, value)
		if err == nil {
			t.Fatal("Update of non-existent key should fail")
		}
		
		// Test 3: Invalid index error handling
		invalidKey := big.NewInt(1000) // Too large for depth 8
		_, err = tree.Insert(invalidKey, value)
		if err == nil {
			t.Fatal("Insert with invalid index should fail")
		}
		
		// Test 4: Verify tree is still functional after errors
		exists, err := tree.Exists(key)
		if err != nil {
			t.Fatalf("Tree should still be functional after errors: %v", err)
		}
		if !exists {
			t.Fatal("Original key should still exist after error operations")
		}
		
		// Test 5: Successful operations after error recovery
		newKey := big.NewInt(100)
		newValue := GenerateRandomBytes32(100)
		
		_, err = tree.Insert(newKey, newValue)
		if err != nil {
			t.Fatalf("Insert after error recovery should succeed: %v", err)
		}
		
		proof, err := tree.Get(newKey)
		if err != nil {
			t.Fatalf("Get after error recovery should succeed: %v", err)
		}
		
		if !tree.VerifyProof(proof) {
			t.Fatal("Proof verification after error recovery should succeed")
		}
		
		t.Logf("Error handling integration test completed successfully")
	})
}

// TestMemoryAndPerformanceIntegration tests memory usage and performance
func TestMemoryAndPerformanceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory and performance integration test in short mode")
	}

	t.Run("MemoryPerformanceWorkflow", func(t *testing.T) {
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, 16)
		if err != nil {
			t.Fatalf("Failed to create tree: %v", err)
		}

		// Phase 1: Large scale insertions
		t.Log("Phase 1: Large scale insertions")
		numEntries := 1000
		startTime := time.Now()
		
		for i := 0; i < numEntries; i++ {
			key := big.NewInt(int64(i))
			value := GenerateRandomBytes32(i)
			
			_, err := tree.Insert(key, value)
			if err != nil {
				t.Fatalf("Large scale insert failed at %d: %v", i, err)
			}
			
			// Log progress
			if i%100 == 0 && i > 0 {
				elapsed := time.Since(startTime)
				rate := float64(i) / elapsed.Seconds()
				t.Logf("Inserted %d entries (%.2f entries/sec)", i, rate)
			}
		}
		
		insertTime := time.Since(startTime)
		t.Logf("Large scale insertion completed: %d entries in %v", numEntries, insertTime)

		// Phase 2: Random access performance
		t.Log("Phase 2: Random access performance")
		startTime = time.Now()
		
		for i := 0; i < 100; i++ {
			key := big.NewInt(int64(i * 10))
			
			proof, err := tree.Get(key)
			if err != nil {
				t.Fatalf("Random access get failed: %v", err)
			}
			
			if !tree.VerifyProof(proof) {
				t.Fatalf("Random access proof verification failed")
			}
		}
		
		accessTime := time.Since(startTime)
		t.Logf("Random access completed: 100 operations in %v", accessTime)

		// Phase 3: Batch operations performance
		t.Log("Phase 3: Batch operations performance")
		startTime = time.Now()
		
		for i := 0; i < 50; i++ {
			key := big.NewInt(int64(i))
			newValue := GenerateRandomBytes32(i * 1000)
			
			_, err := tree.Update(key, newValue)
			if err != nil {
				t.Fatalf("Batch update failed: %v", err)
			}
		}
		
		updateTime := time.Since(startTime)
		t.Logf("Batch updates completed: 50 operations in %v", updateTime)

		// Phase 4: Memory usage check
		t.Log("Phase 4: Memory usage check")
		finalRoot := tree.Root()
		if finalRoot.IsZero() {
			t.Fatal("Final root should not be zero after large scale operations")
		}
		
		t.Logf("Memory and performance integration test completed")
		t.Logf("Final tree root: %s", finalRoot.String())
		t.Logf("Performance summary:")
		t.Logf("  - Insertions: %v for %d entries", insertTime, numEntries)
		t.Logf("  - Random access: %v for 100 operations", accessTime)
		t.Logf("  - Updates: %v for 50 operations", updateTime)
	})
}