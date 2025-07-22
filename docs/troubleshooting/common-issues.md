# Common Issues and Solutions

This guide covers the most frequently encountered issues when using the SMT libraries and their solutions, including testing, coverage, and build system issues.

## Table of Contents

- [Installation and Setup Issues](#installation-and-setup-issues)
- [Go Library Issues](#go-library-issues)
- [Solidity Library Issues](#solidity-library-issues)
- [Cross-Platform Compatibility Issues](#cross-platform-compatibility-issues)
- [Performance Issues](#performance-issues)
- [Memory Issues](#memory-issues)

## Installation and Setup Issues

### Issue: Import Path Not Found (Go)

**Problem:**

```
go: cannot find module providing package github.com/0xanonymeow/smt
```

**Solution:**

1. Ensure the Go module is properly initialized:

```bash
go mod init github.com/0xanonymeow
go mod tidy
```

2. Check that the import path matches your module structure:

```go
import "github.com/0xanonymeow/smt"  // Adjust path as needed
```

3. If using a local module, use replace directive in go.mod:

```go
module github.com/0xanonymeow

go 1.19

replace github.com/0xanonymeow/smt => ./path/to/smt
```

### Issue: Solidity Compilation Errors

**Problem:**

```
Error: Source "SparseMerkleTree.sol" not found
```

**Solution:**

1. Ensure the contract files are in the correct directory:

```bash
contracts/
├── SparseMerkleTree.sol
└── SparseMerkleTreeContract.sol
```

2. For Foundry projects, check foundry.toml:

```toml
[profile.default]
src = "contracts"
out = "out"
libs = ["lib"]
```

3. Import with correct path:

```solidity
import "./SparseMerkleTree.sol";
// or
import "contracts/SparseMerkleTree.sol";
```

## Go Library Issues

### Issue: Panic on Invalid Hex String

**Problem:**

```
panic: Not valid hex: 0xgg123
```

**Solution:**

1. Always validate hex strings before using:

```go
func validateHex(input string) error {
    if !strings.HasPrefix(input, "0x") {
        return fmt.Errorf("hex string must start with 0x")
    }

    hexPart := input[2:]
    if len(hexPart) == 0 {
        return fmt.Errorf("empty hex string")
    }

    for _, char := range hexPart {
        if !((char >= '0' && char <= '9') ||
             (char >= 'a' && char <= 'f') ||
             (char >= 'A' && char <= 'F')) {
            return fmt.Errorf("invalid hex character: %c", char)
        }
    }

    return nil
}

// Safe deserialization
func safeDeserialize(input string) (*big.Int, error) {
    if err := validateHex(input); err != nil {
        return nil, err
    }

    result, ok := new(big.Int).SetString(input[2:], 16)
    if !ok {
        return nil, fmt.Errorf("failed to parse hex: %s", input)
    }

    return result, nil
}
```

2. Use error handling instead of panics:

```go
func safeInsert(tree *smt.SparseMerkleTree, index *big.Int, leaf string) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic: %v", r)
        }
    }()

    _, err := tree.Insert(index, leaf)
    return err
}
```

### Issue: Memory Leaks with Large Trees

**Problem:**
Memory usage keeps growing when performing many operations.

**Solution:**

1. Use memory pools for frequent operations:

```go
type MemoryEfficientSMT struct {
    tree *smt.SparseMerkleTree
    pool sync.Pool
}

func NewMemoryEfficientSMT(depth int) *MemoryEfficientSMT {
    return &MemoryEfficientSMT{
        tree: smt.NewSparseMerkleTree(depth, nil),
        pool: sync.Pool{
            New: func() interface{} {
                return new(big.Int)
            },
        },
    }
}

func (me *MemoryEfficientSMT) Insert(index int64, leaf string) error {
    bigIndex := me.pool.Get().(*big.Int)
    defer func() {
        bigIndex.SetInt64(0)
        me.pool.Put(bigIndex)
    }()

    bigIndex.SetInt64(index)
    _, err := me.tree.Insert(bigIndex, leaf)
    return err
}
```

2. Periodically clean up unused references:

```go
func (tree *SparseMerkleTree) Cleanup() {
    // Force garbage collection
    runtime.GC()
    runtime.GC() // Call twice for better cleanup
}
```

### Issue: Concurrent Access Panics

**Problem:**

```
panic: concurrent map read and map write
```

**Solution:**

1. Use mutex for thread safety:

```go
type ThreadSafeSMT struct {
    tree *smt.SparseMerkleTree
    mu   sync.RWMutex
}

func (ts *ThreadSafeSMT) Insert(index *big.Int, leaf string) (*smt.UpdateProof, error) {
    ts.mu.Lock()
    defer ts.mu.Unlock()
    return ts.tree.Insert(index, leaf)
}

func (ts *ThreadSafeSMT) Get(index *big.Int) *smt.Proof {
    ts.mu.RLock()
    defer ts.mu.RUnlock()
    return ts.tree.Get(index)
}
```

2. Use separate trees for read and write operations:

```go
type ReadWriteSMT struct {
    readTree  *smt.SparseMerkleTree
    writeTree *smt.SparseMerkleTree
    mu        sync.RWMutex
}

func (rw *ReadWriteSMT) Insert(index *big.Int, leaf string) error {
    rw.mu.Lock()
    defer rw.mu.Unlock()

    _, err := rw.writeTree.Insert(index, leaf)
    if err != nil {
        return err
    }

    // Sync to read tree
    _, err = rw.readTree.Insert(index, leaf)
    return err
}
```

## Solidity Library Issues

### Issue: Out of Gas Errors

**Problem:**

```
Error: Transaction ran out of gas
```

**Solution:**

1. Use batch operations for multiple insertions:

```solidity
// Instead of multiple single operations
function inefficientMultipleInserts() external {
    smt.insert(1, leaf1);  // Each call uses gas
    smt.insert(2, leaf2);
    smt.insert(3, leaf3);
}

// Use batch operation
function efficientBatchInsert() external {
    uint256[] memory indices = new uint256[](3);
    bytes32[] memory leaves = new bytes32[](3);

    indices[0] = 1; leaves[0] = leaf1;
    indices[1] = 2; leaves[1] = leaf2;
    indices[2] = 3; leaves[2] = leaf3;

    smt.batchInsert(indices, leaves);
}
```

2. Implement gas-aware operations:

```solidity
function gasAwareInsert(
    uint256[] calldata indices,
    bytes32[] calldata leaves
) external {
    uint256 gasLimit = 200000; // Reserve gas for transaction completion

    for (uint256 i = 0; i < indices.length && gasleft() > gasLimit; i++) {
        try smt.insert(indices[i], leaves[i]) {
            // Success
        } catch {
            // Log failure and continue
            emit InsertFailed(indices[i]);
        }
    }
}
```

### Issue: Stack Too Deep Errors

**Problem:**

```
CompilerError: Stack too deep, try removing local variables
```

**Solution:**

1. Reduce local variables by using structs:

```solidity
// Instead of many local variables
function stackTooDeep(
    uint256 a, uint256 b, uint256 c, uint256 d,
    bytes32 e, bytes32 f, bytes32 g, bytes32 h
) external {
    // Too many variables cause stack issues
}

// Use struct to group related data
struct OperationData {
    uint256 a;
    uint256 b;
    uint256 c;
    uint256 d;
    bytes32 e;
    bytes32 f;
    bytes32 g;
    bytes32 h;
}

function stackOptimized(OperationData calldata data) external {
    // Use data.a, data.b, etc.
}
```

2. Break complex functions into smaller ones:

```solidity
function complexOperation(uint256 index, bytes32 leaf) external {
    _validateInput(index, leaf);
    _performOperation(index, leaf);
    _emitEvents(index, leaf);
}

function _validateInput(uint256 index, bytes32 leaf) private pure {
    require(index > 0, "Invalid index");
    require(leaf != bytes32(0), "Invalid leaf");
}

function _performOperation(uint256 index, bytes32 leaf) private {
    smt.insert(index, leaf);
}

function _emitEvents(uint256 index, bytes32 leaf) private {
    emit LeafInserted(index, leaf);
}
```

### Issue: Invalid Proof Verification

**Problem:**
Proofs that should be valid are failing verification.

**Solution:**

1. Check proof format and ordering:

```solidity
function debugProofVerification(
    bytes32 leaf,
    uint256 index,
    uint256 enables,
    bytes32[] calldata siblings
) external view returns (bool, bytes32) {
    // Return both result and computed root for debugging
    bytes32 computedRoot = SparseMerkleTree.computeRoot(
        smt.depth,
        leaf,
        index,
        enables,
        siblings
    );

    bool isValid = computedRoot == smt.getRoot();
    return (isValid, computedRoot);
}
```

2. Verify sibling ordering matches the implementation:

```solidity
function verifySiblingOrder(
    uint256 index,
    bytes32[] calldata siblings,
    uint256 enables
) external pure returns (bool) {
    uint256 siblingIndex = 0;

    for (uint256 i = 0; i < 256; i++) {
        if ((enables >> i) & 1 == 1) {
            if (siblingIndex >= siblings.length) {
                return false; // Not enough siblings
            }

            // Verify sibling is non-zero
            if (siblings[siblingIndex] == bytes32(0)) {
                return false; // Zero sibling should not be in array
            }

            siblingIndex++;
        }
    }

    return siblingIndex == siblings.length; // All siblings used
}
```

## Cross-Platform Compatibility Issues

### Issue: Root Mismatch Between Go and Solidity

**Problem:**
Same operations produce different roots in Go and Solidity.

**Solution:**

1. Verify hash function consistency:

```go
// Go test
func TestHashConsistency(t *testing.T) {
    left := big.NewInt(123)
    right := big.NewInt(456)

    goHash := smt.Keccak(left, right)
    expectedSolidityHash := "0x..." // From Solidity test

    assert.Equal(t, expectedSolidityHash, smt.Serialize(goHash))
}
```

```solidity
// Solidity test
function testHashConsistency() external pure {
    bytes32 left = bytes32(uint256(123));
    bytes32 right = bytes32(uint256(456));

    bytes32 result = SparseMerkleTree.hash(left, right);
    bytes32 expected = 0x...; // From Go test

    assert(result == expected);
}
```

2. Check serialization format:

```go
func TestSerializationCompatibility(t *testing.T) {
    value := big.NewInt(12345)
    serialized := smt.Serialize(value)

    // Should be 32-byte hex string with 0x prefix
    assert.True(t, strings.HasPrefix(serialized, "0x"))
    assert.Equal(t, 66, len(serialized)) // 0x + 64 hex chars

    // Should deserialize back to same value
    deserialized := smt.Deserialize(serialized)
    assert.Equal(t, value.String(), deserialized.String())
}
```

### Issue: Proof Format Incompatibility

**Problem:**
Proofs generated in Go don't verify in Solidity.

**Solution:**

1. Ensure consistent proof structure:

```go
type CompatibleProof struct {
    Exists   bool     `json:"exists"`
    Leaf     string   `json:"leaf"`     // 0x-prefixed hex
    Value    *string  `json:"value"`    // 0x-prefixed hex or null
    Index    string   `json:"index"`    // Decimal string
    Enables  string   `json:"enables"`  // Decimal string
    Siblings []string `json:"siblings"` // Array of 0x-prefixed hex
}

func ConvertToCompatibleProof(proof *smt.Proof) *CompatibleProof {
    var value *string
    if proof.Value != nil {
        value = proof.Value
    }

    return &CompatibleProof{
        Exists:   proof.Exists,
        Leaf:     proof.Leaf,
        Value:    value,
        Index:    proof.Index.String(),
        Enables:  proof.Enables.String(),
        Siblings: proof.Siblings,
    }
}
```

2. Test cross-platform verification:

```go
func TestCrossPlatformVerification(t *testing.T) {
    // Create tree and insert data
    tree := smt.NewSparseMerkleTree(256, nil)
    index := big.NewInt(42)
    leaf := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

    _, err := tree.Insert(index, leaf)
    require.NoError(t, err)

    // Get proof
    proof := tree.Get(index)

    // Verify in Go
    isValidGo := tree.VerifyProof(proof.Leaf, proof.Index, proof.Enables, proof.Siblings)
    require.True(t, isValidGo)

    // Convert to format suitable for Solidity testing
    compatibleProof := ConvertToCompatibleProof(proof)
    proofJSON, _ := json.Marshal(compatibleProof)

    // This JSON can be used in Solidity tests
    t.Logf("Proof for Solidity: %s", string(proofJSON))
}
```

## Performance Issues

### Issue: Slow Insert Operations

**Problem:**
Insert operations are taking too long for large trees.

**Solution:**

1. Use batch operations:

```go
func BenchmarkInsertPerformance(b *testing.B) {
    tree := smt.NewSparseMerkleTree(256, nil)

    // Single inserts (slow)
    b.Run("SingleInserts", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            index := big.NewInt(int64(i))
            leaf := smt.Serialize(big.NewInt(int64(i * 2)))
            tree.Insert(index, leaf)
        }
    })

    // Batch inserts (fast)
    b.Run("BatchInserts", func(b *testing.B) {
        processor := smt.NewBatchProcessor(tree)
        operations := make([]smt.BatchOperation, b.N)

        for i := 0; i < b.N; i++ {
            operations[i] = smt.BatchOperation{
                Type:  "insert",
                Index: big.NewInt(int64(i + 10000)),
                Value: smt.Serialize(big.NewInt(int64(i * 2))),
            }
        }

        processor.ProcessBatch(operations)
    })
}
```

2. Use batch operations for multiple insertions:

```go
// Process multiple operations together
proofs, err := tree.BatchInsert(indices, leaves)
if err != nil {
    return err
}
```

### Issue: High Gas Costs in Solidity

**Problem:**
Solidity operations are consuming too much gas.

**Solution:**

1. Use assembly optimizations:

```solidity
function optimizedBatchInsert(
    uint256[] calldata indices,
    bytes32[] calldata leaves
) external {
    assembly {
        let length := indices.length
        let indicesPtr := add(indices.offset, 0x20)
        let leavesPtr := add(leaves.offset, 0x20)

        for { let i := 0 } lt(i, length) { i := add(i, 1) } {
            let index := calldataload(add(indicesPtr, mul(i, 0x20)))
            let leaf := calldataload(add(leavesPtr, mul(i, 0x20)))

            // Optimized insert logic here
        }
    }
}
```

2. Minimize storage operations:

```solidity
// Cache storage reads
function efficientMultipleOperations() external {
    SparseMerkleTree.SMTStorage storage smtRef = smt;
    bytes32 currentRoot = smtRef.getRoot();

    // Use cached reference instead of repeated storage access
    for (uint256 i = 0; i < 10; i++) {
        smtRef.insert(i, bytes32(i));
    }
}
```

## Memory Issues

### Issue: Go Memory Usage Growing Unbounded

**Problem:**
Memory usage keeps increasing during long-running operations.

**Solution:**

1. Implement periodic cleanup:

```go
type ManagedSMT struct {
    tree          *smt.SparseMerkleTree
    operationCount int64
    cleanupInterval int64
}

func (m *ManagedSMT) Insert(index *big.Int, leaf string) (*smt.UpdateProof, error) {
    proof, err := m.tree.Insert(index, leaf)

    // Periodic cleanup
    if atomic.AddInt64(&m.operationCount, 1) % m.cleanupInterval == 0 {
        runtime.GC()
        debug.FreeOSMemory()
    }

    return proof, err
}
```

2. Use memory monitoring:

```go
func MonitorMemoryUsage(tree *smt.SparseMerkleTree) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)

        log.Printf("Memory usage: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d",
            m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)

        if m.Alloc > 100*1024*1024 { // 100MB threshold
            log.Printf("High memory usage detected, triggering GC")
            runtime.GC()
        }
    }
}
```

### Issue: Solidity Memory Limit Exceeded

**Problem:**

```
Error: Transaction reverted: out of memory
```

**Solution:**

1. Use storage instead of memory for large arrays:

```solidity
// Instead of memory arrays
function memoryIntensive(bytes32[] memory largeArray) external {
    // This can exceed memory limits
}

// Use storage or process in chunks
mapping(uint256 => bytes32) private tempStorage;

function storageEfficient(bytes32[] calldata largeArray) external {
    for (uint256 i = 0; i < largeArray.length; i++) {
        tempStorage[i] = largeArray[i];
        // Process in smaller chunks
        if (i % 100 == 0) {
            _processChunk(i - 99, i);
        }
    }
}
```

2. Optimize array operations:

```solidity
function optimizedArrayProcessing(uint256[] calldata data) external {
    // Process data in place without copying to memory
    assembly {
        let length := data.length
        let dataPtr := data.offset

        for { let i := 0 } lt(i, length) { i := add(i, 1) } {
            let value := calldataload(add(dataPtr, mul(i, 0x20)))
            // Process value directly from calldata
        }
    }
}
```

## Debugging Tips

### Enable Debug Logging

```go
// Go debugging
func enableDebugLogging() {
    log.SetLevel(log.DebugLevel)
    log.SetFormatter(&log.TextFormatter{
        FullTimestamp: true,
    })
}

func debugInsert(tree *smt.SparseMerkleTree, index *big.Int, leaf string) {
    log.Debugf("Inserting at index %s, leaf %s", index.String(), leaf)

    oldRoot := tree.Root()
    proof, err := tree.Insert(index, leaf)
    newRoot := tree.Root()

    if err != nil {
        log.Errorf("Insert failed: %v", err)
    } else {
        log.Debugf("Insert successful: old root %s, new root %s", oldRoot, newRoot)
        log.Debugf("Proof: %+v", proof)
    }
}
```

```solidity
// Solidity debugging with events
contract DebugSMT {
    event DebugInfo(string message, bytes32 value);
    event DebugProof(uint256 index, bytes32 leaf, uint256 enables, bytes32[] siblings);

    function debugInsert(uint256 index, bytes32 leaf) external {
        emit DebugInfo("Before insert", smt.getRoot());

        SparseMerkleTree.UpdateProof memory proof = smt.insert(index, leaf);

        emit DebugInfo("After insert", smt.getRoot());
        emit DebugProof(index, proof.newLeaf, proof.enables, proof.siblings);
    }
}
```

### Testing Utilities

```go
// Go testing utilities
func AssertTreeConsistency(t *testing.T, tree *smt.SparseMerkleTree, operations []Operation) {
    // Verify all operations can be retrieved
    for _, op := range operations {
        proof := tree.Get(op.Index)
        assert.True(t, proof.Exists, "Operation %v should exist", op)

        // Verify proof is valid
        isValid := tree.VerifyProof(proof.Leaf, proof.Index, proof.Enables, proof.Siblings)
        assert.True(t, isValid, "Proof should be valid for operation %v", op)
    }
}

func CompareTreeStates(t *testing.T, tree1, tree2 *smt.SparseMerkleTree) {
    root1 := tree1.Root()
    root2 := tree2.Root()
    assert.Equal(t, root1, root2, "Tree roots should match")
}
```

This comprehensive troubleshooting guide should help users identify and resolve the most common issues they might encounter when using the SMT libraries.
