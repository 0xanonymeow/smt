package main

import (
	"fmt"
	"log"
	"math/big"
	smt "github.com/0xanonymeow/smt/go"
)

func main() {
	fmt.Println("=== Basic SMT Usage Examples ===\n")

	// Example 1: Basic Tree Operations
	basicTreeOperations()

	// Example 2: Proof Generation and Verification
	proofOperations()

	// Example 3: Error Handling
	errorHandlingExample()

	// Example 4: Multiple Operations
	multipleOperationsExample()
}

func basicTreeOperations() {
	fmt.Println("1. Basic Tree Operations")
	fmt.Println("------------------------")

	// Create a new SMT with depth 256 using in-memory database
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256)
	if err != nil {
		log.Fatal("Failed to create tree:", err)
	}
	fmt.Printf("Initial root: %s\n", tree.Root())

	// Insert some values
	index1 := big.NewInt(42)
	value1, err := smt.NewBytes32FromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != nil {
		log.Fatal("Invalid hex:", err)
	}

	proof1, err := tree.Insert(index1, value1)
	if err != nil {
		log.Fatal("Insert failed:", err)
	}

	fmt.Printf("Inserted at index %s\n", index1.String())
	fmt.Printf("New leaf: %s\n", proof1.NewLeaf)
	fmt.Printf("Root after insert: %s\n", tree.Root())

	// Insert another value
	index2 := big.NewInt(100)
	value2, err := smt.NewBytes32FromHex("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")
	if err != nil {
		log.Fatal("Invalid hex:", err)
	}

	_, err = tree.Insert(index2, value2)
	if err != nil {
		log.Fatal("Insert failed:", err)
	}

	fmt.Printf("Inserted at index %s\n", index2.String())
	fmt.Printf("Root after second insert: %s\n", tree.Root())

	// Check if values exist
	exists1, err := tree.Exists(index1)
	if err != nil {
		log.Fatal("Exists check failed:", err)
	}
	exists999, err := tree.Exists(big.NewInt(999))
	if err != nil {
		log.Fatal("Exists check failed:", err)
	}
	fmt.Printf("Index %s exists: %v\n", index1.String(), exists1)
	fmt.Printf("Index %s exists: %v\n", big.NewInt(999).String(), exists999)

	// Update an existing value
	newValue1, err := smt.NewBytes32FromHex("0x1111111111111111111111111111111111111111111111111111111111111111")
	if err != nil {
		log.Fatal("Invalid hex:", err)
	}
	updateProof, err := tree.Update(index1, newValue1)
	if err != nil {
		log.Fatal("Update failed:", err)
	}

	fmt.Printf("Updated index %s\n", index1.String())
	fmt.Printf("Old leaf: %s\n", updateProof.Leaf)
	fmt.Printf("New leaf: %s\n", updateProof.NewLeaf)
	fmt.Printf("Root after update: %s\n", tree.Root())

	fmt.Println()
}

func proofOperations() {
	fmt.Println("2. Proof Generation and Verification")
	fmt.Println("------------------------------------")

	// Create tree and insert data
	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256)
	if err != nil {
		log.Fatal("Failed to create tree:", err)
	}

	index := big.NewInt(123)
	value, err := smt.NewBytes32FromHex("0x5555555555555555555555555555555555555555555555555555555555555555")
	if err != nil {
		log.Fatal("Invalid hex:", err)
	}

	_, err = tree.Insert(index, value)
	if err != nil {
		log.Fatal("Insert failed:", err)
	}

	// Generate proof
	proof, err := tree.Get(index)
	if err != nil {
		log.Fatal("Get failed:", err)
	}

	fmt.Printf("Generated proof for index %s:\n", index.String())
	fmt.Printf("  Exists: %v\n", proof.Exists)
	fmt.Printf("  Leaf: %s\n", proof.Leaf)
	fmt.Printf("  Value: %s\n", proof.Value)
	fmt.Printf("  Index: %s\n", proof.Index.String())
	fmt.Printf("  Enables: %s\n", proof.Enables.String())
	fmt.Printf("  Siblings count: %d\n", len(proof.Siblings))

	// Print first few siblings
	for i := 0; i < 5 && i < len(proof.Siblings); i++ {
		fmt.Printf("  Sibling[%d]: %s\n", i, proof.Siblings[i])
	}

	// Verify the proof
	isValid := smt.VerifyProof(tree.Root(), tree.Depth(), proof)
	fmt.Printf("Proof verification result: %v\n", isValid)

	// Generate proof for non-existent entry
	nonExistentIndex := big.NewInt(999)
	nonExistentProof, err := tree.Get(nonExistentIndex)
	if err != nil {
		log.Fatal("Get failed:", err)
	}

	fmt.Printf("\nProof for non-existent index %s:\n", nonExistentIndex.String())
	fmt.Printf("  Exists: %v\n", nonExistentProof.Exists)
	fmt.Printf("  Leaf: %s\n", nonExistentProof.Leaf)
	fmt.Printf("  Siblings count: %d\n", len(nonExistentProof.Siblings))

	// Verify non-existence proof
	isValidNonExistence := smt.VerifyProof(tree.Root(), tree.Depth(), nonExistentProof)
	fmt.Printf("Non-existence proof verification: %v\n", isValidNonExistence)

	fmt.Println()
}

func errorHandlingExample() {
	fmt.Println("3. Error Handling")
	fmt.Println("-----------------")

	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256)
	if err != nil {
		log.Fatal("Failed to create tree:", err)
	}

	index := big.NewInt(42)
	value, err := smt.NewBytes32FromHex("0x6666666666666666666666666666666666666666666666666666666666666666")
	if err != nil {
		log.Fatal("Invalid hex:", err)
	}

	// Insert a value
	_, err = tree.Insert(index, value)
	if err != nil {
		log.Fatal("Insert failed:", err)
	}

	// Try to insert the same key again (should fail)
	_, err = tree.Insert(index, value)
	if err != nil {
		fmt.Printf("Expected error when inserting duplicate key: %v\n", err)
		fmt.Printf("Error type: %T\n", err)
	}

	// Try to update a non-existent key (should fail)
	nonExistentIndex := big.NewInt(999)
	_, err = tree.Update(nonExistentIndex, value)
	if err != nil {
		fmt.Printf("Expected error when updating non-existent key: %v\n", err)
	}

	// Try to use invalid hex string
	fmt.Println("Testing invalid hex string handling...")
	_, err = smt.NewBytes32FromHex("0xgg123")
	if err != nil {
		fmt.Printf("Expected error with invalid hex: %v\n", err)
	}

	fmt.Println()
}

func multipleOperationsExample() {
	fmt.Println("4. Multiple Operations")
	fmt.Println("----------------------")

	db := smt.NewInMemoryDatabase()
	tree, err := smt.NewSparseMerkleTree(db, 256)
	if err != nil {
		log.Fatal("Failed to create tree:", err)
	}

	// Simulate batch operations by doing multiple operations
	fmt.Println("Simulating batch operations...")

	// Insert multiple values
	for i := 1; i <= 5; i++ {
		index := big.NewInt(int64(i))
		leafStr := fmt.Sprintf("0x%064x", i*1111)
		value, err := smt.NewBytes32FromHex(leafStr)
		if err != nil {
			log.Printf("Invalid hex for index %d: %v", i, err)
			continue
		}
		_, err = tree.Insert(index, value)
		if err != nil {
			log.Printf("Failed to insert at index %d: %v", i, err)
		} else {
			fmt.Printf("  Inserted at index %d\n", i)
		}
	}

	// Get multiple values
	fmt.Println("Retrieving values:")
	for i := 1; i <= 6; i++ {
		index := big.NewInt(int64(i))
		proof, err := tree.Get(index)
		if err != nil {
			log.Printf("Get failed for index %d: %v", i, err)
			continue
		}
		if proof.Exists {
			fmt.Printf("  Index %d exists, value: %s\n", i, proof.Value)
		} else {
			fmt.Printf("  Index %d does not exist\n", i)
		}
	}

	fmt.Printf("Final root after operations: %s\n", tree.Root())

	fmt.Println()
}