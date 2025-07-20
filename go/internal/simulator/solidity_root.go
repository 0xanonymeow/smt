package simulator

import (
	"fmt"
	"math/big"

	"github.com/0xanonymeow/smt/go/internal/testutils"
	"golang.org/x/crypto/sha3"
)

// SolidityRootSimulator simulates the Solidity computeRoot function behavior
type SolidityRootSimulator struct{}

// NewSolidityRootSimulator creates a new instance of the Solidity root simulator
func NewSolidityRootSimulator() *SolidityRootSimulator {
	return &SolidityRootSimulator{}
}

// ComputeRoot simulates the Solidity computeRoot function exactly
// This function mimics the Solidity assembly code behavior for root computation
func (s *SolidityRootSimulator) ComputeRoot(treeDepth uint16, leaf, index, enables string, siblings []string) (string, error) {
	// Validate tree depth (Solidity typically limits to 256)
	if treeDepth > 256 {
		return "", fmt.Errorf("invalid tree depth: %d, maximum is 256", treeDepth)
	}

	// Convert inputs from hex strings to appropriate types
	leafBytes, err := testutils.HexToBytes(leaf)
	if err != nil {
		return "", fmt.Errorf("invalid leaf hex: %w", err)
	}

	indexBig, err := testutils.HexToBigInt(index)
	if err != nil {
		return "", fmt.Errorf("invalid index hex: %w", err)
	}

	enablesBig, err := testutils.HexToBigInt(enables)
	if err != nil {
		return "", fmt.Errorf("invalid enables hex: %w", err)
	}

	// Convert siblings from hex strings to bytes
	siblingBytes := make([][]byte, len(siblings))
	for i, sibling := range siblings {
		siblingBytes[i], err = testutils.HexToBytes(sibling)
		if err != nil {
			return "", fmt.Errorf("invalid sibling hex at index %d: %w", i, err)
		}
	}

	// Validate index is within tree bounds
	maxIndex := new(big.Int).Lsh(big.NewInt(1), uint(treeDepth))
	maxIndex.Sub(maxIndex, big.NewInt(1))
	if indexBig.Cmp(maxIndex) > 0 {
		return "", fmt.Errorf("index %s exceeds maximum for tree depth %d", index, treeDepth)
	}

	// Start with the leaf as the current hash
	currentHash := make([]byte, 32)
	copy(currentHash, leafBytes)

	// Simulate the Solidity assembly logic for root computation
	// The algorithm processes each level of the tree from leaf to root
	siblingIndex := 0
	for level := uint16(0); level < treeDepth; level++ {
		// Check if this level is enabled (bit is set in enables)
		levelBit := new(big.Int).Rsh(enablesBig, uint(level))
		levelBit.And(levelBit, big.NewInt(1))
		
		if levelBit.Cmp(big.NewInt(0)) == 0 {
			// Level not enabled, skip to next level
			continue
		}

		// Check if we have enough siblings
		if siblingIndex >= len(siblingBytes) {
			return "", fmt.Errorf("insufficient siblings: need at least %d, got %d", siblingIndex+1, len(siblingBytes))
		}

		// Get the sibling for this level
		sibling := siblingBytes[siblingIndex]
		siblingIndex++

		// Determine the bit at this level in the index to decide hash order
		indexBit := new(big.Int).Rsh(indexBig, uint(level))
		indexBit.And(indexBit, big.NewInt(1))

		// Compute the hash for this level
		if indexBit.Cmp(big.NewInt(0)) == 0 {
			// Index bit is 0, current hash goes on the left
			currentHash = s.solidityHash(currentHash, sibling)
		} else {
			// Index bit is 1, current hash goes on the right
			currentHash = s.solidityHash(sibling, currentHash)
		}
	}

	return testutils.BytesToHex(currentHash), nil
}

// solidityHash simulates the Solidity hash function behavior
// This matches the special zero-value handling as in Solidity assembly code
func (s *SolidityRootSimulator) solidityHash(left, right []byte) []byte {
	// Special zero-value handling: if both inputs are zero, return zero
	if s.isZeroBytes(left) && s.isZeroBytes(right) {
		return make([]byte, 32) // Return 32 bytes of zeros
	}

	// For non-zero inputs, use keccak256
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(left)
	hasher.Write(right)
	return hasher.Sum(nil)
}

// isZeroBytes checks if a byte slice contains only zeros
func (s *SolidityRootSimulator) isZeroBytes(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// ValidateInputs validates the inputs for computeRoot function
func (s *SolidityRootSimulator) ValidateInputs(treeDepth uint16, leaf, index, enables string, siblings []string) error {
	// Validate tree depth
	if treeDepth > 256 {
		return fmt.Errorf("invalid tree depth: %d, maximum is 256", treeDepth)
	}

	// Validate leaf format
	if _, err := testutils.HexToBytes(leaf); err != nil {
		return fmt.Errorf("invalid leaf hex format: %w", err)
	}

	// Validate index format and range
	indexBig, err := testutils.HexToBigInt(index)
	if err != nil {
		return fmt.Errorf("invalid index hex format: %w", err)
	}

	maxIndex := new(big.Int).Lsh(big.NewInt(1), uint(treeDepth))
	maxIndex.Sub(maxIndex, big.NewInt(1))
	if indexBig.Cmp(maxIndex) > 0 {
		return fmt.Errorf("index %s exceeds maximum for tree depth %d", index, treeDepth)
	}

	// Validate enables format
	if _, err := testutils.HexToBigInt(enables); err != nil {
		return fmt.Errorf("invalid enables hex format: %w", err)
	}

	// Validate siblings format
	for i, sibling := range siblings {
		if _, err := testutils.HexToBytes(sibling); err != nil {
			return fmt.Errorf("invalid sibling hex format at index %d: %w", i, err)
		}
	}

	return nil
}