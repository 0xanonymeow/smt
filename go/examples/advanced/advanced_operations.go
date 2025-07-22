package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"runtime"
	"sync"
	"time"
	smt "github.com/0xanonymeow/smt/go"
)

func main() {
	fmt.Println("=== Advanced SMT Operations Examples ===\n")

	// Example 1: Performance Measurement
	performanceMeasurementExample()

	// Example 2: Concurrent Operations
	concurrentOperationsExample()

	// Example 3: Memory Management
	memoryManagementExample()

	// Example 4: Custom Hash Functions
	customHashFunctionExample()

	// Example 5: Cross-Platform Proof Export
	crossPlatformProofExample()

	// Example 6: Tree Synchronization
	treeSynchronizationExample()
}

func performanceMeasurementExample() {
	fmt.Println("1. Performance Measurement")
	fmt.Println("--------------------------")

	// Create standard SMT
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256)
	if err != nil {
		log.Fatal("Failed to create tree:", err)
	}

	// Measure single operations
	numOperations := 1000
	fmt.Printf("Measuring performance for %d operations...\n", numOperations)

	// Insert operations
	start := time.Now()
	for i := 0; i < numOperations; i++ {
		index := big.NewInt(int64(i))
		value := smt.ComputeLeafHash(index, smt.Bytes32{byte(i), byte(i >> 8)})
		_, err := tree.Insert(index, value)
		if err != nil {
			log.Printf("Insert error at %d: %v", i, err)
		}
	}
	insertDuration := time.Since(start)

	fmt.Printf("Insert operations: %v (%.2f ops/sec)\n", 
		insertDuration, float64(numOperations)/insertDuration.Seconds())

	// Get operations
	start = time.Now()
	for i := 0; i < numOperations; i++ {
		index := big.NewInt(int64(i))
		_, _ = tree.Get(index)
	}
	getDuration := time.Since(start)

	fmt.Printf("Get operations: %v (%.2f ops/sec)\n", 
		getDuration, float64(numOperations)/getDuration.Seconds())

	// Verify operations
	start = time.Now()
	verifyCount := 0
	for i := 0; i < 100; i++ { // Verify subset for performance
		index := big.NewInt(int64(i))
		proof, err := tree.Get(index)
		if err == nil && smt.VerifyProof(tree.Root(), tree.Depth(), proof) {
			verifyCount++
		}
	}
	verifyDuration := time.Since(start)

	fmt.Printf("Verify operations: %v (%.2f ops/sec)\n", 
		verifyDuration, float64(100)/verifyDuration.Seconds())
	fmt.Printf("Verified %d/100 proofs successfully\n", verifyCount)

	fmt.Println()
}

func concurrentOperationsExample() {
	fmt.Println("2. Concurrent Operations")
	fmt.Println("------------------------")

	// Create thread-safe SMT wrapper
	type ThreadSafeSMT struct {
		tree *smt.SparseMerkleTree
		mu   sync.RWMutex
	}

	db := smt.NewInMemoryDatabase()
	baseTree, err := smt.NewSparseMerkleTree(db, 256)
	if err != nil {
		log.Fatal("Failed to create tree:", err)
	}

	safeSMT := &ThreadSafeSMT{
		tree: baseTree,
	}

	// Thread-safe insert
	safeInsert := func(index *big.Int, value smt.Bytes32) error {
		safeSMT.mu.Lock()
		defer safeSMT.mu.Unlock()
		_, err := safeSMT.tree.Insert(index, value)
		return err
	}

	// Thread-safe get
	safeGet := func(index *big.Int) (*smt.Proof, error) {
		safeSMT.mu.RLock()
		defer safeSMT.mu.RUnlock()
		return safeSMT.tree.Get(index)
	}

	// Concurrent insertions
	numWorkers := 10
	operationsPerWorker := 100
	var wg sync.WaitGroup

	fmt.Printf("Starting %d concurrent workers, %d operations each...\n", 
		numWorkers, operationsPerWorker)

	start := time.Now()

	for worker := 0; worker < numWorkers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for i := 0; i < operationsPerWorker; i++ {
				index := big.NewInt(int64(workerID*1000 + i))
				value := smt.ComputeLeafHash(index, smt.Bytes32{byte(workerID), byte(i)})

				err := safeInsert(index, value)
				if err != nil {
					log.Printf("Worker %d: Insert failed for index %s: %v", 
						workerID, index.String(), err)
				}
			}

			fmt.Printf("Worker %d completed\n", workerID)
		}(worker)
	}

	wg.Wait()
	duration := time.Since(start)

	totalOperations := numWorkers * operationsPerWorker
	fmt.Printf("Completed %d concurrent operations in %v (%.2f ops/sec)\n", 
		totalOperations, duration, float64(totalOperations)/duration.Seconds())

	// Verify some random entries
	fmt.Println("Verifying random entries...")
	for i := 0; i < 10; i++ {
		workerID := rand.Intn(numWorkers)
		opID := rand.Intn(operationsPerWorker)
		index := big.NewInt(int64(workerID*1000 + opID))

		proof, err := safeGet(index)
		if err == nil && proof.Exists {
			fmt.Printf("  ✓ Index %s exists\n", index.String())
		} else {
			fmt.Printf("  ✗ Index %s missing\n", index.String())
		}
	}

	fmt.Printf("Final root: %s\n", safeSMT.tree.Root())
	fmt.Println()
}

func memoryManagementExample() {
	fmt.Println("3. Memory Management")
	fmt.Println("--------------------")

	// Create memory-efficient SMT with pooling
	pool := &sync.Pool{
		New: func() interface{} {
			return new(big.Int)
		},
	}

	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256)
	if err != nil {
		log.Fatal("Failed to create tree:", err)
	}

	// Memory-efficient insert
	efficientInsert := func(index int64, value int64) error {
		bigIndex := pool.Get().(*big.Int)
		defer func() {
			bigIndex.SetInt64(0)
			pool.Put(bigIndex)
		}()

		bigIndex.SetInt64(index)
		leafValue := smt.ComputeLeafHash(bigIndex, smt.Bytes32{byte(value), byte(value >> 8)})
		_, err := tree.Insert(bigIndex, leafValue)
		return err
	}

	// Perform many operations with memory monitoring
	numOperations := 5000
	fmt.Printf("Performing %d memory-efficient operations...\n", numOperations)

	// Force GC before starting
	runtime.GC()
	runtime.GC()

	var m0 runtime.MemStats
	runtime.ReadMemStats(&m0)
	fmt.Printf("Initial memory: Alloc=%d KB\n", m0.Alloc/1024)

	start := time.Now()
	for i := 0; i < numOperations; i++ {
		err := efficientInsert(int64(i), int64(i+1))
		if err != nil {
			log.Printf("Insert failed at index %d: %v", i, err)
		}

		// Periodic memory stats
		if i%1000 == 0 && i > 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			fmt.Printf("  After %d ops: Alloc=%d KB (delta: %d KB)\n", 
				i, m.Alloc/1024, (m.Alloc-m0.Alloc)/1024)
		}
	}

	duration := time.Since(start)
	fmt.Printf("Completed in %v (%.2f ops/sec)\n", 
		duration, float64(numOperations)/duration.Seconds())

	// Force garbage collection and check final memory
	runtime.GC()
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Final memory: Alloc=%d KB, TotalAlloc=%d KB, NumGC=%d\n", 
		m.Alloc/1024, m.TotalAlloc/1024, m.NumGC)

	fmt.Println()
}

func customHashFunctionExample() {
	fmt.Println("4. Custom Hash Functions")
	fmt.Println("------------------------")

	// Custom database with domain-separated hash function
	type CustomDB struct {
		smt.Database
		domain string
	}

	// Create a custom database wrapper that adds domain separation
	customDB := &CustomDB{
		Database: smt.NewInMemoryDatabase(),
		domain: "MyApplication",
	}

	// Create SMT with standard configuration
	customTree, err := smt.NewSparseMerkleTree(customDB, 256)
	if err != nil {
		log.Fatal("Failed to create custom tree:", err)
	}

	standardDB := smt.NewInMemoryDatabase()
	standardTree, err := smt.NewSparseMerkleTree(standardDB, 256)
	if err != nil {
		log.Fatal("Failed to create standard tree:", err)
	}

	// Insert same data in both trees
	index := big.NewInt(42)
	value, err := smt.NewBytes32FromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != nil {
		log.Fatal("Invalid hex:", err)
	}

	customTree.Insert(index, value)
	standardTree.Insert(index, value)

	fmt.Printf("Same data, different configurations:\n")
	fmt.Printf("  Custom tree root:   %s\n", customTree.Root())
	fmt.Printf("  Standard tree root: %s\n", standardTree.Root())
	
	// Note: Roots will be the same unless the underlying hash function is different
	// The SMT package uses a standard hash function internally

	// Verify proofs work with respective trees
	customProof, _ := customTree.Get(index)
	standardProof, _ := standardTree.Get(index)

	customValid := smt.VerifyProof(customTree.Root(), customTree.Depth(), customProof)
	standardValid := smt.VerifyProof(standardTree.Root(), standardTree.Depth(), standardProof)

	fmt.Printf("  Custom proof valid:   %v\n", customValid)
	fmt.Printf("  Standard proof valid: %v\n", standardValid)

	// Cross-verification should work since hash functions are the same
	crossValid := smt.VerifyProof(customTree.Root(), customTree.Depth(), standardProof)
	fmt.Printf("  Cross-verification:   %v (should be true with same hash)\n", crossValid)

	fmt.Println()
}

func crossPlatformProofExample() {
	fmt.Println("5. Cross-Platform Proof Export")
	fmt.Println("------------------------------")

	// Create tree and insert data
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256)
	if err != nil {
		log.Fatal("Failed to create tree:", err)
	}

	testData := []struct {
		index int64
		value string
	}{
		{1, "0x1111111111111111111111111111111111111111111111111111111111111111"},
		{2, "0x2222222222222222222222222222222222222222222222222222222222222222"},
		{3, "0x3333333333333333333333333333333333333333333333333333333333333333"},
	}

	for _, data := range testData {
		index := big.NewInt(data.index)
		value, err := smt.NewBytes32FromHex(data.value)
		if err != nil {
			log.Printf("Invalid hex: %v", err)
			continue
		}
		_, err = tree.Insert(index, value)
		if err != nil {
			log.Printf("Insert failed: %v", err)
		}
	}

	fmt.Printf("Tree populated with %d entries\n", len(testData))
	fmt.Printf("Root: %s\n", tree.Root())

	// Generate proofs for Solidity verification
	type SolidityProof struct {
		Leaf     string   `json:"leaf"`
		Index    string   `json:"index"`
		Enables  string   `json:"enables"`
		Siblings []string `json:"siblings"`
		Root     string   `json:"root"`
	}

	solidityProofs := make([]SolidityProof, len(testData))

	for i, data := range testData {
		index := big.NewInt(data.index)
		proof, err := tree.Get(index)
		if err != nil {
			log.Printf("Get failed: %v", err)
			continue
		}

		// Convert siblings to strings
		siblingStrs := make([]string, len(proof.Siblings))
		for j, sibling := range proof.Siblings {
			siblingStrs[j] = sibling.String()
		}

		solidityProofs[i] = SolidityProof{
			Leaf:     proof.Leaf.String(),
			Index:    proof.Index.String(),
			Enables:  proof.Enables.String(),
			Siblings: siblingStrs,
			Root:     tree.Root().String(),
		}
	}

	// Export as JSON for Solidity tests
	jsonData, err := json.MarshalIndent(solidityProofs, "", "  ")
	if err != nil {
		log.Printf("JSON marshaling failed: %v", err)
	} else {
		fmt.Println("\nProofs for Solidity verification:")
		fmt.Println(string(jsonData))
	}

	// Verify all proofs in Go
	fmt.Println("\nVerifying proofs in Go:")
	for i, proof := range solidityProofs {
		index, _ := new(big.Int).SetString(proof.Index, 10)
		enables, _ := new(big.Int).SetString(proof.Enables, 10)
		
		// Convert string siblings back to Bytes32
		siblings := make([]smt.Bytes32, len(proof.Siblings))
		for j, sibStr := range proof.Siblings {
			siblings[j], _ = smt.NewBytes32FromHex(sibStr)
		}
		
		leaf, _ := smt.NewBytes32FromHex(proof.Leaf)

		isValid := smt.VerifyProofWithLeaf(tree.Root(), tree.Depth(), leaf, index, enables, siblings)
		fmt.Printf("  Proof %d: %v\n", i+1, isValid)
	}

	fmt.Println()
}

func treeSynchronizationExample() {
	fmt.Println("6. Tree Synchronization")
	fmt.Println("-----------------------")

	// Create two trees
	db1 := smt.NewInMemoryDatabase()
	tree1, err := smt.NewSparseMerkleTree(db1, 256)
	if err != nil {
		log.Fatal("Failed to create tree1:", err)
	}

	db2 := smt.NewInMemoryDatabase()
	tree2, err := smt.NewSparseMerkleTree(db2, 256)
	if err != nil {
		log.Fatal("Failed to create tree2:", err)
	}

	// Add data to tree1
	fmt.Println("Populating tree1...")
	for i := 0; i < 10; i++ {
		index := big.NewInt(int64(i))
		value := smt.ComputeLeafHash(index, smt.Bytes32{byte(i + 100)})
		_, err := tree1.Insert(index, value)
		if err != nil {
			log.Printf("Failed to insert in tree1: %v", err)
		}
	}

	fmt.Printf("Tree1 root: %s\n", tree1.Root())
	fmt.Printf("Tree2 root: %s\n", tree2.Root())
	fmt.Println("Trees are different")

	// Synchronize trees by replaying operations
	fmt.Println("\nSynchronizing trees...")
	for i := 0; i < 10; i++ {
		index := big.NewInt(int64(i))
		proof, err := tree1.Get(index)
		if err != nil {
			log.Printf("Failed to get proof for index %d: %v", i, err)
			continue
		}
		
		if proof.Exists {
			// Insert the value from tree1 into tree2
			_, err := tree2.Insert(index, proof.Value)
			if err != nil {
				log.Printf("Failed to sync index %d: %v", i, err)
			}
		}
	}

	fmt.Printf("\nAfter synchronization:")
	fmt.Printf("Tree1 root: %s\n", tree1.Root())
	fmt.Printf("Tree2 root: %s\n", tree2.Root())
	fmt.Printf("Trees are synchronized: %v\n", tree1.Root() == tree2.Root())

	// Verify some proofs cross-tree
	fmt.Println("\nCross-tree proof verification:")
	for i := 0; i < 3; i++ {
		index := big.NewInt(int64(i))
		proof1, err := tree1.Get(index)
		if err != nil {
			log.Printf("Failed to get proof: %v", err)
			continue
		}
		
		// Verify tree1's proof against tree2
		valid := smt.VerifyProof(tree2.Root(), tree2.Depth(), proof1)
		fmt.Printf("  Index %d proof from tree1 valid in tree2: %v\n", i, valid)
	}

	fmt.Println()
}