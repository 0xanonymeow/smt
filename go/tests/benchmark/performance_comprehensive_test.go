package benchmark

import (
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	smt "github.com/0xanonymeow/smt/go"
)

// ComprehensivePerformanceTest provides detailed performance analysis
type ComprehensivePerformanceTest struct {
	standardTree   *smt.SparseMerkleTree
	standardTreeKV *smt.SparseMerkleTree
}

// BenchmarkComprehensivePerformance runs all performance tests
func BenchmarkComprehensivePerformance(b *testing.B) {
	suite := &ComprehensivePerformanceTest{}

	b.Run("MemoryUsage", func(b *testing.B) {
		suite.benchmarkMemoryUsage(b)
	})

	b.Run("BatchOperations", func(b *testing.B) {
		suite.benchmarkBatchOperations(b)
	})

	b.Run("ConcurrentOperations", func(b *testing.B) {
		suite.benchmarkConcurrentOperations(b)
	})

	b.Run("ScalabilityTest", func(b *testing.B) {
		suite.benchmarkScalability(b)
	})

	b.Run("HashFunctionPerformance", func(b *testing.B) {
		suite.benchmarkHashPerformance(b)
	})

	b.Run("ProofOperations", func(b *testing.B) {
		suite.benchmarkProofOperations(b)
	})
}

func (suite *ComprehensivePerformanceTest) benchmarkMemoryUsage(b *testing.B) {
	b.Run("MemoryUsage", func(b *testing.B) {
		keys := generateRandomKeys(100, 42)
		values := generateRandomValues(100, 42)

		b.ReportAllocs()
		var memStart runtime.MemStats
		runtime.ReadMemStats(&memStart)

		for i := 0; i < b.N; i++ {
			db := smt.NewInMemoryDatabase()
			tree, err := smt.NewSparseMerkleTree(db, 16)
			if err != nil {
				b.Fatalf("Failed to create tree: %v", err)
			}
			for j := 0; j < 10 && j < len(keys); j++ {
				tree.Insert(keys[j], values[j])
			}
		}

		var memEnd runtime.MemStats
		runtime.ReadMemStats(&memEnd)
		b.Logf("Memory - Allocs: %d, Bytes: %d",
			memEnd.Mallocs-memStart.Mallocs,
			memEnd.TotalAlloc-memStart.TotalAlloc)
	})
}

func (suite *ComprehensivePerformanceTest) benchmarkBatchOperations(b *testing.B) {
	batchSizes := []int{10, 50, 100, 500}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			keys := generateRandomKeys(batchSize, 42)
			values := generateRandomValues(batchSize, 42)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				db := smt.NewInMemoryDatabase()
				tree, err := smt.NewSparseMerkleTree(db, 16)
				if err != nil {
					b.Fatalf("Failed to create tree: %v", err)
				}

				// Batch insert operations
				for j := 0; j < batchSize; j++ {
					_, err := tree.Insert(keys[j], values[j])
					if err != nil {
						continue // Skip if key already exists
					}
				}
			}

			b.Logf("Batch size: %d, Ops/sec: %.2f",
				batchSize, float64(b.N*batchSize)/b.Elapsed().Seconds())
		})
	}
}

func (suite *ComprehensivePerformanceTest) benchmarkConcurrentOperations(b *testing.B) {
	goroutineCounts := []int{1, 2, 4, 8}

	for _, numGoroutines := range goroutineCounts {
		b.Run(fmt.Sprintf("Goroutines_%d", numGoroutines), func(b *testing.B) {
			opsPerGoroutine := 25

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				db := smt.NewInMemoryDatabase()
				tree, err := smt.NewSparseMerkleTree(db, 16)
				if err != nil {
					b.Fatalf("Failed to create tree: %v", err)
				}
				var wg sync.WaitGroup
				var mu sync.Mutex

				for g := 0; g < numGoroutines; g++ {
					wg.Add(1)
					go func(goroutineID int) {
						defer wg.Done()
						keys := generateRandomKeys(opsPerGoroutine, int64(42+goroutineID+i))
						values := generateRandomValues(opsPerGoroutine, int64(84+goroutineID+i))

						for j := 0; j < opsPerGoroutine; j++ {
							mu.Lock()
							tree.Insert(keys[j], values[j])
							mu.Unlock()
						}
					}(g)
				}

				wg.Wait()
			}

			totalOps := b.N * numGoroutines * opsPerGoroutine
			b.Logf("Goroutines: %d, Total ops: %d, Ops/sec: %.2f",
				numGoroutines, totalOps, float64(totalOps)/b.Elapsed().Seconds())
		})
	}
}

func (suite *ComprehensivePerformanceTest) benchmarkScalability(b *testing.B) {
	treeSizes := []int{100, 500, 1000, 5000}

	for _, treeSize := range treeSizes {
		b.Run(fmt.Sprintf("TreeSize_%d", treeSize), func(b *testing.B) {
			// Pre-populate tree
			db := smt.NewInMemoryDatabase()
			tree, err := smt.NewSparseMerkleTree(db, 20)
			if err != nil {
				b.Fatalf("Failed to create tree: %v", err)
			}
			keys := generateRandomKeys(treeSize, 42)
			values := generateRandomValues(treeSize, 42)

			for i := 0; i < treeSize; i++ {
				tree.Insert(keys[i], values[i])
			}

			// Test operations on populated tree
			testKeys := generateRandomKeys(100, 84)
			testValues := generateRandomValues(100, 84)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				for j := 0; j < 10; j++ {
					// Test get operation
					keyIdx := j % len(keys)
					proof, _ := tree.Get(keys[keyIdx])
					_ = proof

					// Test insert operation
					if j < len(testKeys) {
						tree.Insert(testKeys[j], testValues[j])
					}
				}
			}

			b.Logf("Tree size: %d, Operations completed", treeSize)
		})
	}
}

func (suite *ComprehensivePerformanceTest) benchmarkHashPerformance(b *testing.B) {
	// Test hashing performance with Bytes32 values
	values := generateRandomValues(3, 42)

	b.Run("HashBytes32", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for j := 0; j < len(values)-1; j++ {
				result := smt.HashBytes32(values[j], values[j+1])
				_ = result
			}
		}
	})

	b.Run("ZeroValueHash", func(b *testing.B) {
		zero1 := smt.Bytes32{}
		zero2 := smt.Bytes32{}
		
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Benchmark the current SMT library implementation
			result := smt.HashBytes32(zero1, zero2)
			_ = result
		}
	})
	
	b.Run("DirectKeccakZeroHash", func(b *testing.B) {
		zeros64 := make([]byte, 64) // 32 + 32 zero bytes
		
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Benchmark direct go-ethereum crypto.Keccak256
			result := crypto.Keccak256(zeros64)
			_ = result
		}
	})

	b.Run("ComputeLeafHash", func(b *testing.B) {
		indices := generateRandomKeys(100, 42)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Test leaf hash computation
			idx := i % len(indices)
			result := smt.ComputeLeafHash(indices[idx], values[idx%len(values)])
			_ = result
		}
	})
}

func (suite *ComprehensivePerformanceTest) benchmarkProofOperations(b *testing.B) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 16)
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}
	keys := generateRandomKeysForDepth(100, 42, 16)
	values := generateRandomValues(100, 42)

	// Pre-populate tree
	for i := 0; i < len(keys); i++ {
		tree.Insert(keys[i], values[i])
	}

	b.Run("ProofGeneration", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			keyIdx := i % len(keys)
			proof, _ := tree.Get(keys[keyIdx])
			_ = proof
		}
	})

	b.Run("ProofVerification", func(b *testing.B) {
		// Generate proofs first
		proofs := make([]*smt.Proof, len(keys))
		for i := 0; i < len(keys); i++ {
			proofs[i], _ = tree.Get(keys[i])
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			proofIdx := i % len(proofs)
			proof := proofs[proofIdx]
			valid := tree.VerifyProof(proof)
			if !valid {
				b.Fatalf("Proof verification failed for index %d", proofIdx)
			}
		}
	})

	b.Run("BatchProofGeneration", func(b *testing.B) {
		batchSize := 10
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			proofs := make([]*smt.Proof, batchSize)
			for j := 0; j < batchSize; j++ {
				keyIdx := (i*batchSize + j) % len(keys)
				proofs[j], _ = tree.Get(keys[keyIdx])
			}
			_ = proofs
		}
	})
}

// TestMemoryUsageProfile analyzes memory usage patterns
func TestMemoryUsageProfile(t *testing.T) {
	// Test standard operations
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 16)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}
	keys := generateRandomKeysForDepth(100, 42, 16)
	values := generateRandomValues(100, 42)

	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	insertAllocations := int64(0)
	for i := 0; i < 100; i++ {
		_, err := tree.Insert(keys[i], values[i])
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
		// Approximate allocation tracking
		insertAllocations++
	}

	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	t.Logf("Memory usage - Total allocations: %d", insertAllocations)
	t.Logf("Runtime stats - Allocs: %d, Bytes: %d",
		memEnd.Mallocs-memStart.Mallocs,
		memEnd.TotalAlloc-memStart.TotalAlloc)
}

// TestConcurrentSafety validates thread safety of operations
func TestConcurrentSafety(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 16)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}
	numGoroutines := 10
	opsPerGoroutine := 50

	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make(chan error, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			keys := generateRandomKeysForDepth(opsPerGoroutine, int64(42+goroutineID), 16)
			values := generateRandomValues(opsPerGoroutine, int64(84+goroutineID))

			for i := 0; i < opsPerGoroutine; i++ {
				mu.Lock()
				_, err := tree.Insert(keys[i], values[i])
				mu.Unlock()

				if err != nil {
					// Skip key already exists errors in concurrent test
					if _, ok := err.(*smt.KeyExistsError); !ok {
						errors <- fmt.Errorf("goroutine %d, op %d: %v", goroutineID, i, err)
						return
					}
					// Key already exists is acceptable in concurrent operations
					continue
				}

				// Verify the insertion (only for successful inserts)
				mu.Lock()
				proof, _ := tree.Get(keys[i])
				mu.Unlock()

				if !proof.Exists {
					errors <- fmt.Errorf("goroutine %d, op %d: proof shows non-existence", goroutineID, i)
					return
				}
			}
		}(g)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}

	t.Logf("Concurrent safety test completed with %d goroutines, %d ops each",
		numGoroutines, opsPerGoroutine)
}

// BenchmarkProductionWorkload simulates realistic production usage patterns
func BenchmarkProductionWorkload(b *testing.B) {
	// Simulate a realistic workload: 70% reads, 20% inserts, 10% updates
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 20)
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}

	// Pre-populate with some data
	initialKeys := generateRandomKeys(1000, 42)
	initialValues := generateRandomValues(1000, 42)

	for i := 0; i < len(initialKeys); i++ {
		tree.Insert(initialKeys[i], initialValues[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		operation := i % 10

		switch {
		case operation < 7: // 70% reads
			keyIdx := i % len(initialKeys)
			proof, _ := tree.Get(initialKeys[keyIdx])
			_ = proof

		case operation < 9: // 20% inserts
			newKey := big.NewInt(int64(1000000 + i))
			newValue := smt.Bytes32{}
			for j := range newValue {
				newValue[j] = byte((2000000 + i + j) % 256)
			}
			tree.Insert(newKey, newValue)

		default: // 10% updates
			keyIdx := i % len(initialKeys)
			newValue := smt.Bytes32{}
			for j := range newValue {
				newValue[j] = byte((3000000 + i + j) % 256)
			}
			tree.Update(initialKeys[keyIdx], newValue)
		}
	}

	b.Logf("Production workload completed: %d operations", b.N)
}

// TestBatchOperations tests batch operations performance
func TestBatchOperations(t *testing.T) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 16)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	keys := generateRandomKeys(50, 42)
	values := generateRandomValues(50, 42)

	start := time.Now()

	// Batch insert operations
	for i := 0; i < len(keys); i++ {
		_, err := tree.Insert(keys[i], values[i])
		if err != nil {
			continue // Skip if key already exists
		}
	}

	duration := time.Since(start)

	// Verify all insertions
	insertedCount := 0
	for i := 0; i < len(keys); i++ {
		proof, err := tree.Get(keys[i])
		if err == nil && proof != nil && proof.Exists {
			insertedCount++
		}
	}

	t.Logf("Processed %d operations in %v, %d keys inserted", len(keys), duration, insertedCount)
}

// Helper functions are now in helpers_test.go
