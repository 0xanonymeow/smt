# API Documentation

Complete API reference for both Go and Solidity SMT libraries.

## Libraries Overview

- **[Go API](go.md)** - Complete Go library reference
- **[Solidity API](solidity.md)** - Complete Solidity library reference

## Common Data Structures

Both libraries use identical data structures for cross-platform compatibility:

### Proof Structure

```go
// Go
type Proof struct {
    Exists   bool     `json:"exists"`
    Leaf     string   `json:"leaf"`
    Value    *string  `json:"value"`
    Index    *big.Int `json:"index"`
    Enables  *big.Int `json:"enables"`
    Siblings []string `json:"siblings"`
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

### UpdateProof Structure

```go
// Go
type UpdateProof struct {
    Proof
    NewLeaf string `json:"newLeaf"`
}
```

```solidity
// Solidity
struct UpdateProof {
    bool exists;
    bytes32 leaf;
    bytes32 value;
    uint256 index;
    uint256 enables;
    bytes32[] siblings;
    bytes32 newLeaf;
}
```

## Hash Function Compatibility

Both libraries implement identical keccak256 hash functions:

- **Zero Optimization**: Returns 0 if all inputs are 0
- **Consistent Serialization**: 32-byte big-endian format
- **Cross-Platform Verification**: Proofs generated in one library verify in the other

## Error Handling

Both libraries provide comprehensive error handling with consistent error types and messages.

See individual API documentation for detailed method signatures and usage examples.
