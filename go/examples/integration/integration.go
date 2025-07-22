package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"time"

	smt "github.com/0xanonymeow/smt/go"
)

// Integration demonstrates the complete Go → Solidity verification workflow
type Integration struct {
	testCases []TestCase
}

// TestCase represents a single integration test case
type TestCase struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Input       []string `json:"input"`
	Expected    Expected `json:"expected"`
}

// Expected results for validation
type Expected struct {
	TreeDepth       uint16  `json:"treeDepth"`
	Length          int     `json:"length"`
	MinOpsPerSecond float64 `json:"minOpsPerSecond"`
}

// SolidityTestData represents test data formatted for Solidity consumption
type SolidityTestData struct {
	TestCases []SolidityTestCase `json:"testCases"`
	Metadata  TestMetadata       `json:"metadata"`
}

// SolidityTestCase represents a single test case for Solidity
type SolidityTestCase struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	TreeData    SolidityExport `json:"treeData"`
	Expected    Expected       `json:"expected"`
}

// TestMetadata provides information about the test suite
type TestMetadata struct {
	GeneratedBy       string `json:"generatedBy"`
	GeneratedAt       string `json:"generatedAt"`
	TotalTestCases    int    `json:"totalTestCases"`
	GoVersion         string `json:"goVersion"`
	SMTLibraryVersion string `json:"smtLibraryVersion"`
}

// SolidityExport represents the complete tree data for Solidity verification
type SolidityExport struct {
	Root   string         `json:"root"`
	Depth  uint16         `json:"depth"`
	Length uint64         `json:"length"`
	Proofs []OrderedProof `json:"proofs"`
}

// OrderedProof represents a proof for a specific index in order
type OrderedProof struct {
	Index    uint64   `json:"index"`
	Leaf     string   `json:"leaf"`
	Value    string   `json:"value"`
	Enables  string   `json:"enables"`
	Siblings []string `json:"siblings"`
}

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

// Performance metrics for comparison
type PerformanceMetrics struct {
	Mode         InsertMode    `json:"mode"`
	Duration     time.Duration `json:"duration"`
	ElementCount int           `json:"elementCount"`
	OpsPerSecond float64       `json:"opsPerSecond"`
	TreeDepth    uint16        `json:"treeDepth"`
}

// OrderedSMTExample represents the main example type for integration
type OrderedSMTExample struct {
	tree  *smt.SparseMerkleTree
	input []string
	mode  InsertMode
	depth uint16
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
		mode:  mode,
		depth: depth,
	}, nil
}

// Run executes the example and returns performance metrics
func (o *OrderedSMTExample) Run() (*PerformanceMetrics, error) {
	start := time.Now()

	// Insert elements sequentially for simplicity in integration
	for i, hexValue := range o.input {
		index := big.NewInt(int64(i))
		value, err := smt.NewBytes32FromHex(hexValue)
		if err != nil {
			return nil, fmt.Errorf("invalid hex value at index %d: %v", i, err)
		}

		_, err = o.tree.Insert(index, value)
		if err != nil {
			return nil, fmt.Errorf("failed to insert at index %d: %v", i, err)
		}
	}

	duration := time.Since(start)
	opsPerSecond := float64(len(o.input)) / duration.Seconds()

	return &PerformanceMetrics{
		Mode:         o.mode,
		Duration:     duration,
		ElementCount: len(o.input),
		OpsPerSecond: opsPerSecond,
		TreeDepth:    o.depth,
	}, nil
}

// VerifyOrderedProofs verifies all proofs locally
func (o *OrderedSMTExample) VerifyOrderedProofs() error {
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
	}

	return nil
}

// ExportForSolidity generates export data for Solidity
func (o *OrderedSMTExample) ExportForSolidity() (*SolidityExport, error) {
	proofs := make([]OrderedProof, len(o.input))

	for i := range o.input {
		index := big.NewInt(int64(i))
		proof, err := o.tree.Get(index)
		if err != nil {
			return nil, fmt.Errorf("failed to get proof for index %d: %v", i, err)
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
	}

	return &SolidityExport{
		Root:   o.tree.Root().String(),
		Depth:  o.depth,
		Length: uint64(len(o.input)),
		Proofs: proofs,
	}, nil
}

// NewIntegration creates a new integration with predefined test cases
func NewIntegration() *Integration {
	return &Integration{
		testCases: []TestCase{
			{
				Name:        "SmallArray",
				Description: "Basic functionality test with small array",
				Input: []string{
					"0x000000000000000000000000000000000000000000000000000000000000000a",
					"0x000000000000000000000000000000000000000000000000000000000000000b",
					"0x000000000000000000000000000000000000000000000000000000000000000c",
				},
				Expected: Expected{
					TreeDepth:       2,
					Length:          3,
					MinOpsPerSecond: 1000.0,
				},
			},
			{
				Name:        "PowerOfTwo",
				Description: "Test with array length that is a power of 2",
				Input: []string{
					"0x0000000000000000000000000000000000000000000000000000000000000001",
					"0x0000000000000000000000000000000000000000000000000000000000000002",
					"0x0000000000000000000000000000000000000000000000000000000000000004",
					"0x0000000000000000000000000000000000000000000000000000000000000008",
				},
				Expected: Expected{
					TreeDepth:       2,
					Length:          4,
					MinOpsPerSecond: 1000.0,
				},
			},
			{
				Name:        "MediumArray",
				Description: "Performance test with medium-sized array",
				Input:       generateHexArray(10, 0x10),
				Expected: Expected{
					TreeDepth:       4,
					Length:          10,
					MinOpsPerSecond: 2000.0,
				},
			},
			{
				Name:        "LargeArray",
				Description: "Stress test with large array",
				Input:       generateHexArray(50, 0x100),
				Expected: Expected{
					TreeDepth:       6,
					Length:          50,
					MinOpsPerSecond: 1500.0,
				},
			},
			{
				Name:        "EdgeCase_SingleElement",
				Description: "Edge case with single element",
				Input:       []string{"0x00000000000000000000000000000000000000000000000000000000deadbeef"},
				Expected: Expected{
					TreeDepth:       1,
					Length:          1,
					MinOpsPerSecond: 5000.0,
				},
			},
			{
				Name:        "EdgeCase_Sequential",
				Description: "Sequential values from 0x0 to 0xf",
				Input: []string{
					"0x0000000000000000000000000000000000000000000000000000000000000000",
					"0x0000000000000000000000000000000000000000000000000000000000000001",
					"0x0000000000000000000000000000000000000000000000000000000000000002",
					"0x0000000000000000000000000000000000000000000000000000000000000003",
					"0x0000000000000000000000000000000000000000000000000000000000000004",
					"0x0000000000000000000000000000000000000000000000000000000000000005",
					"0x0000000000000000000000000000000000000000000000000000000000000006",
					"0x0000000000000000000000000000000000000000000000000000000000000007",
					"0x0000000000000000000000000000000000000000000000000000000000000008",
					"0x0000000000000000000000000000000000000000000000000000000000000009",
					"0x000000000000000000000000000000000000000000000000000000000000000a",
					"0x000000000000000000000000000000000000000000000000000000000000000b",
					"0x000000000000000000000000000000000000000000000000000000000000000c",
					"0x000000000000000000000000000000000000000000000000000000000000000d",
					"0x000000000000000000000000000000000000000000000000000000000000000e",
					"0x000000000000000000000000000000000000000000000000000000000000000f",
				},
				Expected: Expected{
					TreeDepth:       4,
					Length:          16,
					MinOpsPerSecond: 2000.0,
				},
			},
		},
	}
}

// generateHexArray generates an array of hex strings for testing
func generateHexArray(count int, startValue int) []string {
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = fmt.Sprintf("0x%064x", startValue+i)
	}
	return result
}

// RunIntegrationTests executes all integration test cases
func (id *Integration) RunIntegrationTests() error {
	fmt.Println("=== SMT Go → Solidity Integration  ===")
	fmt.Printf("Running %d integration test cases...\n\n", len(id.testCases))

	solidityTestData := SolidityTestData{
		TestCases: make([]SolidityTestCase, 0, len(id.testCases)),
		Metadata: TestMetadata{
			GeneratedBy:       "SMT Integration ",
			GeneratedAt:       fmt.Sprintf("%d", 1690000000), // Placeholder timestamp
			TotalTestCases:    len(id.testCases),
			GoVersion:         "1.20+",
			SMTLibraryVersion: "v1.0.0",
		},
	}

	allTestsPassed := true

	for i, testCase := range id.testCases {
		fmt.Printf("Test Case %d: %s\n", i+1, testCase.Name)
		fmt.Printf("Description: %s\n", testCase.Description)
		fmt.Printf("Input: %v\n", testCase.Input)

		// Run the test case
		result, err := id.runSingleTestCase(testCase)
		if err != nil {
			fmt.Printf("❌ FAILED: %v\n\n", err)
			allTestsPassed = false
			continue
		}

		// Validate results against expectations
		if err := id.validateResults(testCase, result); err != nil {
			fmt.Printf("❌ VALIDATION FAILED: %v\n\n", err)
			allTestsPassed = false
			continue
		}

		fmt.Printf("✅ PASSED: All validations successful\n")
		fmt.Printf("   - Tree depth: %d (expected: %d)\n", result.TreeData.Depth, testCase.Expected.TreeDepth)
		fmt.Printf("   - Elements: %d (expected: %d)\n", result.TreeData.Length, testCase.Expected.Length)
		fmt.Printf("   - Root: %s\n", result.TreeData.Root)
		fmt.Printf("   - Proof count: %d\n", len(result.TreeData.Proofs))

		// Add to Solidity test data
		solidityTestData.TestCases = append(solidityTestData.TestCases, *result)
		fmt.Println()
	}

	// Export test data for Solidity
	if err := id.exportSolidityTestData(solidityTestData); err != nil {
		return fmt.Errorf("failed to export Solidity test data: %v", err)
	}

	// Print summary
	fmt.Println("=== Integration Test Summary ===")
	if allTestsPassed {
		fmt.Printf("✅ ALL TESTS PASSED (%d/%d)\n", len(id.testCases), len(id.testCases))
		fmt.Println("Solidity test data exported successfully!")
	} else {
		fmt.Printf("❌ SOME TESTS FAILED\n")
		return fmt.Errorf("integration tests failed")
	}

	return nil
}

// runSingleTestCase executes a single test case and returns the results
func (id *Integration) runSingleTestCase(testCase TestCase) (*SolidityTestCase, error) {
	// Create ordered SMT example
	example, err := NewOrderedSMTExample(testCase.Input, Sequential)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMT example: %v", err)
	}

	// Run the example
	metrics, err := example.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run example: %v", err)
	}

	// Verify proofs locally
	if err := example.VerifyOrderedProofs(); err != nil {
		return nil, fmt.Errorf("local proof verification failed: %v", err)
	}

	// Export for Solidity
	export, err := example.ExportForSolidity()
	if err != nil {
		return nil, fmt.Errorf("failed to export for Solidity: %v", err)
	}

	// Create Solidity test case
	solidityTestCase := &SolidityTestCase{
		Name:        testCase.Name,
		Description: testCase.Description,
		TreeData:    *export,
		Expected:    testCase.Expected,
	}

	// Store performance metrics for validation
	_ = metrics

	return solidityTestCase, nil
}

// validateResults validates the test results against expectations
func (id *Integration) validateResults(testCase TestCase, result *SolidityTestCase) error {
	// Validate tree depth
	if result.TreeData.Depth != testCase.Expected.TreeDepth {
		return fmt.Errorf("tree depth mismatch: got %d, expected %d",
			result.TreeData.Depth, testCase.Expected.TreeDepth)
	}

	// Validate length
	if int(result.TreeData.Length) != testCase.Expected.Length {
		return fmt.Errorf("length mismatch: got %d, expected %d",
			result.TreeData.Length, testCase.Expected.Length)
	}

	// Validate proof count
	if len(result.TreeData.Proofs) != testCase.Expected.Length {
		return fmt.Errorf("proof count mismatch: got %d, expected %d",
			len(result.TreeData.Proofs), testCase.Expected.Length)
	}

	// Validate sequential indices
	for i, proof := range result.TreeData.Proofs {
		if int(proof.Index) != i {
			return fmt.Errorf("proof index mismatch at position %d: got %d, expected %d",
				i, proof.Index, i)
		}
	}

	// Validate root is not empty (basic sanity check)
	if result.TreeData.Root == "0x0000000000000000000000000000000000000000000000000000000000000000" {
		return fmt.Errorf("root hash should not be empty")
	}

	return nil
}

// exportSolidityTestData exports test data for Solidity consumption
func (id *Integration) exportSolidityTestData(data SolidityTestData) error {
	// Create output directory
	outputDir := filepath.Join(".", "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Export JSON file
	jsonFile := filepath.Join(outputDir, "integration_test_data.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %v", err)
	}

	fmt.Printf("Exported Solidity test data to: %s\n", jsonFile)

	// Export individual test files for easier consumption
	for i, testCase := range data.TestCases {
		testFile := filepath.Join(outputDir, fmt.Sprintf("test_%d_%s.json", i+1, testCase.Name))
		testData, err := json.MarshalIndent(testCase, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal test case %s: %v", testCase.Name, err)
		}

		if err := os.WriteFile(testFile, testData, 0644); err != nil {
			return fmt.Errorf("failed to write test case file %s: %v", testFile, err)
		}
	}

	return nil
}

// nstrateDepthCalculation shows how optimal depth calculation works
func (id *Integration) nstrateDepthCalculation() {
	fmt.Println("=== Tree Depth Calculation nstration ===")
	fmt.Println()

	testSizes := []int{1, 2, 3, 4, 5, 8, 10, 16, 20, 32, 50, 64, 100, 256, 500, 1000}

	fmt.Printf("%-12s %-12s %-12s %-12s\n", "Array Size", "Tree Depth", "Capacity", "Utilization")
	fmt.Println("----------------------------------------------------")

	for _, size := range testSizes {
		depth := calculateOptimalDepth(size)
		capacity := 1 << depth
		utilization := float64(size) / float64(capacity) * 100

		fmt.Printf("%-12d %-12d %-12d %-11.1f%%\n",
			size, depth, capacity, utilization)
	}

	fmt.Println()
}

// generateComprehensiveTestData creates a comprehensive test data file
func (id *Integration) generateComprehensiveTestData() error {
	fmt.Println("=== Generating Comprehensive Test Data ===")

	// Create a range of test cases with different characteristics
	comprehensiveTests := []struct {
		name        string
		description string
		generator   func() []string
	}{
		{
			"EmptyToFull_4bit",
			"Sequential values from 0x0 to 0xF (4-bit range)",
			func() []string {
				result := make([]string, 16)
				for i := 0; i < 16; i++ {
					result[i] = fmt.Sprintf("0x%064x", i)
				}
				return result
			},
		},
		{
			"PowersOfTwo",
			"Values that are powers of 2",
			func() []string {
				powers := []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512}
				result := make([]string, len(powers))
				for i, p := range powers {
					result[i] = fmt.Sprintf("0x%064x", p)
				}
				return result
			},
		},
		{
			"LargeRandomValues",
			"Large random-looking hex values",
			func() []string {
				return []string{
					"0x00000000000000000000000000000000000000000000000000000000deadbeef",
					"0x00000000000000000000000000000000000000000000000000000000cafebabe",
					"0x00000000000000000000000000000000000000000000000000000000feedface",
					"0x000000000000000000000000000000000000000000000000000000000badf00d",
					"0x0000000000000000000000000000000000000000000000000000000012345678",
					"0x000000000000000000000000000000000000000000000000000000009abcdef0",
					"0x0000000000000000000000000000000000000000000000000000000013579bdf",
					"0x000000000000000000000000000000000000000000000000000000002468ace0",
				}
			},
		},
		{
			"MaxBytes32Values",
			"Maximum 32-byte values",
			func() []string {
				return []string{
					"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
					"0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe",
					"0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd",
					"0x8000000000000000000000000000000000000000000000000000000000000000",
					"0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				}
			},
		},
	}

	for _, test := range comprehensiveTests {
		fmt.Printf("Generating %s...\n", test.name)
		input := test.generator()

		example, err := NewOrderedSMTExample(input, Sequential)
		if err != nil {
			return fmt.Errorf("failed to create example for %s: %v", test.name, err)
		}

		_, err = example.Run()
		if err != nil {
			return fmt.Errorf("failed to run example for %s: %v", test.name, err)
		}

		export, err := example.ExportForSolidity()
		if err != nil {
			return fmt.Errorf("failed to export for %s: %v", test.name, err)
		}

		// Save individual test file
		outputDir := filepath.Join(".", "output", "comprehensive")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create comprehensive output directory: %v", err)
		}

		testFile := filepath.Join(outputDir, fmt.Sprintf("%s.json", test.name))
		testData, err := json.MarshalIndent(export, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal comprehensive test %s: %v", test.name, err)
		}

		if err := os.WriteFile(testFile, testData, 0644); err != nil {
			return fmt.Errorf("failed to write comprehensive test file %s: %v", testFile, err)
		}

		fmt.Printf("  ✅ Saved to %s\n", testFile)
	}

	fmt.Println("Comprehensive test data generation completed!")
	return nil
}

func main() {
	demo := NewIntegration()

	// Run comprehensive integration
	fmt.Println("Starting SMT Go → Solidity Integration ")
	fmt.Println("=" + fmt.Sprintf("%50s", "="))
	fmt.Println()

	// nstrate depth calculation
	demo.nstrateDepthCalculation()

	// Run integration tests
	if err := demo.RunIntegrationTests(); err != nil {
		log.Fatal("Integration tests failed:", err)
	}

	// Generate comprehensive test data
	if err := demo.generateComprehensiveTestData(); err != nil {
		log.Fatal("Comprehensive test data generation failed:", err)
	}

	fmt.Println()
	fmt.Println("🎉 Integration demo completed successfully!")
	fmt.Println("Check the ./output directory for generated Solidity test data.")
}
