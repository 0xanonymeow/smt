# Sparse Merkle Tree

Sparse Merkle Tree implementations in Go and Solidity with cross-platform proof compatibility.

## Repository Structure

```
smt/
├── go/                          # Go library implementation
│   ├── smt.go                  # Main SMT implementation
│   ├── batch.go                # Batch operations
│   ├── database.go             # Database interface
│   ├── examples/               # Usage examples (basic, advanced, integration, sequential)
│   └── tests/                  # Test suite
│   └── cmd/                    # Command-line utilities
│
├── contracts/                   # Solidity implementation
│   ├── src/SparseMerkleTree.sol # Main contract
│   ├── test/                   # Comprehensive test suite
│   └── examples/               # Solidity usage examples
│
└── docs/                       # Complete documentation
    ├── api/                    # API references
    ├── guides/                 # Integration guides
    └── troubleshooting/        # Common issues and solutions
```

## Features

- **Cross-Platform Compatibility**: Proofs generated in Go can be verified in Solidity
- **Standard SMT Operations**: Insert, Update, Get, Delete, and Exists methods
- **Batch Operations**: Bulk insertions and updates
- **Simple Build System**: Makefile with test, build, and clean targets

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

## Cross-Platform Proof Compatibility

Proofs generated in Go can be verified in Solidity contracts, enabling off-chain computation with on-chain verification.

## Implementation Details

Both implementations follow the same core algorithm:

- **Tree Depth**: Configurable up to 256 levels
- **Hash Function**: Keccak256
- **Proof Format**: Identical structure across platforms
- **Serialization**: Consistent hex encoding with 0x prefix

## Development

### Using the Makefile

```bash
# Run all tests
make test

# Run tests with coverage report
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

### Manual Testing

```bash
# Go tests with coverage
cd go && go test -coverprofile=coverage.out ./tests ./tests/benchmark
cd go && go tool cover -html=coverage.out -o coverage.html

# Solidity tests
cd contracts && forge test

# Cross-platform validation
cd go && go run cmd/generate_test_data.go
cd contracts && forge test --match-test "testGoGeneratedProofs" -vv --ffi
```

### Building

```bash
# Build everything
make build

# Or manually:
cd go && go build ./...
cd contracts && forge build
```

## Documentation

- [API Documentation](docs/api/)
- [Getting Started Guide](docs/guides/getting-started.md)
- [Cross-Platform Integration](docs/guides/cross-platform.md)
- [Advanced Usage](docs/guides/advanced.md)
- [Troubleshooting](docs/troubleshooting/)

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.
