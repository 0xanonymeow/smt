package benchmark

import (
	"math/big"
	"math/rand"

	smt "github.com/0xanonymeow/smt/go"
)

// generateRandomKeys generates random big.Int keys for testing
func generateRandomKeys(count int, seed int64) []*big.Int {
	rand.Seed(seed)
	keys := make([]*big.Int, count)
	for i := 0; i < count; i++ {
		keys[i] = big.NewInt(rand.Int63())
	}
	return keys
}

// generateRandomKeysForDepth generates random keys within the specified tree depth
func generateRandomKeysForDepth(count int, seed int64, depth uint16) []*big.Int {
	rand.Seed(seed)
	keys := make([]*big.Int, count)
	maxVal := new(big.Int).Lsh(big.NewInt(1), uint(depth))
	for i := 0; i < count; i++ {
		keys[i] = new(big.Int).Rand(rand.New(rand.NewSource(seed+int64(i))), maxVal)
	}
	return keys
}

// generateRandomValues generates random Bytes32 values for testing
func generateRandomValues(count int, seed int64) []smt.Bytes32 {
	rand.Seed(seed)
	values := make([]smt.Bytes32, count)
	for i := 0; i < count; i++ {
		for j := 0; j < 32; j++ {
			values[i][j] = byte(rand.Intn(256))
		}
	}
	return values
}