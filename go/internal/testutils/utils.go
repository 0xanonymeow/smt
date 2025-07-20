package testutils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// HexToBytes converts a hex string to bytes, handling both prefixed and non-prefixed formats
func HexToBytes(hexStr string) ([]byte, error) {
	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")
	
	// Ensure even length by padding with leading zero if necessary
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	
	return hex.DecodeString(hexStr)
}

// BytesToHex converts bytes to a hex string with 0x prefix
func BytesToHex(data []byte) string {
	return "0x" + hex.EncodeToString(data)
}

// HexToBigInt converts a hex string to a big.Int
func HexToBigInt(hexStr string) (*big.Int, error) {
	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")
	
	// Handle empty string as zero
	if hexStr == "" {
		return big.NewInt(0), nil
	}
	
	bigInt := new(big.Int)
	bigInt, ok := bigInt.SetString(hexStr, 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex string: %s", hexStr)
	}
	
	return bigInt, nil
}

// BigIntToHex converts a big.Int to a hex string with 0x prefix
func BigIntToHex(bigInt *big.Int) string {
	if bigInt == nil || bigInt.Sign() == 0 {
		return "0x0"
	}
	return "0x" + bigInt.Text(16)
}

// PadHexTo32Bytes pads a hex string to 32 bytes (64 hex characters)
func PadHexTo32Bytes(hexStr string) string {
	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")
	
	// Pad to 64 characters (32 bytes)
	for len(hexStr) < 64 {
		hexStr = "0" + hexStr
	}
	
	return "0x" + hexStr
}

// IsZeroHash checks if a hex string represents a zero hash
func IsZeroHash(hexStr string) bool {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	for _, char := range hexStr {
		if char != '0' {
			return false
		}
	}
	return true
}

// CompareHexStrings compares two hex strings for equality, handling different formats
func CompareHexStrings(hex1, hex2 string) bool {
	// Normalize both strings
	hex1 = strings.TrimPrefix(strings.ToLower(hex1), "0x")
	hex2 = strings.TrimPrefix(strings.ToLower(hex2), "0x")
	
	// Remove leading zeros
	hex1 = strings.TrimLeft(hex1, "0")
	hex2 = strings.TrimLeft(hex2, "0")
	
	// Handle empty strings as zero
	if hex1 == "" {
		hex1 = "0"
	}
	if hex2 == "" {
		hex2 = "0"
	}
	
	return hex1 == hex2
}