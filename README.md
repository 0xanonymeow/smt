# Sparse Merkle Tree Libraries

Production-ready Sparse Merkle Tree implementations in Go and Solidity with cross-platform compatibility.

## Repository Structure

```
smt/
├── go/                 # Go library implementation
│   ├── smt.go         # Main library code
│   ├── internal/      # Internal packages
│   ├── examples/      # Go usage examples
│   └── tests/         # Comprehensive test suite
│
├── contracts/         # Solidity implementation
│   ├── src/          # Contract sources
│   ├── test/         # Contract tests
│   └── examples/     # Solidity examples
│
└── docs/             # Shared documentation
```

## Features

- **Cross-Platform Compatibility**: Proofs generated in Go can be verified in Solidity and vice versa
- **Complete CRUD Operations**: Insert, Update, Get, and Exists methods
- **Performance Optimized**: Memory pooling, batch operations, and gas optimizations
- **Comprehensive Testing**: Unit tests, integration tests, and cross-platform validation
- **Production Ready**: Thread-safe, error handling, and extensive documentation

## Quick Start

### Go Library

```bash
go get github.com/0xanonymeow/smt/go
```

```go
import smt "github.com/0xanonymeow/smt/go"

tree := smt.NewSparseMerkleTree(16, nil)
```

See [go/README.md](go/README.md) for detailed Go documentation.

### Solidity Library

```bash
forge install 0xanonymeow/smt
```

```solidity
import {SparseMerkleTreeLib} from "smt/contracts/src/SparseMerkleTree.sol";
```

See [contracts/README.md](contracts/README.md) for detailed Solidity documentation.

## Cross-Platform Example

### Generate Proof in Go

```go
tree := smt.NewSparseMerkleTree(16, nil)
key := big.NewInt(1)
value := smt.Serialize(big.NewInt(100))
tree.Insert(key, value)

proof := tree.Get(key)
// Export proof for Solidity verification
```

### Verify Proof in Solidity

```solidity
function verifyGoProof(
    bytes32 leaf,
    uint256 index,
    uint256 enables,
    bytes32[] calldata siblings
) external pure returns (bool) {
    return SparseMerkleTreeLib.verifyProof(leaf, index, enables, siblings);
}
```

## Implementation Details

Both implementations follow the same core algorithm:

- **Tree Depth**: Configurable up to 256 levels
- **Hash Function**: Keccak256 with special zero-value handling
- **Proof Format**: Identical structure across platforms
- **Serialization**: Consistent hex encoding with 0x prefix

## Development

### Testing

```bash
# Go tests
cd go && go test ./tests/...

# Solidity tests
cd contracts && forge test
```

### Building

```bash
# Go library
cd go && go build

# Solidity contracts
cd contracts && forge build
```

## Documentation

- [API Documentation](docs/api/)
- [Integration Guide](docs/guides/integration-guide.md)
- [Troubleshooting](docs/troubleshooting/)

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.
