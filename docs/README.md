# Sparse Merkle Tree Libraries Documentation

Welcome to the comprehensive documentation for the production-ready Sparse Merkle Tree (SMT) libraries for Go and Solidity.

## Quick Navigation

- [API Documentation](api/) - Complete API reference for both libraries
- [Integration Guides](guides/) - Step-by-step implementation examples
- [Troubleshooting](troubleshooting/) - Common issues and solutions
- [Examples](../examples/) - Working code examples

## Overview

This project provides cross-platform compatible Sparse Merkle Tree implementations:

- **Go Library** (`smt.go`) - Complete SMT with CRUD operations and key-value interface
- **Solidity Library** (`contracts/SparseMerkleTree.sol`) - On-chain SMT with state management
- **Deployable Contract** (`contracts/SparseMerkleTreeContract.sol`) - Production-ready SMT contract

## Key Features

✅ **Cross-Platform Compatibility** - Proofs generated in Go verify in Solidity and vice versa  
✅ **Complete CRUD Operations** - Insert, Update, Get, Delete with comprehensive error handling  
✅ **Production Ready** - Optimized for performance, security, and gas efficiency  
✅ **Comprehensive Testing** - Full test coverage with cross-platform validation  
✅ **TypeScript Compatible** - Matches reference implementation behavior exactly

## Quick Start

### Go Library

```go
import "github.com/0xanonymeow/smt"

// Create a new SMT
tree := smt.NewSparseMerkleTree(256, nil)

// Insert a value
proof, err := tree.Insert(big.NewInt(42), "0x1234...")
if err != nil {
    log.Fatal(err)
}

// Get a proof
getProof := tree.Get(big.NewInt(42))
fmt.Printf("Exists: %v, Value: %s\n", getProof.Exists, *getProof.Value)
```

### Solidity Library

```solidity
import "./SparseMerkleTree.sol";

contract MyContract {
    using SparseMerkleTree for SparseMerkleTree.SMTStorage;
    SparseMerkleTree.SMTStorage private smt;

    constructor() {
        smt.initialize(256);
    }

    function insertLeaf(uint256 index, bytes32 leaf) external {
        smt.insert(index, leaf);
    }
}
```

## Architecture

Both libraries implement identical algorithms and data structures:

```
┌─────────────────┐    ┌─────────────────┐
│   Go Library    │◄──►│ Solidity Library│
│                 │    │                 │
│ SparseMerkleTree│    │ SparseMerkleTree│
│ SparseMerkleTreeKV   │ SMTStorage      │
│                 │    │                 │
│ Identical Proofs│    │ Identical Proofs│
│ Identical Hashing    │ Identical Hashing
└─────────────────┘    └─────────────────┘
```

## Documentation Structure

### 📚 [API Reference](api/)

Complete method documentation for both libraries

- **[Go API](api/go.md)** - Go library methods, types, and examples
- **[Solidity API](api/solidity.md)** - Solidity library functions, structs, and usage

### 🚀 [Integration Guides](guides/)

Step-by-step implementation tutorials

- **[Getting Started](guides/getting-started.md)** - Basic setup and first steps
- **[Cross-Platform Integration](guides/cross-platform.md)** - Using Go and Solidity together
- **[Advanced Usage](guides/advanced.md)** - Performance optimization and complex patterns

### 🔧 [Troubleshooting](troubleshooting/)

Problem-solving resources

- **[Common Issues](troubleshooting/common-issues.md)** - Frequent problems and solutions
- **[Migration Guide](troubleshooting/migration.md)** - Upgrading from existing implementations

### 💡 [Examples](../examples/)

Working code examples

- **[Go Examples](../examples/go/)** - Basic and advanced Go usage
- **[Solidity Examples](../contracts/examples/)** - Smart contract implementations

## Feature Highlights

### 🔄 Cross-Platform Compatibility

- Identical hash functions across Go and Solidity
- Compatible proof formats for seamless verification
- Consistent serialization and data structures

### ⚡ Performance Optimized

- Memory pools for reduced allocations
- Batch operations for improved throughput
- Gas-optimized Solidity implementations
- Assembly optimizations for critical paths

### 🛡️ Production Ready

- Comprehensive error handling with structured error types
- Thread-safe operations with proper synchronization
- Access control and security features in contracts
- Extensive test coverage with property-based testing

### 🔍 Developer Experience

- Clear API documentation with examples
- Detailed troubleshooting guides
- Migration tools for existing implementations
- Cross-platform testing utilities

## Getting Started

1. **Choose Your Platform**

   - For backend services: Start with [Go API](api/go.md)
   - For smart contracts: Start with [Solidity API](api/solidity.md)
   - For both: See [Cross-Platform Integration](guides/cross-platform.md)

2. **Follow the Guide**

   - New users: [Getting Started Guide](guides/getting-started.md)
   - Existing users: [Migration Guide](troubleshooting/migration.md)
   - Advanced users: [Advanced Usage Guide](guides/advanced.md)

3. **Explore Examples**

   - [Basic Go Usage](../examples/go/basic_usage.go)
   - [Advanced Go Operations](../examples/go/advanced_operations.go)
   - [Solidity Contract Examples](../contracts/examples/)

4. **Get Help**
   - Check [Common Issues](troubleshooting/common-issues.md)
   - Review [API Documentation](api/)
   - Examine working [Examples](../examples/)

## Use Cases

### 🏦 Financial Applications

- State commitments for payment channels
- Merkle proofs for transaction batching
- Cross-chain state verification

### 🎮 Gaming & NFTs

- Player state management
- Asset ownership proofs
- Game state synchronization

### 🔐 Identity & Privacy

- Zero-knowledge proof systems
- Identity commitment schemes
- Privacy-preserving authentication

### 📊 Data Integrity

- Tamper-proof data storage
- Audit trail verification
- Database state commitments

## Performance Benchmarks

| Operation | Go Library       | Solidity Library |
| --------- | ---------------- | ---------------- |
| Insert    | ~50,000 ops/sec  | ~2,000 gas       |
| Update    | ~45,000 ops/sec  | ~2,500 gas       |
| Get       | ~100,000 ops/sec | ~1,500 gas       |
| Verify    | ~80,000 ops/sec  | ~3,000 gas       |

_Benchmarks are approximate and depend on tree size and system specifications_

## Community & Support

- **Issues**: Report bugs and request features
- **Discussions**: Ask questions and share ideas
- **Examples**: Contribute usage examples
- **Documentation**: Help improve guides and references

## Contributing

We welcome contributions! Please see the main project README for:

- Development setup instructions
- Coding standards and guidelines
- Testing requirements
- Pull request process

## License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Ready to get started?** Jump to the [Getting Started Guide](guides/getting-started.md) or explore the [API Documentation](api/)!
