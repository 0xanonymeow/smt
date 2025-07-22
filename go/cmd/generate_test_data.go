package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	smt "github.com/0xanonymeow/smt/go"
)

func main() {
	// Create a simple 4-bit tree with values 0x0 to 0x3
	tree, err := smt.NewSparseMerkleTree(smt.NewInMemoryDatabase(), 4)
	if err != nil {
		panic(err)
	}

	// Generate 4 random 32-byte values
	var values []smt.Bytes32
	for i := 0; i < 4; i++ {
		randomBytes := make([]byte, 32)
		_, err := rand.Read(randomBytes)
		if err != nil {
			panic(err)
		}
		values = append(values, smt.Bytes32(randomBytes))
	}

	// Insert values
	for i, val := range values {
		_, err := tree.Insert(big.NewInt(int64(i)), val)
		if err != nil {
			panic(err)
		}
	}

	// Generate proofs
	var proofs []map[string]interface{}
	for i := 0; i < 4; i++ {
		proof, err := tree.Get(big.NewInt(int64(i)))
		if err != nil {
			panic(err)
		}

		siblings := make([]string, len(proof.Siblings))
		for j, sibling := range proof.Siblings {
			siblings[j] = "0x" + sibling.Hex()
		}

		proofs = append(proofs, map[string]interface{}{
			"index":    proof.Index.Int64(),
			"leaf":     "0x" + proof.Leaf.Hex(),
			"value":    "0x" + proof.Value.Hex(), 
			"enables":  proof.Enables.String(),
			"siblings": siblings,
		})
	}

	// Create output data
	output := map[string]interface{}{
		"root":   "0x" + tree.Root().Hex(),
		"depth":  tree.Depth(),
		"length": 4,
		"proofs": proofs,
	}

	// Write to contracts directory
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("../contracts/test_data.json", jsonData, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated test data with root: %s\n", output["root"])
	fmt.Println("Random values generated:")
	for i, val := range values {
		fmt.Printf("  [%d]: 0x%s\n", i, val.Hex())
	}
}