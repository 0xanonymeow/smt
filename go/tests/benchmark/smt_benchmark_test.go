package benchmark

import (
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"testing"
	"time"

	smt "github.com/0xanonymeow/smt/go"
)

// BenchmarkConfig holds configuration for benchmark tests
type BenchmarkConfig struct {
	TreeDepth     uint16
	NumOperations int
	BatchSize     int
	KeySize       int
}

// Standard benchmark configurations
var (
	SmallConfig = BenchmarkConfig{
		TreeDepth:     8,
		NumOperations: 100,
		BatchSize:     10,
		KeySize:       32,
	}
	MediumConfig = BenchmarkConfig{
		TreeDepth:     16,
		NumOperations: 1000,
		BatchSize:     50,
		KeySize:       32,
	}
	LargeConfig = BenchmarkConfig{
		TreeDepth:     20,
		NumOperations: 10000,
		BatchSize:     100,
		KeySize:       32,
	}
	ProductionConfig = BenchmarkConfig{
		TreeDepth:     256,
		NumOperations: 1000,
		BatchSize:     50,
		KeySize:       32,
	}
)

// Memory tracking utilities
type MemStats struct {
	AllocsBefore  uint64
	AllocsAfter   uint64
	BytesBefore   uint64
	BytesAfter    uint64
	SysBefore     uint64
	SysAfter      uint64
	GCCountBefore uint32
	GCCountAfter  uint32
}

func captureMemStats() MemStats {
	runtime.GC() // Force GC to get accurate measurements
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemStats{
		AllocsBefore:  m.TotalAlloc,
		BytesBefore:   m.Alloc,
		SysBefore:     m.Sys,
		GCCountBefore: m.NumGC,
	}
}

func (ms *MemStats) capture() {
	runtime.GC() // Force GC to get accurate measurements
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ms.AllocsAfter = m.TotalAlloc
	ms.BytesAfter = m.Alloc
	ms.SysAfter = m.Sys
	ms.GCCountAfter = m.NumGC
}

func (ms MemStats) String() string {
	return fmt.Sprintf("Allocs: %d->%d (+%d), Bytes: %d->%d (+%d), Sys: %d->%d (+%d), GC: %d->%d (+%d)",
		ms.AllocsBefore, ms.AllocsAfter, ms.AllocsAfter-ms.AllocsBefore,
		ms.BytesBefore, ms.BytesAfter, ms.BytesAfter-ms.BytesBefore,
		ms.SysBefore, ms.SysAfter, ms.SysAfter-ms.SysBefore,
		ms.GCCountBefore, ms.GCCountAfter, ms.GCCountAfter-ms.GCCountBefore)
}

// Test data generation utilities - using functions from comprehensive_performance_test.go

// SparseMerkleTree Benchmarks

func BenchmarkSparseMerkleTree_Insert_Small(b *testing.B) {
	benchmarkSMTInsert(b, SmallConfig)
}

func BenchmarkSparseMerkleTree_Insert_Medium(b *testing.B) {
	benchmarkSMTInsert(b, MediumConfig)
}

func BenchmarkSparseMerkleTree_Insert_Large(b *testing.B) {
	benchmarkSMTInsert(b, LargeConfig)
}

func BenchmarkSparseMerkleTree_Insert_Production(b *testing.B) {
	benchmarkSMTInsert(b, ProductionConfig)
}

func benchmarkSMTInsert(b *testing.B, config BenchmarkConfig) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, config.TreeDepth)
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}
	
	keys := generateRandomKeys(config.NumOperations, 42)
	values := generateRandomValues(config.NumOperations, 42)

	b.ResetTimer()
	b.ReportAllocs()

	memStart := captureMemStats()

	for i := 0; i < b.N; i++ {
		for j := 0; j < config.NumOperations && j < len(keys); j++ {
			// Use modulo to cycle through keys if b.N is large
			keyIdx := (i*config.NumOperations + j) % len(keys)
			valueIdx := (i*config.NumOperations + j) % len(values)

			// Create fresh tree for each benchmark iteration to avoid key conflicts
			if j == 0 {
				db = smt.NewInMemoryDatabase()
				tree, _ = smt.NewSparseMerkleTree(db, config.TreeDepth)
			}

			_, err := tree.Insert(keys[keyIdx], values[valueIdx])
			if err != nil {
				// Skip if key already exists (expected in repeated runs)
				continue
			}
		}
	}

	memStart.capture()
	b.StopTimer()

	b.Logf("Memory usage: %s", memStart.String())
	b.Logf("Operations per iteration: %d", config.NumOperations)
	b.Logf("Tree depth: %d", config.TreeDepth)
}

func BenchmarkSparseMerkleTree_Get_Small(b *testing.B) {
	benchmarkSMTGet(b, SmallConfig)
}

func BenchmarkSparseMerkleTree_Get_Medium(b *testing.B) {
	benchmarkSMTGet(b, MediumConfig)
}

func BenchmarkSparseMerkleTree_Get_Large(b *testing.B) {
	benchmarkSMTGet(b, LargeConfig)
}

func BenchmarkSparseMerkleTree_Get_Production(b *testing.B) {
	benchmarkSMTGet(b, ProductionConfig)
}

func benchmarkSMTGet(b *testing.B, config BenchmarkConfig) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, config.TreeDepth)
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}
	
	keys := generateRandomKeys(config.NumOperations, 42)
	values := generateRandomValues(config.NumOperations, 42)

	// Pre-populate tree
	for i := 0; i < config.NumOperations && i < len(keys); i++ {
		tree.Insert(keys[i], values[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	memStart := captureMemStats()

	for i := 0; i < b.N; i++ {
		for j := 0; j < config.NumOperations; j++ {
			keyIdx := j % len(keys)
			proof, err := tree.Get(keys[keyIdx])
			if err != nil {
				b.Fatalf("Get failed: %v", err)
			}
			_ = proof
		}
	}

	memStart.capture()
	b.StopTimer()

	b.Logf("Memory usage: %s", memStart.String())
	b.Logf("Operations per iteration: %d", config.NumOperations)
	b.Logf("Tree depth: %d", config.TreeDepth)
}

func BenchmarkSparseMerkleTree_Update_Small(b *testing.B) {
	benchmarkSMTUpdate(b, SmallConfig)
}

func BenchmarkSparseMerkleTree_Update_Medium(b *testing.B) {
	benchmarkSMTUpdate(b, MediumConfig)
}

func BenchmarkSparseMerkleTree_Update_Large(b *testing.B) {
	benchmarkSMTUpdate(b, LargeConfig)
}

func BenchmarkSparseMerkleTree_Update_Production(b *testing.B) {
	benchmarkSMTUpdate(b, ProductionConfig)
}

func benchmarkSMTUpdate(b *testing.B, config BenchmarkConfig) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, config.TreeDepth)
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}
	
	keys := generateRandomKeys(config.NumOperations, 42)
	values := generateRandomValues(config.NumOperations, 42)
	newValues := generateRandomValues(config.NumOperations, 84)

	// Pre-populate tree
	for i := 0; i < config.NumOperations && i < len(keys); i++ {
		tree.Insert(keys[i], values[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	memStart := captureMemStats()

	for i := 0; i < b.N; i++ {
		for j := 0; j < config.NumOperations; j++ {
			keyIdx := j % len(keys)
			valueIdx := j % len(newValues)
			_, err := tree.Update(keys[keyIdx], newValues[valueIdx])
			if err != nil {
				b.Fatalf("Update failed: %v", err)
			}
		}
	}

	memStart.capture()
	b.StopTimer()

	b.Logf("Memory usage: %s", memStart.String())
	b.Logf("Operations per iteration: %d", config.NumOperations)
	b.Logf("Tree depth: %d", config.TreeDepth)
}

func BenchmarkSparseMerkleTree_Verify_Small(b *testing.B) {
	benchmarkSMTVerify(b, SmallConfig)
}

func BenchmarkSparseMerkleTree_Verify_Medium(b *testing.B) {
	benchmarkSMTVerify(b, MediumConfig)
}

func BenchmarkSparseMerkleTree_Verify_Large(b *testing.B) {
	benchmarkSMTVerify(b, LargeConfig)
}

func BenchmarkSparseMerkleTree_Verify_Production(b *testing.B) {
	benchmarkSMTVerify(b, ProductionConfig)
}

func benchmarkSMTVerify(b *testing.B, config BenchmarkConfig) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, config.TreeDepth)
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}
	
	keys := generateRandomKeys(config.NumOperations, 42)
	values := generateRandomValues(config.NumOperations, 42)
	proofs := make([]*smt.Proof, config.NumOperations)

	// Pre-populate tree and collect proofs
	for i := 0; i < config.NumOperations && i < len(keys); i++ {
		tree.Insert(keys[i], values[i])
		proof, err := tree.Get(keys[i])
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
		proofs[i] = proof
	}

	b.ResetTimer()
	b.ReportAllocs()

	memStart := captureMemStats()

	for i := 0; i < b.N; i++ {
		for j := 0; j < config.NumOperations; j++ {
			proofIdx := j % len(proofs)
			proof := proofs[proofIdx]
			valid := tree.VerifyProof(proof)
			if !valid {
				b.Fatal("Proof should be valid")
			}
		}
	}

	memStart.capture()
	b.StopTimer()

	b.Logf("Memory usage: %s", memStart.String())
	b.Logf("Operations per iteration: %d", config.NumOperations)
	b.Logf("Tree depth: %d", config.TreeDepth)
}

// Key-Value SparseMerkleTree Benchmarks

func BenchmarkSparseMerkleTreeKV_Insert_Small(b *testing.B) {
	benchmarkSMTKVInsert(b, SmallConfig)
}

func BenchmarkSparseMerkleTreeKV_Insert_Medium(b *testing.B) {
	benchmarkSMTKVInsert(b, MediumConfig)
}

func BenchmarkSparseMerkleTreeKV_Insert_Large(b *testing.B) {
	benchmarkSMTKVInsert(b, LargeConfig)
}

func BenchmarkSparseMerkleTreeKV_Insert_Production(b *testing.B) {
	benchmarkSMTKVInsert(b, ProductionConfig)
}

func benchmarkSMTKVInsert(b *testing.B, config BenchmarkConfig) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256) // Full depth for KV
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}
	
	// Generate string keys
	keys := make([]string, config.NumOperations)
	for i := 0; i < config.NumOperations; i++ {
		keys[i] = fmt.Sprintf("key-%d-%d", i, time.Now().UnixNano())
	}
	values := generateRandomValues(config.NumOperations, 42)

	b.ResetTimer()
	b.ReportAllocs()

	memStart := captureMemStats()

	for i := 0; i < b.N; i++ {
		for j := 0; j < config.NumOperations && j < len(keys); j++ {
			// Use modulo to cycle through keys if b.N is large
			keyIdx := (i*config.NumOperations + j) % len(keys)
			valueIdx := (i*config.NumOperations + j) % len(values)

			// Create fresh tree for each benchmark iteration to avoid key conflicts
			if j == 0 {
				db = smt.NewInMemoryDatabase()
				tree, _ = smt.NewSparseMerkleTree(db, 256)
			}

			_, err := tree.InsertKV(keys[keyIdx], values[valueIdx])
			if err != nil {
				// Skip if key already exists (expected in repeated runs)
				continue
			}
		}
	}

	memStart.capture()
	b.StopTimer()

	b.Logf("Memory usage: %s", memStart.String())
	b.Logf("Operations per iteration: %d", config.NumOperations)
	b.Logf("Tree depth: 256 (KV mode)")
}

func BenchmarkSparseMerkleTreeKV_Get_Small(b *testing.B) {
	benchmarkSMTKVGet(b, SmallConfig)
}

func BenchmarkSparseMerkleTreeKV_Get_Medium(b *testing.B) {
	benchmarkSMTKVGet(b, MediumConfig)
}

func BenchmarkSparseMerkleTreeKV_Get_Large(b *testing.B) {
	benchmarkSMTKVGet(b, LargeConfig)
}

func BenchmarkSparseMerkleTreeKV_Get_Production(b *testing.B) {
	benchmarkSMTKVGet(b, ProductionConfig)
}

func benchmarkSMTKVGet(b *testing.B, config BenchmarkConfig) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256) // Full depth for KV
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}
	
	// Generate string keys
	keys := make([]string, config.NumOperations)
	for i := 0; i < config.NumOperations; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
	}
	values := generateRandomValues(config.NumOperations, 42)

	// Pre-populate tree
	for i := 0; i < config.NumOperations && i < len(keys); i++ {
		tree.InsertKV(keys[i], values[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	memStart := captureMemStats()

	for i := 0; i < b.N; i++ {
		for j := 0; j < config.NumOperations; j++ {
			keyIdx := j % len(keys)
			value, exists, err := tree.GetKV(keys[keyIdx])
			if err != nil {
				b.Fatalf("GetKV failed: %v", err)
			}
			if !exists {
				b.Fatalf("Key should exist: %s", keys[keyIdx])
			}
			_ = value
		}
	}

	memStart.capture()
	b.StopTimer()

	b.Logf("Memory usage: %s", memStart.String())
	b.Logf("Operations per iteration: %d", config.NumOperations)
	b.Logf("Tree depth: 256 (KV mode)")
}

// Serialization Benchmarks

func BenchmarkSerialization_Bytes32ToString(b *testing.B) {
	values := generateRandomValues(1000, 42)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		idx := i % len(values)
		result := values[idx].String()
		_ = result
	}
}

func BenchmarkSerialization_HexToBytes32(b *testing.B) {
	// Generate hex strings
	hexStrings := make([]string, 1000)
	values := generateRandomValues(1000, 42)
	for i := range hexStrings {
		hexStrings[i] = values[i].String()
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		idx := i % len(hexStrings)
		result, err := smt.HexToBytes32(hexStrings[idx])
		if err != nil {
			b.Fatalf("HexToBytes32 failed: %v", err)
		}
		_ = result
	}
}

// Mixed Operations Benchmarks

func BenchmarkMixedOperations_Small(b *testing.B) {
	benchmarkMixedOperations(b, SmallConfig)
}

func BenchmarkMixedOperations_Medium(b *testing.B) {
	benchmarkMixedOperations(b, MediumConfig)
}

func BenchmarkMixedOperations_Large(b *testing.B) {
	benchmarkMixedOperations(b, LargeConfig)
}

func benchmarkMixedOperations(b *testing.B, config BenchmarkConfig) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, config.TreeDepth)
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}
	
	keys := generateRandomKeys(config.NumOperations, 42)
	values := generateRandomValues(config.NumOperations, 42)

	b.ResetTimer()
	b.ReportAllocs()

	memStart := captureMemStats()

	for i := 0; i < b.N; i++ {
		// Perform mixed operations: 50% inserts, 30% gets, 20% updates
		for j := 0; j < config.NumOperations; j++ {
			keyIdx := j % len(keys)
			valueIdx := j % len(values)

			switch j % 10 {
			case 0, 1, 2, 3, 4: // 50% inserts
				tree.Insert(keys[keyIdx], values[valueIdx])
			case 5, 6, 7: // 30% gets
				proof, _ := tree.Get(keys[keyIdx])
				_ = proof
			case 8, 9: // 20% updates
				tree.Update(keys[keyIdx], values[valueIdx])
			}
		}
	}

	memStart.capture()
	b.StopTimer()

	b.Logf("Memory usage: %s", memStart.String())
	b.Logf("Operations per iteration: %d (50%% insert, 30%% get, 20%% update)", config.NumOperations)
	b.Logf("Tree depth: %d", config.TreeDepth)
}

// Concurrent Operations Benchmarks

func BenchmarkConcurrentInserts(b *testing.B) {
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 16)
	if err != nil {
		b.Fatalf("Failed to create tree: %v", err)
	}
	
	numGoroutines := 4
	opsPerGoroutine := 100

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		done := make(chan bool, numGoroutines)
		
		for g := 0; g < numGoroutines; g++ {
			go func(goroutineID int) {
				defer func() { done <- true }()
				
				keys := generateRandomKeys(opsPerGoroutine, int64(42+goroutineID+i))
				values := generateRandomValues(opsPerGoroutine, int64(84+goroutineID+i))
				
				for j := 0; j < opsPerGoroutine; j++ {
					tree.Insert(keys[j], values[j])
				}
			}(g)
		}
		
		// Wait for all goroutines
		for g := 0; g < numGoroutines; g++ {
			<-done
		}
	}
}

// Memory Pool Benchmarks

func BenchmarkMemoryPool_BigInt(b *testing.B) {
	// Create a local memory pool since GetBigIntPool doesn't exist in new API
	pool := &sync.Pool{
		New: func() interface{} {
			return new(big.Int)
		},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		x := pool.Get().(*big.Int)
		x.SetInt64(int64(i))
		y := pool.Get().(*big.Int)
		y.Set(x)
		y.Add(y, big.NewInt(1))
		pool.Put(x)
		pool.Put(y)
	}
}

func BenchmarkWithoutMemoryPool_BigInt(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		x := big.NewInt(int64(i))
		y := new(big.Int).Set(x)
		y.Add(y, big.NewInt(1))
		// Let GC handle cleanup
	}
}

// Example Usage Benchmark

func BenchmarkExampleUsage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db := smt.NewInMemoryDatabase()
		tree, err := smt.NewSparseMerkleTree(db, 16)
		if err != nil {
			b.Fatalf("Failed to create tree: %v", err)
		}
		
		key := big.NewInt(12345)
		value := smt.Bytes32{}
		for j := range value {
			value[j] = byte(j)
		}

		// Insert
		tree.Insert(key, value)

		// Get and verify
		proof, err := tree.Get(key)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
		
		// Verify proof
		valid := tree.VerifyProof(proof)
		if !valid {
			b.Fatal("Proof should be valid")
		}
	}
}