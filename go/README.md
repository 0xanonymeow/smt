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
    key := big.NewInt(1)
    value := smt.Bytes32{1, 2, 3, 4, 5}
    root, err := tree.Insert(key, value)
    if err != nil {
        panic(err)
    }
    
    // Get the proof for a key
    proof, err := tree.Get(key)
    if err != nil {
        panic(err)
    }
    
    if proof.Exists {
        fmt.Printf("Value exists at key %s\n", key.String())
        fmt.Printf("Tree root: %s\n", root)
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
    Exists   bool
    Value    Bytes32
    Index    *big.Int
    Enables  []bool
    Siblings []Bytes32
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
- `Insert(index *big.Int, value Bytes32) (Bytes32, error)`
- `Update(index *big.Int, value Bytes32) (Bytes32, error)`
- `Upsert(index *big.Int, value Bytes32) (Bytes32, error)`
- `Delete(index *big.Int) (Bytes32, error)`
- `Get(index *big.Int) (*Proof, error)`
- `Exists(index *big.Int) (bool, error)`
- `Root() Bytes32`
- `VerifyProof(proof *Proof) bool`

### Key-Value Operations

- `InsertKV(key string, value Bytes32) (Bytes32, error)`
- `UpdateKV(key string, value Bytes32) (Bytes32, error)`
- `DeleteKV(key string) (Bytes32, error)`
- `GetKV(key string) (Bytes32, bool, error)`
- `ExistsKV(key string) (bool, error)`

### Batch Operations

- `BatchInsert(indices []*big.Int, values []Bytes32) (Bytes32, error)`
- `BatchUpdate(indices []*big.Int, values []Bytes32) (Bytes32, error)`
- `BatchUpsert(indices []*big.Int, values []Bytes32) (Bytes32, error)`
- `BatchDelete(indices []*big.Int) (Bytes32, error)`

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