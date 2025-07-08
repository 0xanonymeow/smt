package smt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	
	"github.com/ethereum/go-ethereum/crypto"
)

// Hash computes Keccak256 hash of two 32-byte values
func Hash(left, right []byte) []byte {
	data := append(left, right...)
	return crypto.Keccak256(data)
}

// HashBytes32 computes Keccak256 hash of two Bytes32 values
func HashBytes32(left, right Bytes32) Bytes32 {
	result := Hash(left[:], right[:])
	var b32 Bytes32
	copy(b32[:], result)
	return b32
}

// ComputeLeafHash computes the hash for a leaf node
func ComputeLeafHash(index *big.Int, value Bytes32) Bytes32 {
	indexBytes := index.Bytes()
	if len(indexBytes) == 0 {
		indexBytes = []byte{0}
	}
	
	// Pad index to 32 bytes
	paddedIndex := make([]byte, 32)
	copy(paddedIndex[32-len(indexBytes):], indexBytes)
	
	// Concatenate index || value || 1
	data := make([]byte, 65)
	copy(data[0:32], paddedIndex)
	copy(data[32:64], value[:])
	data[64] = 1
	
	result := crypto.Keccak256(data)
	var b32 Bytes32
	copy(b32[:], result)
	return b32
}

// GetBit extracts a bit at given position from a big.Int
func GetBit(value *big.Int, position uint) uint {
	return uint(value.Bit(int(position)))
}

// SetBit sets a bit at given position in a big.Int
func SetBit(value *big.Int, position uint, bit uint) *big.Int {
	result := new(big.Int).Set(value)
	result.SetBit(result, int(position), bit)
	return result
}

// CountSetBits counts the number of set bits in a big.Int
func CountSetBits(value *big.Int) int {
	count := 0
	for i := 0; i < value.BitLen(); i++ {
		if value.Bit(i) == 1 {
			count++
		}
	}
	return count
}

// HexToBytes32 converts hex string to Bytes32
func HexToBytes32(s string) (Bytes32, error) {
	// Remove 0x prefix if present
	if len(s) >= 2 && s[0:2] == "0x" {
		s = s[2:]
	}
	
	// Ensure correct length
	if len(s) != 64 {
		return Bytes32{}, fmt.Errorf("invalid hex string format: %s", s)
	}
	
	bytes, err := hex.DecodeString(s)
	if err != nil { // coverage-ignore
		return Bytes32{}, err
	}
	
	var b32 Bytes32
	copy(b32[:], bytes)
	return b32, nil
}

// Bytes32ToHex converts Bytes32 to hex string with 0x prefix
func Bytes32ToHex(b Bytes32) string {
	return "0x" + hex.EncodeToString(b[:])
}

// BigIntToBytes32 converts big.Int to Bytes32
func BigIntToBytes32(value *big.Int) Bytes32 {
	var b32 Bytes32
	bytes := value.Bytes()
	if len(bytes) > 32 {
		copy(b32[:], bytes[len(bytes)-32:])
	} else {
		copy(b32[32-len(bytes):], bytes)
	}
	return b32
}

// Bytes32ToBigInt converts Bytes32 to big.Int
func Bytes32ToBigInt(b Bytes32) *big.Int {
	return new(big.Int).SetBytes(b[:])
}


