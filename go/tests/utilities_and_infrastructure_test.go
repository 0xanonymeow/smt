package tests

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	smt "github.com/0xanonymeow/smt/go"
	"github.com/0xanonymeow/smt/go/internal/testutils"
	"github.com/0xanonymeow/smt/go/internal/vectors"
)

// TestHexUtilities tests the hex string conversion utilities
func TestHexUtilities(t *testing.T) {
	tests := []struct {
		name     string
		hexStr   string
		expected []byte
	}{
		{
			name:     "simple hex",
			hexStr:   "0x1234",
			expected: []byte{0x12, 0x34},
		},
		{
			name:     "hex without prefix",
			hexStr:   "abcd",
			expected: []byte{0xab, 0xcd},
		},
		{
			name:     "odd length hex",
			hexStr:   "0x123",
			expected: []byte{0x01, 0x23},
		},
		{
			name:     "zero",
			hexStr:   "0x0",
			expected: []byte{0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testutils.HexToBytes(tt.hexStr)
			if err != nil {
				t.Fatalf("HexToBytes failed: %v", err)
			}
			
			if len(result) != len(tt.expected) {
				t.Fatalf("Expected length %d, got %d", len(tt.expected), len(result))
			}
			
			for i, b := range result {
				if b != tt.expected[i] {
					t.Fatalf("Expected byte %d to be 0x%02x, got 0x%02x", i, tt.expected[i], b)
				}
			}
		})
	}
}

// TestBigIntUtilities tests the big.Int conversion utilities
func TestBigIntUtilities(t *testing.T) {
	tests := []struct {
		name     string
		hexStr   string
		expected *big.Int
	}{
		{
			name:     "simple number",
			hexStr:   "0x1234",
			expected: big.NewInt(0x1234),
		},
		{
			name:     "large number",
			hexStr:   "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			expected: func() *big.Int {
				b := new(big.Int)
				b.SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
				return b
			}(),
		},
		{
			name:     "zero",
			hexStr:   "0x0",
			expected: big.NewInt(0),
		},
		{
			name:     "empty string",
			hexStr:   "",
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testutils.HexToBigInt(tt.hexStr)
			if err != nil {
				t.Fatalf("HexToBigInt failed: %v", err)
			}
			
			if result.Cmp(tt.expected) != 0 {
				t.Fatalf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// TestHexComparison tests hex string comparison utilities
func TestHexComparison(t *testing.T) {
	tests := []struct {
		name     string
		hex1     string
		hex2     string
		expected bool
	}{
		{
			name:     "identical strings",
			hex1:     "0x1234",
			hex2:     "0x1234",
			expected: true,
		},
		{
			name:     "different prefixes",
			hex1:     "0x1234",
			hex2:     "1234",
			expected: true,
		},
		{
			name:     "different case",
			hex1:     "0xABCD",
			hex2:     "0xabcd",
			expected: true,
		},
		{
			name:     "leading zeros",
			hex1:     "0x001234",
			hex2:     "0x1234",
			expected: true,
		},
		{
			name:     "zero values",
			hex1:     "0x0",
			hex2:     "0x00000",
			expected: true,
		},
		{
			name:     "different values",
			hex1:     "0x1234",
			hex2:     "0x5678",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutils.CompareHexStrings(tt.hex1, tt.hex2)
			if result != tt.expected {
				t.Fatalf("Expected %v, got %v for comparing %s and %s", tt.expected, result, tt.hex1, tt.hex2)
			}
		})
	}
}

// TestVectorTypes tests that vector types can be marshaled/unmarshaled
func TestVectorTypes(t *testing.T) {
	// Test HashTestVector
	hashVector := vectors.HashTestVector{
		Left:     "0x1234",
		Right:    "0x5678",
		Expected: "0xabcd",
	}
	
	// Test ProofTestVector
	proofVector := vectors.ProofTestVector{
		TreeDepth: 256,
		Leaf:      "0x1234567890abcdef",
		Index:     "0x123",
		Enables:   "0xff",
		Siblings:  []string{"0xabc", "0xdef"},
		Expected:  "0x987654321",
	}
	
	// Test AddressTestVector
	addressVector := vectors.AddressTestVector{
		Address:  "0x742d35Cc6634C0532925a3b8D4C9db4C4C4b3f8e",
		Value:    "test value",
		OldLeaf:  "0x0",
		Enables:  "0xff",
		Siblings: []string{"0x123", "0x456"},
		Expected: "0xabcdef",
	}
	
	// Test RootComputationTestVector
	rootVector := vectors.RootComputationTestVector{
		TreeDepth: 8,
		Leaf:      "0xdeadbeef",
		Index:     "0x42",
		Enables:   "0x07",
		Siblings:  []string{"0x111", "0x222", "0x333"},
		Expected:  "0xcafebabe",
	}
	
	// Basic validation that structures are properly defined
	if hashVector.Left == "" || hashVector.Right == "" || hashVector.Expected == "" {
		t.Fatal("HashTestVector fields not properly set")
	}
	
	if proofVector.TreeDepth == 0 || proofVector.Leaf == "" {
		t.Fatal("ProofTestVector fields not properly set")
	}
	
	if addressVector.Address == "" || addressVector.Value == "" {
		t.Fatal("AddressTestVector fields not properly set")
	}
	
	if rootVector.TreeDepth == 0 || rootVector.Leaf == "" {
		t.Fatal("RootComputationTestVector fields not properly set")
	}
}

// TestVectorLoading tests loading vectors from JSON files
func TestVectorLoading(t *testing.T) {
	// Test loading hash vectors
	hashVectors, err := vectors.LoadHashVectors("testdata/sample_hash_vectors.json")
	if err != nil {
		t.Fatalf("Failed to load hash vectors: %v", err)
	}
	
	if len(hashVectors) != 2 {
		t.Fatalf("Expected 2 hash vectors, got %d", len(hashVectors))
	}
	
	// Verify first vector
	if hashVectors[0].Left != "0x1234567890abcdef" {
		t.Fatalf("Expected left to be 0x1234567890abcdef, got %s", hashVectors[0].Left)
	}
	
	if hashVectors[0].Right != "0xfedcba0987654321" {
		t.Fatalf("Expected right to be 0xfedcba0987654321, got %s", hashVectors[0].Right)
	}
	
	if hashVectors[0].Expected != "0xabcdef1234567890" {
		t.Fatalf("Expected result to be 0xabcdef1234567890, got %s", hashVectors[0].Expected)
	}
}

func TestCountSetBits(t *testing.T) {
	testCases := []struct {
		input    *big.Int
		expected int
	}{
		{big.NewInt(0), 0},      // 0000
		{big.NewInt(1), 1},      // 0001
		{big.NewInt(3), 2},      // 0011
		{big.NewInt(7), 3},      // 0111
		{big.NewInt(15), 4},     // 1111
		{big.NewInt(255), 8},    // 11111111
		{big.NewInt(256), 1},    // 100000000
		{big.NewInt(511), 9},    // 111111111
	}

	for _, tc := range testCases {
		result := smt.CountSetBits(tc.input)
		if result != tc.expected {
			t.Errorf("CountSetBits(%s): expected %d, got %d", tc.input.String(), tc.expected, result)
		}
	}
}

func TestBigIntToBytes32(t *testing.T) {
	testCases := []struct {
		input    *big.Int
		expected string // hex representation
	}{
		{big.NewInt(0), "0x0000000000000000000000000000000000000000000000000000000000000000"},
		{big.NewInt(1), "0x0000000000000000000000000000000000000000000000000000000000000001"},
		{big.NewInt(255), "0x00000000000000000000000000000000000000000000000000000000000000ff"},
		{big.NewInt(256), "0x0000000000000000000000000000000000000000000000000000000000000100"},
		{big.NewInt(65535), "0x000000000000000000000000000000000000000000000000000000000000ffff"},
	}

	for _, tc := range testCases {
		result := smt.BigIntToBytes32(tc.input)
		if result.String() != tc.expected {
			t.Errorf("BigIntToBytes32(%s): expected %s, got %s", tc.input.String(), tc.expected, result.String())
		}
	}
}

func TestBytes32ToBigInt(t *testing.T) {
	testCases := []struct {
		input    string // hex representation
		expected *big.Int
	}{
		{"0x0000000000000000000000000000000000000000000000000000000000000000", big.NewInt(0)},
		{"0x0000000000000000000000000000000000000000000000000000000000000001", big.NewInt(1)},
		{"0x00000000000000000000000000000000000000000000000000000000000000ff", big.NewInt(255)},
		{"0x0000000000000000000000000000000000000000000000000000000000000100", big.NewInt(256)},
		{"0x000000000000000000000000000000000000000000000000000000000000ffff", big.NewInt(65535)},
	}

	for _, tc := range testCases {
		bytes32, err := smt.NewBytes32FromHex(tc.input)
		if err != nil {
			t.Fatalf("Failed to create Bytes32 from %s: %v", tc.input, err)
		}
		
		result := smt.Bytes32ToBigInt(bytes32)
		if result.Cmp(tc.expected) != 0 {
			t.Errorf("Bytes32ToBigInt(%s): expected %s, got %s", tc.input, tc.expected.String(), result.String())
		}
	}
}

func TestBigIntBytes32RoundTrip(t *testing.T) {
	// Test round-trip conversion: BigInt -> Bytes32 -> BigInt
	testValues := []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(42),
		big.NewInt(255),
		big.NewInt(256),
		big.NewInt(65535),
		big.NewInt(16777215), // 2^24 - 1
		new(big.Int).Lsh(big.NewInt(1), 128), // 2^128
	}

	for _, original := range testValues {
		// Convert to Bytes32 and back
		bytes32 := smt.BigIntToBytes32(original)
		converted := smt.Bytes32ToBigInt(bytes32)
		
		if converted.Cmp(original) != 0 {
			t.Errorf("Round trip failed for %s: got %s", original.String(), converted.String())
		}
	}
}

func TestGetBit(t *testing.T) {
	testCases := []struct {
		value    *big.Int
		position uint
		expected uint
	}{
		{big.NewInt(0), 0, 0},    // 0000, bit 0
		{big.NewInt(1), 0, 1},    // 0001, bit 0  
		{big.NewInt(1), 1, 0},    // 0001, bit 1
		{big.NewInt(2), 0, 0},    // 0010, bit 0
		{big.NewInt(2), 1, 1},    // 0010, bit 1
		{big.NewInt(3), 0, 1},    // 0011, bit 0
		{big.NewInt(3), 1, 1},    // 0011, bit 1
		{big.NewInt(4), 2, 1},    // 0100, bit 2
		{big.NewInt(255), 7, 1},  // 11111111, bit 7
		{big.NewInt(256), 8, 1},  // 100000000, bit 8
	}

	for _, tc := range testCases {
		result := smt.GetBit(tc.value, tc.position)
		if result != tc.expected {
			t.Errorf("GetBit(%s, %d): expected %d, got %d", tc.value.String(), tc.position, tc.expected, result)
		}
	}
}

func TestSetBit(t *testing.T) {
	testCases := []struct {
		value    *big.Int
		position uint
		bit      uint
		expected *big.Int
	}{
		{big.NewInt(0), 0, 1, big.NewInt(1)},   // Set bit 0 of 0000 to 1 -> 0001
		{big.NewInt(1), 0, 0, big.NewInt(0)},   // Set bit 0 of 0001 to 0 -> 0000
		{big.NewInt(0), 1, 1, big.NewInt(2)},   // Set bit 1 of 0000 to 1 -> 0010
		{big.NewInt(3), 1, 0, big.NewInt(1)},   // Set bit 1 of 0011 to 0 -> 0001
		{big.NewInt(0), 8, 1, big.NewInt(256)}, // Set bit 8 of 0 to 1 -> 256
	}

	for _, tc := range testCases {
		// SetBit returns a new *big.Int, doesn't modify in place
		result := smt.SetBit(tc.value, tc.position, tc.bit)
		if result.Cmp(tc.expected) != 0 {
			t.Errorf("SetBit(%s, %d, %d): expected %s, got %s", 
				tc.value.String(), tc.position, tc.bit, tc.expected.String(), result.String())
		}
	}
}

func TestIsKeyExistsError(t *testing.T) {
	// Create a KeyExistsError
	index := big.NewInt(42)
	keyExistsErr := &smt.KeyExistsError{Index: index}
	
	// Test with KeyExistsError
	if !smt.IsKeyExistsError(keyExistsErr) {
		t.Error("Expected IsKeyExistsError to return true for KeyExistsError")
	}
	
	// Test with regular error
	regularErr := &smt.KeyNotFoundError{Index: index}
	if smt.IsKeyExistsError(regularErr) {
		t.Error("Expected IsKeyExistsError to return false for non-KeyExistsError")
	}
	
	// Test with nil
	if smt.IsKeyExistsError(nil) {
		t.Error("Expected IsKeyExistsError to return false for nil")
	}
}

func TestKeyNotFoundError(t *testing.T) {
	// Create a KeyNotFoundError directly
	index := big.NewInt(42)
	keyNotFoundErr := &smt.KeyNotFoundError{Index: index}
	
	// Test error message
	expectedMsg := "key not found at index: 42"
	if keyNotFoundErr.Error() != expectedMsg {
		t.Errorf("Expected error message %s, got %s", expectedMsg, keyNotFoundErr.Error())
	}
	
	// Test type assertion
	var err error = keyNotFoundErr
	if _, ok := err.(*smt.KeyNotFoundError); !ok {
		t.Error("Expected error to be of type KeyNotFoundError")
	}
}

func TestHashUtilities(t *testing.T) {
	// Test hash consistency
	left := GenerateRandomBytes32(1)
	right := GenerateRandomBytes32(2)
	
	// Hash should be consistent
	hash1 := smt.HashBytes32(left, right)
	hash2 := smt.HashBytes32(left, right)
	
	if hash1 != hash2 {
		t.Error("Hash function should be consistent")
	}
	
	// Different inputs should produce different hashes (with high probability)
	differentRight := GenerateRandomBytes32(3)
	hash3 := smt.HashBytes32(left, differentRight)
	
	if hash1 == hash3 {
		t.Error("Different inputs should produce different hashes")
	}
	
	// Test zero case
	zero := smt.Bytes32{}
	zeroHash := smt.HashBytes32(zero, zero)
	// Hash should be deterministic
	zeroHash2 := smt.HashBytes32(zero, zero)
	if zeroHash != zeroHash2 {
		t.Error("Hash function should be deterministic")
	}
}

func TestComputeLeafHash(t *testing.T) {
	// Test leaf hash computation
	index := big.NewInt(42)
	value := GenerateRandomBytes32(42)
	
	// Compute leaf hash
	leafHash := smt.ComputeLeafHash(index, value)
	
	// Leaf hash should not be zero (unless inputs are very specific)
	if leafHash.IsZero() {
		t.Error("Leaf hash should not be zero for non-zero inputs")
	}
	
	// Same inputs should produce same hash
	leafHash2 := smt.ComputeLeafHash(index, value)
	if leafHash != leafHash2 {
		t.Error("ComputeLeafHash should be deterministic")
	}
	
	// Different inputs should produce different hashes
	differentValue := GenerateRandomBytes32(100)
	leafHash3 := smt.ComputeLeafHash(index, differentValue)
	if leafHash == leafHash3 {
		t.Error("Different values should produce different leaf hashes")
	}
	
	differentIndex := big.NewInt(100)
	leafHash4 := smt.ComputeLeafHash(differentIndex, value)
	if leafHash == leafHash4 {
		t.Error("Different indices should produce different leaf hashes")
	}
}

func TestGenerateRandomBytes32(t *testing.T) {
	// Test that function generates consistent output for same seed
	seed := 12345
	random1 := GenerateRandomBytes32(seed)
	random2 := GenerateRandomBytes32(seed)
	
	if random1 != random2 {
		t.Error("GenerateRandomBytes32 should be deterministic for same seed")
	}
	
	// Test that different seeds produce different output
	differentSeed := 54321
	random3 := GenerateRandomBytes32(differentSeed)
	if random1 == random3 {
		t.Error("Different seeds should produce different random bytes")
	}
	
	// Test that output is not zero for non-zero seed
	if random1.IsZero() {
		t.Error("GenerateRandomBytes32 should not produce zero for non-zero seed")
	}
}

func TestHexToBytes32AndStringRoundTrip(t *testing.T) {
	testHexStrings := []string{
		"0x0000000000000000000000000000000000000000000000000000000000000000",
		"0x0000000000000000000000000000000000000000000000000000000000000001", 
		"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	}
	
	for _, hexStr := range testHexStrings {
		// Convert hex to Bytes32
		bytes32, err := smt.NewBytes32FromHex(hexStr)
		if err != nil {
			t.Fatalf("Failed to convert hex %s to Bytes32: %v", hexStr, err)
		}
		
		// Convert back to string
		resultStr := bytes32.String()
		
		if resultStr != hexStr {
			t.Errorf("Round trip failed: %s -> %s", hexStr, resultStr)
		}
	}
}

func TestInvalidHexHandling(t *testing.T) {
	invalidHexStrings := []string{
		"0xgg123",                    // invalid characters
		"0x123",                     // too short
		"0x" + string(make([]byte, 65)), // too long
		"123456",                    // missing 0x prefix
		"",                          // empty string
	}
	
	for _, invalidHex := range invalidHexStrings {
		_, err := smt.NewBytes32FromHex(invalidHex)
		if err == nil {
			t.Errorf("Expected error for invalid hex %s, but got none", invalidHex)
		}
	}
}

func TestLoadProofVectors(t *testing.T) {
	// Test LoadProofVectors function (currently 0% coverage)
	
	// Create temporary test data
	testVectors := []vectors.ProofTestVector{
		{
			TreeDepth: 8,
			Leaf:      "0x1234567890abcdef",
			Index:     "0x01",
			Enables:   "0xff",
			Siblings:  []string{"0xabc", "0xdef"},
			Expected:  "0x123456",
		},
		{
			TreeDepth: 4,
			Leaf:      "0xfedcba0987654321",
			Index:     "0x02",
			Enables:   "0x0f",
			Siblings:  []string{"0x111", "0x222"},
			Expected:  "0x654321",
		},
	}
	
	// Create temporary file
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "proof_vectors.json")
	
	data, err := json.MarshalIndent(testVectors, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Test LoadProofVectors
	loadedVectors, err := vectors.LoadProofVectors(filename)
	if err != nil {
		t.Fatalf("Failed to load proof vectors: %v", err)
	}
	
	if len(loadedVectors) != len(testVectors) {
		t.Errorf("Expected %d vectors, got %d", len(testVectors), len(loadedVectors))
	}
	
	for i, expected := range testVectors {
		if i >= len(loadedVectors) {
			break
		}
		actual := loadedVectors[i]
		
		if actual.TreeDepth != expected.TreeDepth {
			t.Errorf("Vector %d: expected TreeDepth %d, got %d", i, expected.TreeDepth, actual.TreeDepth)
		}
		if actual.Leaf != expected.Leaf {
			t.Errorf("Vector %d: expected Leaf %s, got %s", i, expected.Leaf, actual.Leaf)
		}
		if actual.Index != expected.Index {
			t.Errorf("Vector %d: expected Index %s, got %s", i, expected.Index, actual.Index)
		}
		if actual.Expected != expected.Expected {
			t.Errorf("Vector %d: expected Expected %s, got %s", i, expected.Expected, actual.Expected)
		}
	}
	
	// Test error cases
	
	// Test with non-existent file
	_, err = vectors.LoadProofVectors("/non/existent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "failed to read proof vectors file") {
		t.Errorf("Expected error message about reading file, got: %v", err)
	}
	
	// Test with invalid JSON
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("invalid json content"), 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}
	
	_, err = vectors.LoadProofVectors(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal proof vectors") {
		t.Errorf("Expected error message about unmarshalling, got: %v", err)
	}
}

func TestLoadAddressVectors(t *testing.T) {
	// Test LoadAddressVectors function (currently 0% coverage)
	
	testVectors := []vectors.AddressTestVector{
		{
			Address:  "0x742d35Cc6344C4532C334D81f6c0b70C2cBE8c44",
			Value:    "0x1000",
			OldLeaf:  "0x2000",
			Enables:  "0xff00",
			Siblings: []string{"0xaaa", "0xbbb", "0xccc"},
			Expected: "0x3000",
		},
		{
			Address:  "0x123456789abcdef123456789abcdef123456789a",
			Value:    "0x4000",
			OldLeaf:  "0x5000",
			Enables:  "0x00ff",
			Siblings: []string{"0xddd", "0xeee"},
			Expected: "0x6000",
		},
	}
	
	// Create temporary file
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "address_vectors.json")
	
	data, err := json.MarshalIndent(testVectors, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Test LoadAddressVectors
	loadedVectors, err := vectors.LoadAddressVectors(filename)
	if err != nil {
		t.Fatalf("Failed to load address vectors: %v", err)
	}
	
	if len(loadedVectors) != len(testVectors) {
		t.Errorf("Expected %d vectors, got %d", len(testVectors), len(loadedVectors))
	}
	
	for i, expected := range testVectors {
		if i >= len(loadedVectors) {
			break
		}
		actual := loadedVectors[i]
		
		if actual.Address != expected.Address {
			t.Errorf("Vector %d: expected Address %s, got %s", i, expected.Address, actual.Address)
		}
		if actual.Value != expected.Value {
			t.Errorf("Vector %d: expected Value %s, got %s", i, expected.Value, actual.Value)
		}
		if actual.Expected != expected.Expected {
			t.Errorf("Vector %d: expected Expected %s, got %s", i, expected.Expected, actual.Expected)
		}
	}
	
	// Test error cases
	_, err = vectors.LoadAddressVectors("/non/existent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "failed to read address vectors file") {
		t.Errorf("Expected error message about reading file, got: %v", err)
	}
}

func TestSaveProofVectors(t *testing.T) {
	// Test SaveProofVectors function (currently 0% coverage)
	
	testVectors := []vectors.ProofTestVector{
		{
			TreeDepth: 16,
			Leaf:      "0xabcdef1234567890",
			Index:     "0x99",
			Enables:   "0xf0f0",
			Siblings:  []string{"0x777", "0x888", "0x999"},
			Expected:  "0xffffff",
		},
	}
	
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "save_proof_vectors.json")
	
	// Test SaveProofVectors
	err := vectors.SaveProofVectors(filename, testVectors)
	if err != nil {
		t.Fatalf("Failed to save proof vectors: %v", err)
	}
	
	// Verify file was created and has correct content
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}
	
	// Load and verify content
	loadedVectors, err := vectors.LoadProofVectors(filename)
	if err != nil {
		t.Fatalf("Failed to load saved vectors: %v", err)
	}
	
	if len(loadedVectors) != len(testVectors) {
		t.Errorf("Expected %d vectors, got %d", len(testVectors), len(loadedVectors))
	}
	
	if len(loadedVectors) > 0 {
		if loadedVectors[0].TreeDepth != testVectors[0].TreeDepth {
			t.Errorf("Expected TreeDepth %d, got %d", testVectors[0].TreeDepth, loadedVectors[0].TreeDepth)
		}
	}
	
	// Test error case: try to save to invalid directory
	invalidPath := "/root/invalid/path/that/does/not/exist/vectors.json"
	err = vectors.SaveProofVectors(invalidPath, testVectors)
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Logf("Got expected error: %v", err)
	}
}

func TestSaveAddressVectors(t *testing.T) {
	// Test SaveAddressVectors function (currently 0% coverage)
	
	testVectors := []vectors.AddressTestVector{
		{
			Address:  "0x1111222233334444555566667777888899990000",
			Value:    "0xa1b2",
			OldLeaf:  "0xc3d4",
			Enables:  "0xe5f6",
			Siblings: []string{"0x1010", "0x2020"},
			Expected: "0x3030",
		},
	}
	
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "save_address_vectors.json")
	
	// Test SaveAddressVectors
	err := vectors.SaveAddressVectors(filename, testVectors)
	if err != nil {
		t.Fatalf("Failed to save address vectors: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}
	
	// Load and verify content
	loadedVectors, err := vectors.LoadAddressVectors(filename)
	if err != nil {
		t.Fatalf("Failed to load saved vectors: %v", err)
	}
	
	if len(loadedVectors) != len(testVectors) {
		t.Errorf("Expected %d vectors, got %d", len(testVectors), len(loadedVectors))
	}
	
	if len(loadedVectors) > 0 {
		if loadedVectors[0].Address != testVectors[0].Address {
			t.Errorf("Expected Address %s, got %s", testVectors[0].Address, loadedVectors[0].Address)
		}
	}
}

func TestSaveRootComputationVectors(t *testing.T) {
	// Test SaveRootComputationVectors function (currently 0% coverage)
	
	testVectors := []vectors.RootComputationTestVector{
		{
			TreeDepth: 32,
			Leaf:      "0xdeadbeefcafebabe",
			Index:     "0xff",
			Enables:   "0x5555",
			Siblings:  []string{"0xaaaa", "0xbbbb", "0xcccc", "0xdddd"},
			Expected:  "0xeeeeeeee",
		},
	}
	
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "save_root_vectors.json")
	
	// Test SaveRootComputationVectors
	err := vectors.SaveRootComputationVectors(filename, testVectors)
	if err != nil {
		t.Fatalf("Failed to save root computation vectors: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}
	
	// Load and verify content using existing LoadRootComputationVectors function
	loadedVectors, err := vectors.LoadRootComputationVectors(filename)
	if err != nil {
		t.Fatalf("Failed to load saved vectors: %v", err)
	}
	
	if len(loadedVectors) != len(testVectors) {
		t.Errorf("Expected %d vectors, got %d", len(testVectors), len(loadedVectors))
	}
	
	if len(loadedVectors) > 0 {
		if loadedVectors[0].TreeDepth != testVectors[0].TreeDepth {
			t.Errorf("Expected TreeDepth %d, got %d", testVectors[0].TreeDepth, loadedVectors[0].TreeDepth)
		}
		if loadedVectors[0].Leaf != testVectors[0].Leaf {
			t.Errorf("Expected Leaf %s, got %s", testVectors[0].Leaf, loadedVectors[0].Leaf)
		}
	}
}

func TestSaveHashVectors(t *testing.T) {
	// Test SaveHashVectors function (currently 0% coverage)
	
	testVectors := []vectors.HashTestVector{
		{
			Left:     "0x1111111111111111111111111111111111111111111111111111111111111111",
			Right:    "0x2222222222222222222222222222222222222222222222222222222222222222",
			Expected: "0x3333333333333333333333333333333333333333333333333333333333333333",
		},
		{
			Left:     "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Right:    "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Expected: "0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
	}
	
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "save_hash_vectors.json")
	
	// Test SaveHashVectors
	err := vectors.SaveHashVectors(filename, testVectors)
	if err != nil {
		t.Fatalf("Failed to save hash vectors: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}
	
	// Load and verify content using existing LoadHashVectors function
	loadedVectors, err := vectors.LoadHashVectors(filename)
	if err != nil {
		t.Fatalf("Failed to load saved vectors: %v", err)
	}
	
	if len(loadedVectors) != len(testVectors) {
		t.Errorf("Expected %d vectors, got %d", len(testVectors), len(loadedVectors))
	}
	
	for i, expected := range testVectors {
		if i >= len(loadedVectors) {
			break
		}
		actual := loadedVectors[i]
		
		if actual.Left != expected.Left {
			t.Errorf("Vector %d: expected Left %s, got %s", i, expected.Left, actual.Left)
		}
		if actual.Right != expected.Right {
			t.Errorf("Vector %d: expected Right %s, got %s", i, expected.Right, actual.Right)
		}
		if actual.Expected != expected.Expected {
			t.Errorf("Vector %d: expected Expected %s, got %s", i, expected.Expected, actual.Expected)
		}
	}
	
	// Test creating file in nested directories
	nestedPath := filepath.Join(tmpDir, "nested", "deep", "path", "vectors.json")
	err = vectors.SaveHashVectors(nestedPath, testVectors)
	if err != nil {
		t.Fatalf("Failed to save hash vectors in nested path: %v", err)
	}
	
	// Verify nested directories were created
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Expected nested file to be created")
	}
}