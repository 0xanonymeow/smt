# Sparse Merkle Tree Go Library

Go implementation of Sparse Merkle Trees with cross-platform proof compatibility.

## Installation

```bash
go get github.com/0xanonymeow/smt/go
```

## Features

- **Standard SMT Operations**: Insert, Update, Get, Delete, and Exists methods
- **Batch Operations**: Bulk insertions and updates
- **Key-Value Support**: String-based keys with automatic hashing
- **Cross-Platform Compatibility**: Generates proofs verifiable in Solidity

## Quick Start


### Basic Example

```go
package main

import (
    "fmt"
    "math/big"
    smt "github.com/0xanonymeow/smt/go"
)

func main() {
    // Create a new tree with depth 16
    db := smt.NewInMemoryDatabase()
    tree, err := smt.NewSparseMerkleTree(db, 16)
    if err != nil {
        panic(err)
    }
    
    // Insert a key-value pair
    index := big.NewInt(1)
    value, err := smt.NewBytes32FromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
    if err != nil {
        panic(err)
    }
    
    updateProof, err := tree.Insert(index, value)
    if err != nil {
        panic(err)
    }
    
    // Get the proof for the index
    proof, err := tree.Get(index)
    if err != nil {
        panic(err)
    }
    
    if proof.Exists {
        fmt.Printf("Value exists at index %s\n", index.String())
        fmt.Printf("Tree root: %s\n", tree.Root())
        fmt.Printf("Proof leaf: %s\n", proof.Leaf)
    }
}
```

## API Reference

### Core Types

```go
// 32-byte value type
type Bytes32 [32]byte

// Proof structure
type Proof struct {
    Exists   bool      // Whether the leaf exists
    Leaf     Bytes32   // Computed leaf hash
    Value    Bytes32   // Raw value stored at the index
    Index    *big.Int  // Tree index
    Enables  *big.Int  // Sibling enable bitmask
    Siblings []Bytes32 // Non-zero sibling hashes
}

// UpdateProof represents the proof data for insert/update operations
type UpdateProof struct {
    Exists   bool
    Leaf     Bytes32
    Value    Bytes32
    Index    *big.Int
    Enables  *big.Int
    Siblings []Bytes32
    NewLeaf  Bytes32   // New leaf hash after operation
}

// Database interface
type Database interface {
    Get(key Bytes32) (Bytes32, error)
    Set(key, value Bytes32) error
    Delete(key Bytes32) error
}
```

### SparseMerkleTree Methods

- `NewSparseMerkleTree(db Database, depth uint16) (*SparseMerkleTree, error)`
- `Insert(index *big.Int, leaf Bytes32) (*UpdateProof, error)`
- `Update(index *big.Int, newLeaf Bytes32) (*UpdateProof, error)`
- `Delete(index *big.Int) (*UpdateProof, error)`
- `Get(index *big.Int) (*Proof, error)`
- `Exists(index *big.Int) (bool, error)`
- `Root() Bytes32`

### Utility Functions

- `NewBytes32FromHex(hex string) (Bytes32, error)`
- `VerifyProof(root Bytes32, depth uint16, proof *Proof) bool`

## Testing

```bash
# Run tests
go test ./tests/...

# Run with coverage
go test -cover ./tests/...

# Run benchmarks
go test -bench=. ./tests/benchmark/...
```

## Performance

The library is optimized for:
- Efficient memory usage with `Bytes32` type
- Fast hash computations
- Minimal allocations

Benchmark results on typical hardware:
- Insert: ~50μs per operation
- Get: ~30μs per operation
- Update: ~50μs per operation

## Cross-Platform Compatibility

This library generates proofs that are compatible with the Solidity implementation. Proofs generated in Go can be verified in Solidity contracts and vice versa.

## License

See the main project LICENSE file.