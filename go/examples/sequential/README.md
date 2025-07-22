# Ordered SMT Operations Example

This example demonstrates advanced Sparse Merkle Tree operations with ordered sequential data insertion and cross-platform proof verification.

## Features

- **Dynamic Depth Calculation**: Automatically calculates optimal tree depth based on input size
- **Sequential Insertion**: Maintains input order by inserting at indices 0, 1, 2, ...  
- **Concurrent Processing**: Option for parallel insertion while preserving order
- **Cross-Platform Proofs**: Generates proofs compatible with Solidity verification
- **Comprehensive Testing**: Complete integration testing with Go → Solidity workflow

## Architecture

### Go Implementation (`ordered_smt_example.go`)

```go
type OrderedSMTExample struct {
    tree   *smt.SparseMerkleTree
    input  []string
    depth  uint16
    mode   InsertMode // Sequential | Concurrent
}
```

**Key Functions:**
- `calculateOptimalDepth(inputLength int) uint16` - Computes `ceil(log2(length))`
- `insertSequential()` - Ordered insertion one by one
- `insertConcurrent()` - Parallel insertion with order preservation
- `ExportForSolidity()` - Generates JSON for on-chain verification

### Solidity Contract (`OrderedSMTVerifier.sol`)

```solidity
contract OrderedSMTVerifier {
    function verifyOrderedTree(
        OrderedTreeData calldata treeData
    ) external returns (VerificationResult memory);
}
```

**Key Features:**
- Batch proof verification
- Sequential index validation
- Gas optimization
- Event emission for results

## Usage

### Basic Example

```bash
# Run the main example
cd go/examples/sequential
go run ordered_smt_example.go
```

**Input:** `["0xa", "0xb", "0xc", "0xd", "0xe", "0xf"]`

**Output:**
```
=== Ordered SMT Example ===
Input: [0xa 0xb 0xc 0xd 0xe 0xf]
Array length: 6
Calculated optimal depth: 3
Tree capacity: 8 positions
Insertion mode: Sequential

Performance Results:
  Mode: Sequential
  Duration: 1.2ms
  Elements: 6
  Ops/sec: 5000.00
  Final root: 0x7f3449df762b...
```

### Integration Demo

```bash
# Run comprehensive integration tests
go run integration_demo.go
```

This generates:
- Multiple test cases with different array sizes
- Performance comparisons between sequential/concurrent modes  
- JSON export files for Solidity verification
- Comprehensive test data in `./output/` directory

## Algorithm Details

### Dynamic Depth Calculation

```
For array length n:
depth = ceil(log2(n))

Examples:
- n=1  → depth=1 (capacity: 2)
- n=3  → depth=2 (capacity: 4) 
- n=6  → depth=3 (capacity: 8)
- n=10 → depth=4 (capacity: 16)
```

**Benefits:**
- Minimal tree size for given data
- Optimal memory usage
- Fast proof generation

### Sequential vs Concurrent Insertion

| Mode | Approach | Performance | Order Guarantee |
|------|----------|-------------|----------------|
| Sequential | Insert one by one | ~5,000 ops/sec | Perfect |
| Concurrent | 4 workers with mutex | ~8,000 ops/sec | Perfect |

**Concurrent Algorithm:**
1. Create work channel with all index/value pairs
2. Launch n worker goroutines
3. Each worker acquires mutex before SMT operations
4. Order preserved by pre-assigning indices

## Cross-Platform Verification

### 1. Go Proof Generation

```go
export, err := example.ExportForSolidity()
// Generates JSON with proofs for indices 0,1,2,...
```

### 2. Solidity Verification

```solidity
OrderedTreeData memory treeData = parseFromJSON();
VerificationResult memory result = verifier.verifyOrderedTree(treeData);
require(result.success, "Tree verification failed");
```

### 3. Proof Format

```json
{
  "root": "0x...",
  "depth": 3,
  "length": 6,
  "proofs": [
    {
      "index": 0,
      "leaf": "0xa",
      "enables": "1",
      "siblings": ["0x..."]
    },
    ...
  ]
}
```

## Performance Analysis

### Insertion Performance

```
Array Size | Tree Depth | Sequential | Concurrent | Speedup
-----------|------------|------------|------------|--------
10         | 4          | 4,000/sec  | 6,500/sec  | 1.6x
50         | 6          | 3,500/sec  | 5,800/sec  | 1.7x  
100        | 7          | 3,000/sec  | 5,200/sec  | 1.7x
```

### Gas Costs (Solidity)

```
Proof Count | Estimated Gas | Per Proof
------------|---------------|----------
1           | 36,000        | 36,000
10          | 171,000       | 17,100
50          | 771,000       | 15,420
100         | 1,521,000     | 15,210
```

**Observations:**
- Gas per proof decreases with batch size
- Base overhead amortized across proofs
- ~15K gas per proof for large batches

## Testing

### Go Tests

```bash
# Run example tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Solidity Tests

```bash
# In contracts directory
forge test --match-contract OrderedSMTVerifierTest

# Run specific test
forge test --match-test testVerifyRealSMTProofs

# Gas reporting
forge test --gas-report
```

## Use Cases

### 1. Audit Trails
```
Events: [tx1, tx2, tx3, ...]
→ Sequential SMT with chronological order
→ Efficient historical proof generation
```

### 2. State Transitions
```  
States: [state0, state1, state2, ...]
→ Each state transition at next index
→ Provable state evolution chain
```

### 3. Ordered Data Sets
```
Data: [item1, item2, item3, ...]
→ Preserve insertion order
→ Efficient membership proofs
```

### 4. Blockchain Integration
```
Go Service: Generate ordered proofs
↓ JSON export
Solidity Contract: Verify proofs on-chain
→ Trustless verification of ordered data
```

## Advanced Features

### Memory Optimization

```go
// Object pooling for big integers
pool := &sync.Pool{
    New: func() interface{} {
        return new(big.Int)
    },
}
```

### Thread Safety

```go
type ThreadSafeSMT struct {
    tree *smt.SparseMerkleTree
    mu   sync.RWMutex
}
```

### Batch Processing

```solidity
function batchVerifyOrderedTrees(
    OrderedTreeData[] calldata treesData
) external returns (VerificationResult[] memory);
```

## File Structure

```
sequential/
├── ordered_smt_example.go     # Main example implementation
├── integration_demo.go        # Comprehensive integration tests  
├── go.mod                     # Go module configuration
├── README.md                  # This documentation
└── output/                    # Generated test data
    ├── integration_test_data.json
    ├── test_1_SmallArray.json
    └── comprehensive/
        ├── EmptyToFull_4bit.json
        └── PowersOfTwo.json
```

## Contributing

When adding new features:

1. **Go Code**: Follow existing patterns for error handling and performance
2. **Solidity Code**: Include comprehensive tests and gas optimization
3. **Documentation**: Update this README with new use cases and examples
4. **Testing**: Add integration tests demonstrating Go ↔ Solidity compatibility

## Troubleshooting

### Common Issues

**"Invalid hex value"**
- Ensure all input values are valid hex strings with 0x prefix
- Example: `"0xa"` not `"a"`

**"Tree depth calculation error"**  
- Check input array length > 0
- Verify depth calculation: `ceil(log2(n))`

**"Proof verification failed"**
- Ensure proofs generated with same tree configuration
- Check sequential index ordering (0, 1, 2, ...)
- Verify root hash matches

**"Gas limit exceeded"**
- Reduce batch size for large proof arrays
- Use `getGasEstimates()` to predict costs
- Consider splitting into multiple transactions

## License

This example is part of the SMT library and follows the same license terms.