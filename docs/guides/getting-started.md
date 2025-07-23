# Getting Started Guide

This guide will help you get up and running with the SMT libraries.

## Table of Contents

- [Build System](#build-system)
- [Installation](#installation)
- [Basic Go Usage](#basic-go-usage)
- [Basic Solidity Usage](#basic-solidity-usage)
- [Testing and Coverage](#testing-and-coverage)
- [Cross-Platform Verification](#cross-platform-verification)
- [Next Steps](#next-steps)

## Build System

Basic build commands:

```bash
# Run all tests (100% coverage)
make test-coverage

# Run cross-platform compatibility tests
make test-cross-platform

# Build all Go code and examples
make build

# Clean generated files
make clean

# Show all available commands
make help
```

## Installation

### Go Library

```go
import smt "github.com/0xanonymeow/smt/go"
```

### Solidity Library

Using Foundry:

```bash
forge install 0xanonymeow/smt
```

Then import:

```solidity
import {SparseMerkleTreeLib} from "smt/contracts/src/SparseMerkleTree.sol";
```

## Basic Go Usage

### Creating a Tree

```go
package main

import (
    "fmt"
    "log"
    "math/big"
    smt "github.com/0xanonymeow/smt/go"
)

func main() {
    // Create a new SMT with in-memory database and depth 16
    db := smt.NewInMemoryDatabase()
    tree, err := smt.NewSparseMerkleTree(db, 16)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Initial root: %s\n", tree.Root())
}
```

### Inserting Values

```go
// Insert a value at index 42
index := big.NewInt(42)
leaf := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

proof, err := tree.Insert(index, leaf)
if err != nil {
    log.Fatal("Insert failed:", err)
}

fmt.Printf("Inserted successfully!\n")
fmt.Printf("New leaf: %s\n", proof.NewLeaf)
fmt.Printf("New root: %s\n", tree.Root())
```

### Retrieving Values

```go
// Get a proof for the inserted value
getProof := tree.Get(index)

if getProof.Exists {
    fmt.Printf("Value exists: %s\n", *getProof.Value)
    fmt.Printf("Leaf hash: %s\n", getProof.Leaf)
    fmt.Printf("Number of siblings: %d\n", len(getProof.Siblings))
} else {
    fmt.Printf("Value does not exist at index %s\n", index.String())
}
```

### Updating Values

```go
// Update the existing value
newLeaf := "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"

updateProof, err := tree.Update(index, newLeaf)
if err != nil {
    log.Fatal("Update failed:", err)
}

fmt.Printf("Updated successfully!\n")
fmt.Printf("Old leaf: %s\n", updateProof.Leaf)
fmt.Printf("New leaf: %s\n", updateProof.NewLeaf)
```

### Key-Value Interface

```go
// Create a key-value SMT
treeKV := smt.NewSparseMerkleTreeKV(nil)

// Insert key-value pairs
key := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
value := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

kvProof, err := treeKV.Insert(key, value)
if err != nil {
    log.Fatal("KV Insert failed:", err)
}

fmt.Printf("KV inserted with leaf: %s\n", kvProof.NewLeaf)

// Retrieve key-value pairs
kvGetProof := treeKV.Get(key)
if kvGetProof.Exists {
    fmt.Printf("KV exists with value: %s\n", *kvGetProof.Value)
}
```

## Basic Solidity Usage

### Using the Library

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./SparseMerkleTree.sol";

contract MyContract {
    using SparseMerkleTree for SparseMerkleTree.SMTStorage;

    SparseMerkleTree.SMTStorage private smt;

    event ValueInserted(uint256 indexed index, bytes32 leaf);
    event ValueUpdated(uint256 indexed index, bytes32 oldLeaf, bytes32 newLeaf);

    constructor() {
        // Initialize with depth 256
        smt.initialize(256);
    }

    function getRoot() external view returns (bytes32) {
        return smt.getRoot();
    }

    function insertValue(uint256 index, bytes32 leaf) external {
        SparseMerkleTree.UpdateProof memory proof = smt.insert(index, leaf);
        emit ValueInserted(index, proof.newLeaf);
    }

    function updateValue(uint256 index, bytes32 newLeaf) external {
        SparseMerkleTree.UpdateProof memory proof = smt.update(index, newLeaf);
        emit ValueUpdated(index, proof.leaf, proof.newLeaf);
    }

    function getValue(uint256 index) external view returns (
        bool exists,
        bytes32 leaf,
        bytes32 value
    ) {
        SparseMerkleTree.Proof memory proof = smt.getView(index);
        return (proof.exists, proof.leaf, proof.value);
    }

    function verifyProof(
        bytes32 leaf,
        uint256 index,
        uint256 enables,
        bytes32[] calldata siblings
    ) external view returns (bool) {
        return smt.verifyProof(leaf, index, enables, siblings);
    }
}
```

### Using the Deployable Contract

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./SparseMerkleTreeContract.sol";

contract MyApp {
    SparseMerkleTreeContract public smt;

    constructor() {
        // Deploy SMT contract with depth 256
        smt = new SparseMerkleTreeContract(256, "MyApp SMT", "1.0.0");
    }

    function insertData(uint256 index, bytes32 data) external {
        // Create leaf hash (you might want to hash your data differently)
        bytes32 leaf = keccak256(abi.encodePacked(data));

        // Insert into SMT
        smt.insert(index, leaf);
    }

    function verifyData(uint256 index, bytes32 data) external view returns (bool) {
        bytes32 leaf = keccak256(abi.encodePacked(data));
        SparseMerkleTree.Proof memory proof = smt.get(index);

        return proof.exists && proof.leaf == leaf;
    }
}
```

### Foundry Testing

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "../src/SparseMerkleTreeContract.sol";

contract SMTTest is Test {
    SparseMerkleTreeContract smt;

    function setUp() public {
        smt = new SparseMerkleTreeContract(256, "Test SMT", "1.0.0");
    }

    function testInsertAndGet() public {
        uint256 index = 42;
        bytes32 leaf = keccak256("test data");

        // Insert
        smt.insert(index, leaf);

        // Verify
        SparseMerkleTree.Proof memory proof = smt.get(index);
        assertTrue(proof.exists);
        assertEq(proof.leaf, leaf);
    }

    function testProofVerification() public {
        uint256 index = 42;
        bytes32 leaf = keccak256("test data");

        // Insert
        smt.insert(index, leaf);

        // Get proof
        SparseMerkleTree.Proof memory proof = smt.get(index);

        // Verify proof
        bool isValid = smt.verifyProof(
            proof.leaf,
            proof.index,
            proof.enables,
            proof.siblings
        );
        assertTrue(isValid);
    }
}
```

## Cross-Platform Verification

One of the key features is cross-platform proof compatibility. Here's how to verify a Go-generated proof in Solidity:

### Go Side: Generate Proof

```go
// Create and populate tree
tree := smt.NewSparseMerkleTree(256, nil)
index := big.NewInt(42)
leaf := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

_, err := tree.Insert(index, leaf)
if err != nil {
    log.Fatal(err)
}

// Generate proof
proof := tree.Get(index)

// Convert to format suitable for Solidity
fmt.Printf("Leaf: %s\n", proof.Leaf)
fmt.Printf("Index: %s\n", proof.Index.String())
fmt.Printf("Enables: %s\n", proof.Enables.String())
fmt.Printf("Siblings: %v\n", proof.Siblings)
fmt.Printf("Root: %s\n", tree.Root())
```

### Solidity Side: Verify Proof

```solidity
function verifyGoProof(
    bytes32 leaf,
    uint256 index,
    uint256 enables,
    bytes32[] calldata siblings,
    bytes32 expectedRoot
) external pure returns (bool) {
    bytes32 computedRoot = SparseMerkleTree.computeRoot(
        256, leaf, index, enables, siblings
    );
    return computedRoot == expectedRoot;
}
```

### JavaScript/TypeScript Integration

You can also integrate with web frontends:

```javascript
// Convert Go proof to JavaScript format
const proof = {
  exists: true,
  leaf: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
  index: "42",
  enables: "123456789",
  siblings: ["0xabc...", "0xdef..."],
};

// Verify in smart contract
const isValid = await contract.verifyProof(
  proof.leaf,
  proof.index,
  proof.enables,
  proof.siblings
);
```

## Error Handling

### Go Error Handling

```go
proof, err := tree.Insert(index, leaf)
if err != nil {
    // Check error type
    var smtErr *smt.SMTError
    if errors.As(err, &smtErr) {
        switch smtErr.Code {
        case 1003: // KeyExists
            fmt.Printf("Key already exists at index %s\n", index.String())
        case 1002: // OutOfRange
            fmt.Printf("Index %s is out of range\n", index.String())
        default:
            fmt.Printf("SMT error: %s\n", smtErr.Message)
        }
    } else {
        fmt.Printf("Unknown error: %v\n", err)
    }
    return
}
```

### Solidity Error Handling

```solidity
function safeInsert(uint256 index, bytes32 leaf) external {
    try smt.insert(index, leaf) returns (SparseMerkleTree.UpdateProof memory proof) {
        emit InsertSuccess(index, proof.newLeaf);
    } catch Error(string memory reason) {
        emit InsertFailed(index, reason);
    } catch (bytes memory lowLevelData) {
        // Handle custom errors
        if (lowLevelData.length >= 4) {
            bytes4 selector = bytes4(lowLevelData);
            if (selector == SparseMerkleTree.KeyExists.selector) {
                emit KeyAlreadyExists(index);
            } else if (selector == SparseMerkleTree.OutOfRange.selector) {
                emit IndexOutOfRange(index);
            }
        }
    }
}
```

## Performance Tips

### Go Performance

```go
// Use batch operations for multiple insertions
processor := smt.NewBatchProcessor(tree)

operations := []smt.BatchOperation{
    {Type: "insert", Index: big.NewInt(1), Value: "0x1111..."},
    {Type: "insert", Index: big.NewInt(2), Value: "0x2222..."},
    {Type: "insert", Index: big.NewInt(3), Value: "0x3333..."},
}

err := processor.ProcessBatch(operations)
if err != nil {
    log.Fatal("Batch processing failed:", err)
}
```

### Solidity Performance

```solidity
// Use batch operations to save gas
uint256[] memory indices = new uint256[](3);
bytes32[] memory leaves = new bytes32[](3);

indices[0] = 1; leaves[0] = 0x1111...;
indices[1] = 2; leaves[1] = 0x2222...;
indices[2] = 3; leaves[2] = 0x3333...;

// Single transaction for multiple insertions
smt.batchInsert(indices, leaves);
```

## Next Steps

Now that you have the basics working:

1. **Read the [Advanced Usage Guide](advanced.md)** for performance optimization and complex scenarios
2. **Check the [Cross-Platform Integration Guide](cross-platform.md)** for detailed multi-platform workflows
3. **Review the [API Documentation](../api/)** for complete method references
4. **Explore the [Examples](../../examples/)** for more complex use cases
5. **See the [Troubleshooting Guide](../troubleshooting/)** if you encounter issues

## Common Patterns

### Data Integrity Verification

```go
// Store data hash in SMT for integrity verification
data := []byte("important data")
dataHash := crypto.Keccak256(data)
leaf := smt.Serialize(new(big.Int).SetBytes(dataHash))

proof, err := tree.Insert(big.NewInt(1), leaf)
if err != nil {
    log.Fatal(err)
}

// Later, verify data integrity
retrievedProof := tree.Get(big.NewInt(1))
if retrievedProof.Exists {
    expectedHash := smt.Deserialize(*retrievedProof.Value)
    actualHash := new(big.Int).SetBytes(crypto.Keccak256(data))

    if expectedHash.Cmp(actualHash) == 0 {
        fmt.Println("Data integrity verified!")
    } else {
        fmt.Println("Data has been tampered with!")
    }
}
```

### State Commitment

```solidity
contract StateCommitment {
    using SparseMerkleTree for SparseMerkleTree.SMTStorage;
    SparseMerkleTree.SMTStorage private stateSMT;

    mapping(uint256 => bytes32) public stateCommitments;
    uint256 public currentEpoch;

    function commitState(uint256 userId, bytes32 stateHash) external {
        stateSMT.insert(userId, stateHash);
        stateCommitments[currentEpoch] = stateSMT.getRoot();
    }

    function finalizeEpoch() external {
        currentEpoch++;
    }

    function verifyHistoricalState(
        uint256 epoch,
        uint256 userId,
        bytes32 stateHash,
        uint256 enables,
        bytes32[] calldata siblings
    ) external view returns (bool) {
        bytes32 historicalRoot = stateCommitments[epoch];
        return SparseMerkleTree._verifyProofAgainstRoot(
            historicalRoot,
            256,
            stateHash,
            userId,
            enables,
            siblings
        );
    }
}
```
