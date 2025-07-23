# API Documentation

API reference for both Go and Solidity SMT libraries.

## Libraries Overview

- **[Go API](go.md)** - Complete Go library reference with comprehensive examples
- **[Solidity API](solidity.md)** - Complete Solidity library reference with gas optimization notes

## Testing

The APIs include tests for:
- Core functionality (insert, update, get, delete)
- Cross-platform proof compatibility
- Basic performance benchmarks

## Common Data Structures

Both libraries use identical data structures for cross-platform compatibility:

### Bytes32 Type (Go)

```go
// 32-byte fixed-size array for consistent data handling
type Bytes32 [32]byte
```

### Proof Structure

```go
// Go
type Proof struct {
    Exists   bool     `json:"exists"`
    Value    Bytes32  `json:"value"`
    Index    *big.Int `json:"index"`
    Enables  []bool   `json:"enables"`
    Siblings []Bytes32 `json:"siblings"`
}
```

```solidity
// Solidity  
struct Proof {
    bool exists;
    bytes32 leaf;
    bytes32 value;
    uint256 index;
    uint256 enables;
    bytes32[] siblings;
}
```

### Database Interface (Go)

```go
// Database interface for pluggable storage backends
type Database interface {
    Get(key Bytes32) (Bytes32, error)
    Set(key, value Bytes32) error
    Delete(key Bytes32) error
}
```

## Hash Function Compatibility

Both libraries implement identical keccak256 hash functions:

- **Consistent Hashing**: Always uses keccak256 regardless of input values
- **Consistent Serialization**: 32-byte big-endian format
- **Cross-Platform Verification**: Proofs generated in one library verify in the other

## Error Handling

Both libraries provide comprehensive error handling with consistent error types and messages.

See individual API documentation for detailed method signatures and usage examples.
