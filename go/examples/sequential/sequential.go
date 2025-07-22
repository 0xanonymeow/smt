package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"sync"
	"time"
	smt "github.com/0xanonymeow/smt/go"
)

// InsertMode defines how elements are inserted into the tree
type InsertMode int

const (
	Sequential InsertMode = iota
	Concurrent
)

// String returns the string representation of InsertMode
func (m InsertMode) String() string {
	switch m {
	case Sequential:
		return "Sequential"
	case Concurrent:
		return "Concurrent"
	default:
		return "Unknown"
	}
}

// OrderedSMTExample demonstrates ordered SMT operations with dynamic depth calculation
type OrderedSMTExample struct {
	tree   *smt.SparseMerkleTree
	input  []string
	depth  uint16
	mode   InsertMode
}

// OrderedProof represents a proof for a specific index in order
type OrderedProof struct {
	Index    uint64   `json:"index"`
	Leaf     string   `json:"leaf"`
	Value    string   `json:"value"`
	Enables  string   `json:"enables"`
	Siblings []string `json:"siblings"`
}

// SolidityExport represents the complete tree data for Solidity verification
type SolidityExport struct {
	Root   string         `json:"root"`
	Depth  uint16         `json:"depth"`
	Length uint64         `json:"length"`
	Proofs []OrderedProof `json:"proofs"`
}

// Performance metrics for comparison
type PerformanceMetrics struct {
	Mode         InsertMode    `json:"mode"`
	Duration     time.Duration `json:"duration"`
	ElementCount int           `json:"elementCount"`
	OpsPerSecond float64       `json:"opsPerSecond"`
	TreeDepth    uint16        `json:"treeDepth"`
}

// NewOrderedSMTExample creates a new ordered SMT example
func NewOrderedSMTExample(input []string, mode InsertMode) (*OrderedSMTExample, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("input cannot be empty")
	}

	depth := calculateOptimalDepth(len(input))
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMT: %v", err)
	}

	return &OrderedSMTExample{
		tree:  tree,
		input: input,
		depth: depth,
		mode:  mode,
	}, nil
}

// calculateOptimalDepth calculates the minimum depth needed to fit all elements
func calculateOptimalDepth(inputLength int) uint16 {
	if inputLength <= 0 {
		return 1
	}
	if inputLength == 1 {
		return 1
	}
	
	// Calculate ceil(log2(inputLength))
	depth := uint16(math.Ceil(math.Log2(float64(inputLength))))
	
	// Ensure minimum depth of 1 and maximum of 256
	if depth < 1 {
		depth = 1
	} else if depth > 256 {
		depth = 256
	}
	
	return depth
}

// Run executes the ordered SMT example
func (o *OrderedSMTExample) Run() (*PerformanceMetrics, error) {
	fmt.Printf("=== Ordered SMT Example ===\n")
	fmt.Printf("Input: %v\n", o.input)
	fmt.Printf("Array length: %d\n", len(o.input))
	fmt.Printf("Calculated optimal depth: %d\n", o.depth)
	fmt.Printf("Tree capacity: %d positions\n", 1<<o.depth)
	fmt.Printf("Insertion mode: %s\n", o.mode)
	fmt.Println()

	var duration time.Duration
	var err error

	switch o.mode {
	case Sequential:
		duration, err = o.insertSequential()
	case Concurrent:
		duration, err = o.insertConcurrent(4) // Use 4 workers
	default:
		return nil, fmt.Errorf("unknown insertion mode: %v", o.mode)
	}

	if err != nil {
		return nil, fmt.Errorf("insertion failed: %v", err)
	}

	// Calculate performance metrics
	opsPerSecond := float64(len(o.input)) / duration.Seconds()
	metrics := &PerformanceMetrics{
		Mode:         o.mode,
		Duration:     duration,
		ElementCount: len(o.input),
		OpsPerSecond: opsPerSecond,
		TreeDepth:    o.depth,
	}

	fmt.Printf("Performance Results:\n")
	fmt.Printf("  Mode: %s\n", metrics.Mode)
	fmt.Printf("  Duration: %v\n", metrics.Duration)
	fmt.Printf("  Elements: %d\n", metrics.ElementCount)
	fmt.Printf("  Ops/sec: %.2f\n", metrics.OpsPerSecond)
	fmt.Printf("  Final root: %s\n", o.tree.Root())
	fmt.Println()

	return metrics, nil
}

// insertSequential inserts elements one by one in order
func (o *OrderedSMTExample) insertSequential() (time.Duration, error) {
	fmt.Println("Starting sequential insertion...")
	
	start := time.Now()
	for i, hexValue := range o.input {
		index := big.NewInt(int64(i))
		value, err := smt.NewBytes32FromHex(hexValue)
		if err != nil {
			return 0, fmt.Errorf("invalid hex value at index %d: %v", i, err)
		}

		_, err = o.tree.Insert(index, value)
		if err != nil {
			return 0, fmt.Errorf("failed to insert at index %d: %v", i, err)
		}

		fmt.Printf("  Inserted index %d: %s\n", i, hexValue)
	}
	duration := time.Since(start)
	
	fmt.Printf("Sequential insertion completed in %v\n\n", duration)
	return duration, nil
}

// insertConcurrent inserts elements using concurrent workers while maintaining order
func (o *OrderedSMTExample) insertConcurrent(numWorkers int) (time.Duration, error) {
	fmt.Printf("Starting concurrent insertion with %d workers...\n", numWorkers)
	
	// Create a thread-safe wrapper for the SMT
	type SafeSMT struct {
		tree *smt.SparseMerkleTree
		mu   sync.Mutex
	}
	
	safeSMT := &SafeSMT{tree: o.tree}
	
	// Create work channel
	type work struct {
		index int
		value string
	}
	
	workChan := make(chan work, len(o.input))
	errorChan := make(chan error, numWorkers)
	var wg sync.WaitGroup
	
	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range workChan {
				index := big.NewInt(int64(job.index))
				value, err := smt.NewBytes32FromHex(job.value)
				if err != nil {
					errorChan <- fmt.Errorf("worker %d: invalid hex value at index %d: %v", 
						workerID, job.index, err)
					return
				}

				safeSMT.mu.Lock()
				_, err = safeSMT.tree.Insert(index, value)
				safeSMT.mu.Unlock()
				
				if err != nil {
					errorChan <- fmt.Errorf("worker %d: failed to insert at index %d: %v", 
						workerID, job.index, err)
					return
				}

				fmt.Printf("  Worker %d inserted index %d: %s\n", workerID, job.index, job.value)
			}
		}(w)
	}
	
	// Send work maintaining order
	start := time.Now()
	for i, hexValue := range o.input {
		workChan <- work{index: i, value: hexValue}
	}
	close(workChan)
	
	// Wait for completion
	wg.Wait()
	duration := time.Since(start)
	
	// Check for errors
	close(errorChan)
	for err := range errorChan {
		return 0, err
	}
	
	fmt.Printf("Concurrent insertion completed in %v\n\n", duration)
	return duration, nil
}

// GenerateOrderedProofs generates proofs for all inserted elements in order
func (o *OrderedSMTExample) GenerateOrderedProofs() ([]OrderedProof, error) {
	fmt.Println("Generating ordered proofs...")
	
	proofs := make([]OrderedProof, len(o.input))
	
	for i := range o.input {
		index := big.NewInt(int64(i))
		proof, err := o.tree.Get(index)
		if err != nil {
			return nil, fmt.Errorf("failed to get proof for index %d: %v", i, err)
		}
		
		if !proof.Exists {
			return nil, fmt.Errorf("proof at index %d should exist but doesn't", i)
		}
		
		// Convert siblings to string array
		siblings := make([]string, len(proof.Siblings))
		for j, sibling := range proof.Siblings {
			siblings[j] = sibling.String()
		}
		
		proofs[i] = OrderedProof{
			Index:    uint64(i),
			Leaf:     proof.Leaf.String(),
			Value:    proof.Value.String(),
			Enables:  proof.Enables.String(),
			Siblings: siblings,
		}
		
		fmt.Printf("  Generated proof for index %d: %d siblings\n", i, len(siblings))
	}
	
	fmt.Printf("Generated %d ordered proofs\n\n", len(proofs))
	return proofs, nil
}

// ExportForSolidity exports the tree data in a format suitable for Solidity verification
func (o *OrderedSMTExample) ExportForSolidity() (*SolidityExport, error) {
	proofs, err := o.GenerateOrderedProofs()
	if err != nil {
		return nil, fmt.Errorf("failed to generate proofs: %v", err)
	}
	
	export := &SolidityExport{
		Root:   o.tree.Root().String(),
		Depth:  o.depth,
		Length: uint64(len(o.input)),
		Proofs: proofs,
	}
	
	return export, nil
}

// VerifyOrderedProofs verifies all proofs locally before export
func (o *OrderedSMTExample) VerifyOrderedProofs() error {
	fmt.Println("Verifying all proofs locally...")
	
	root := o.tree.Root()
	depth := o.tree.Depth()
	
	for i := range o.input {
		index := big.NewInt(int64(i))
		proof, err := o.tree.Get(index)
		if err != nil {
			return fmt.Errorf("failed to get proof for index %d: %v", i, err)
		}
		
		isValid := smt.VerifyProof(root, depth, proof)
		if !isValid {
			return fmt.Errorf("proof verification failed for index %d", i)
		}
		
		fmt.Printf("  ✓ Proof %d verified successfully\n", i)
	}
	
	fmt.Printf("All %d proofs verified successfully!\n\n", len(o.input))
	return nil
}

func main() {
	fmt.Println("=== Ordered SMT Operations Examples ===\n")

	// Example 1: Small array demonstration
	smallExample()
	
	// Example 2: Medium array with performance comparison
	mediumExample()
	
	// Example 3: Large array with different modes
	largeExample()
}

func smallExample() {
	fmt.Println("1. Small Array Example")
	fmt.Println("----------------------")
	
	input := []string{
		"0x000000000000000000000000000000000000000000000000000000000000000a",
		"0x000000000000000000000000000000000000000000000000000000000000000b", 
		"0x000000000000000000000000000000000000000000000000000000000000000c",
		"0x000000000000000000000000000000000000000000000000000000000000000d",
		"0x000000000000000000000000000000000000000000000000000000000000000e",
		"0x000000000000000000000000000000000000000000000000000000000000000f",
	}
	
	// Test sequential mode
	seqExample, err := NewOrderedSMTExample(input, Sequential)
	if err != nil {
		log.Fatal("Failed to create sequential example:", err)
	}
	
	seqMetrics, err := seqExample.Run()
	if err != nil {
		log.Fatal("Sequential example failed:", err)
	}
	
	// Verify proofs
	if err := seqExample.VerifyOrderedProofs(); err != nil {
		log.Fatal("Proof verification failed:", err)
	}
	
	// Export for Solidity
	export, err := seqExample.ExportForSolidity()
	if err != nil {
		log.Fatal("Export failed:", err)
	}
	
	fmt.Println("Solidity Export Preview:")
	jsonData, _ := json.MarshalIndent(export, "", "  ")
	fmt.Printf("%s\n\n", string(jsonData))
	
	_ = seqMetrics // Use the metrics
}

func mediumExample() {
	fmt.Println("2. Medium Array Performance Comparison")
	fmt.Println("--------------------------------------")
	
	// Generate medium-sized input
	input := make([]string, 20)
	for i := 0; i < 20; i++ {
		input[i] = fmt.Sprintf("0x%064x", i+1)
	}
	
	// Test both modes
	modes := []InsertMode{Sequential, Concurrent}
	metrics := make([]*PerformanceMetrics, len(modes))
	
	for i, mode := range modes {
		example, err := NewOrderedSMTExample(input, mode)
		if err != nil {
			log.Fatal("Failed to create example:", err)
		}
		
		metric, err := example.Run()
		if err != nil {
			log.Fatal("Example failed:", err)
		}
		
		metrics[i] = metric
	}
	
	// Compare performance
	fmt.Println("Performance Comparison:")
	for _, metric := range metrics {
		fmt.Printf("  %s: %.2f ops/sec (%v total)\n", 
			metric.Mode, metric.OpsPerSecond, metric.Duration)
	}
	
	if len(metrics) >= 2 {
		speedup := metrics[1].OpsPerSecond / metrics[0].OpsPerSecond
		fmt.Printf("  Concurrent speedup: %.2fx\n", speedup)
	}
	fmt.Println()
}

func largeExample() {
	fmt.Println("3. Large Array Concurrent Processing")
	fmt.Println("------------------------------------")
	
	// Generate large input array
	input := make([]string, 100)
	for i := 0; i < 100; i++ {
		input[i] = fmt.Sprintf("0x%064x", i+1)
	}
	
	example, err := NewOrderedSMTExample(input, Concurrent)
	if err != nil {
		log.Fatal("Failed to create large example:", err)
	}
	
	metrics, err := example.Run()
	if err != nil {
		log.Fatal("Large example failed:", err)
	}
	
	fmt.Printf("Large Array Results:\n")
	fmt.Printf("  Processed %d elements in %v\n", metrics.ElementCount, metrics.Duration)
	fmt.Printf("  Throughput: %.2f ops/sec\n", metrics.OpsPerSecond)
	fmt.Printf("  Tree depth: %d (capacity: %d)\n", metrics.TreeDepth, 1<<metrics.TreeDepth)
	fmt.Printf("  Final root: %s\n", example.tree.Root())
	
	// Quick verification of a few random proofs
	fmt.Println("\nSpot checking proofs...")
	indices := []int{0, 25, 50, 75, 99}
	for _, idx := range indices {
		proof, _ := example.tree.Get(big.NewInt(int64(idx)))
		isValid := smt.VerifyProof(example.tree.Root(), example.tree.Depth(), proof)
		fmt.Printf("  Index %d proof: %v\n", idx, isValid)
	}
	fmt.Println()
}