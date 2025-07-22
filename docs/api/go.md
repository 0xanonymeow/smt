# Go API Reference

Complete API documentation for the Go SMT library with 100% test coverage, batch operations, and comprehensive examples.

## Table of Contents

- [SparseMerkleTree](#sparsemerkletree)
- [SparseMerkleTreeKV](#sparsemerkletreekv)
- [Data Structures](#data-structures)
- [Utility Functions](#utility-functions)
- [Error Handling](#error-handling)
- [Performance Optimizations](#performance-optimizations)

## SparseMerkleTree

Raw sparse Merkle tree implementation for direct index-based operations.

### Constructor

```go
func NewSparseMerkleTree(depth int, options *SparseMerkleTreeKVOptions) *SparseMerkleTree
```

Creates a new SparseMerkleTree instance.

**Parameters:**

- `depth` (int): Tree depth (1-256)
- `options` (\*SparseMerkleTreeKVOptions): Configuration options (can be nil for defaults)

**Returns:** \*SparseMerkleTree

**Example:**

```go
// Create with default options
tree := smt.NewSparseMerkleTree(256, nil)

// Create with custom options
options := &smt.SparseMerkleTreeKVOptions{
    HashFn: smt.Keccak,
    SerializerFn: smt.Serialize,
    DeserializerFn: smt.Deserialize,
}
tree := smt.NewSparseMerkleTree(256, options)
```

### Methods

#### Insert

```go
func (smt *SparseMerkleTree) Insert(index *big.Int, leaf string) (*UpdateProof, error)
```

Inserts a new leaf at the specified index.

**Parameters:**

- `index` (\*big.Int): Tree index where to insert
- `leaf` (string): Hex-encoded leaf hash to insert

**Returns:**

- \*UpdateProof: Proof of the insertion operation
- error: Error if key already exists or operation fails

**Errors:**

- Returns error if key already exists at index
- Returns error if index is out of range for tree depth

**Example:**

```go
index := big.NewInt(42)
leaf := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

proof, err := tree.Insert(index, leaf)
if err != nil {
    log.Fatal("Insert failed:", err)
}

fmt.Printf("Inserted leaf: %s\n", proof.NewLeaf)
fmt.Printf("Root after insert: %s\n", tree.Root())
```

#### Update

```go
func (smt *SparseMerkleTree) Update(index *big.Int, newLeaf string) (*UpdateProof, error)
```

Updates an existing leaf at the specified index.

**Parameters:**

- `index` (\*big.Int): Tree index to update
- `newLeaf` (string): Hex-encoded new leaf hash

**Returns:**

- \*UpdateProof: Proof of the update operation
- error: Error if key doesn't exist or operation fails

**Errors:**

- Returns error if key doesn't exist at index
- Returns error if index is out of range for tree depth

**Example:**

```go
index := big.NewInt(42)
newLeaf := "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"

proof, err := tree.Update(index, newLeaf)
if err != nil {
    log.Fatal("Update failed:", err)
}

fmt.Printf("Old leaf: %s\n", proof.Leaf)
fmt.Printf("New leaf: %s\n", proof.NewLeaf)
```

#### Get

```go
func (smt *SparseMerkleTree) Get(index *big.Int) *Proof
```

Retrieves a proof of membership or non-membership for the specified index.

**Parameters:**

- `index` (\*big.Int): Tree index to get proof for

**Returns:** \*Proof: Proof structure with membership information

**Example:**

```go
index := big.NewInt(42)
proof := tree.Get(index)

if proof.Exists {
    fmt.Printf("Value exists: %s\n", *proof.Value)
    fmt.Printf("Leaf hash: %s\n", proof.Leaf)
} else {
    fmt.Printf("Value does not exist at index %s\n", index.String())
}

fmt.Printf("Proof siblings: %v\n", proof.Siblings)
fmt.Printf("Enables bitmask: %s\n", proof.Enables.String())
```

#### Exists

```go
func (smt *SparseMerkleTree) Exists(index *big.Int) bool
```

Checks if a key exists at the specified index.

**Parameters:**

- `index` (\*big.Int): Tree index to check

**Returns:** bool: True if key exists, false otherwise

**Example:**

```go
index := big.NewInt(42)
if tree.Exists(index) {
    fmt.Printf("Key exists at index %s\n", index.String())
} else {
    fmt.Printf("Key does not exist at index %s\n", index.String())
}
```

#### VerifyProof

```go
func (smt *SparseMerkleTree) VerifyProof(leaf string, index *big.Int, enables *big.Int, siblings []string) bool
```

Verifies a Merkle proof against the current tree state.

**Parameters:**

- `leaf` (string): Hex-encoded leaf hash to verify
- `index` (\*big.Int): Tree index of the leaf
- `enables` (\*big.Int): Bitmask indicating which siblings are non-zero
- `siblings` ([]string): Array of hex-encoded sibling hashes

**Returns:** bool: True if proof is valid, false otherwise

**Example:**

```go
// Get a proof first
proof := tree.Get(big.NewInt(42))

// Verify the proof
isValid := tree.VerifyProof(proof.Leaf, proof.Index, proof.Enables, proof.Siblings)
if isValid {
    fmt.Println("Proof is valid")
} else {
    fmt.Println("Proof is invalid")
}
```

#### Root

```go
func (smt *SparseMerkleTree) Root() string
```

Returns the current root hash of the tree.

**Returns:** string: Hex-encoded root hash

**Example:**

```go
root := tree.Root()
fmt.Printf("Current root: %s\n", root)
```

#### Hash

```go
func (smt *SparseMerkleTree) Hash(inputs ...*big.Int) *big.Int
```

Applies the configured hash function to the inputs.

**Parameters:**

- `inputs` (...\*big.Int): Variable number of big.Int inputs to hash

**Returns:** \*big.Int: Hash result

**Example:**

```go
left := big.NewInt(123)
right := big.NewInt(456)
hash := tree.Hash(left, right)
fmt.Printf("Hash result: %s\n", hash.String())
```

## SparseMerkleTreeKV

Key-value interface for the sparse Merkle tree with automatic key hashing.

### Constructor

```go
func NewSparseMerkleTreeKV(options *SparseMerkleTreeKVOptions) *SparseMerkleTreeKV
```

Creates a new SparseMerkleTreeKV instance with depth 256.

**Parameters:**

- `options` (\*SparseMerkleTreeKVOptions): Configuration options (can be nil for defaults)

**Returns:** \*SparseMerkleTreeKV

**Example:**

```go
// Create with default options
treeKV := smt.NewSparseMerkleTreeKV(nil)

// Create with custom options
options := &smt.SparseMerkleTreeKVOptions{
    HashFn: smt.Keccak,
    SerializerFn: smt.Serialize,
    DeserializerFn: smt.Deserialize,
}
treeKV := smt.NewSparseMerkleTreeKV(options)
```

### Methods

#### Insert

```go
func (smt *SparseMerkleTreeKV) Insert(key, value string) (*UpdateProof, error)
```

Inserts a new key-value pair.

**Parameters:**

- `key` (string): Hex-encoded key
- `value` (string): Hex-encoded value

**Returns:**

- \*UpdateProof: Proof of the insertion operation
- error: Error if key already exists or operation fails

**Example:**

```go
key := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
value := "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"

proof, err := treeKV.Insert(key, value)
if err != nil {
    log.Fatal("Insert failed:", err)
}

fmt.Printf("Inserted with leaf: %s\n", proof.NewLeaf)
```

#### Update

```go
func (smt *SparseMerkleTreeKV) Update(key, value string) (*UpdateProof, error)
```

Updates an existing key with a new value.

**Parameters:**

- `key` (string): Hex-encoded key to update
- `value` (string): Hex-encoded new value

**Returns:**

- \*UpdateProof: Proof of the update operation
- error: Error if key doesn't exist or operation fails

**Example:**

```go
key := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
newValue := "0x1111111111111111111111111111111111111111111111111111111111111111"

proof, err := treeKV.Update(key, newValue)
if err != nil {
    log.Fatal("Update failed:", err)
}

fmt.Printf("Updated to leaf: %s\n", proof.NewLeaf)
```

#### Get

```go
func (smt *SparseMerkleTreeKV) Get(key string) *Proof
```

Retrieves a proof for the specified key.

**Parameters:**

- `key` (string): Hex-encoded key to get proof for

**Returns:** \*Proof: Proof structure with membership information

**Example:**

```go
key := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
proof := treeKV.Get(key)

if proof.Exists {
    fmt.Printf("Key exists with value: %s\n", *proof.Value)
} else {
    fmt.Printf("Key does not exist\n")
}
```

#### Exists

```go
func (smt *SparseMerkleTreeKV) Exists(key *big.Int) bool
```

Checks if a key exists in the tree.

**Parameters:**

- `key` (\*big.Int): Key to check (as big.Int)

**Returns:** bool: True if key exists, false otherwise

**Example:**

```go
keyBigInt := smt.Deserialize("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
if treeKV.Exists(keyBigInt) {
    fmt.Println("Key exists")
} else {
    fmt.Println("Key does not exist")
}
```

## Data Structures

### Proof

```go
type Proof struct {
    Exists   bool     `json:"exists"`    // Whether entry exists
    Leaf     string   `json:"leaf"`      // Leaf hash
    Value    *string  `json:"value"`     // Leaf value (if exists)
    Index    *big.Int `json:"index"`     // Tree index
    Enables  *big.Int `json:"enables"`   // Sibling enable bitmask
    Siblings []string `json:"siblings"`  // Non-zero sibling hashes
}
```

### UpdateProof

```go
type UpdateProof struct {
    Proof
    NewLeaf string `json:"newLeaf"`  // New leaf hash after operation
}
```

### SparseMerkleTreeKVOptions

```go
type SparseMerkleTreeKVOptions struct {
    ZeroElement    *big.Int                 // Zero element for the tree
    HashFn         HashFunction             // Hash function to use
    DeserializerFn DeserializerFunction     // String to big.Int converter
    SerializerFn   SerializerFunction       // Big.Int to string converter
}
```

### Function Types

```go
type HashFunction func(inputs ...*big.Int) *big.Int
type DeserializerFunction func(input string) *big.Int
type SerializerFunction func(input *big.Int) string
```

## Utility Functions

### Keccak

```go
func Keccak(inputs ...*big.Int) *big.Int
```

Keccak256 hash function with zero optimization.

**Parameters:**

- `inputs` (...\*big.Int): Variable number of inputs to hash

**Returns:** \*big.Int: Hash result (0 if all inputs are 0)

### Serialize

```go
func Serialize(input *big.Int) string
```

Converts big.Int to zero-padded 32-byte hex string.

**Parameters:**

- `input` (\*big.Int): Value to serialize

**Returns:** string: Hex string with "0x" prefix

### Deserialize

```go
func Deserialize(input string) *big.Int
```

Converts hex string to big.Int.

**Parameters:**

- `input` (string): Hex string with "0x" prefix

**Returns:** \*big.Int: Converted value

**Panics:** If input is not valid hex format

### NumToBytes

```go
func NumToBytes(value *big.Int) []byte
```

Converts big.Int to 32-byte array.

### BytesToNum

```go
func BytesToNum(bytes []byte) *big.Int
```

Converts byte array to big.Int.

## Error Handling

### Error Types

```go
type SMTError struct {
    Type    string  // Error type identifier
    Message string  // Human-readable message
    Code    int     // Numeric error code
    Cause   error   // Underlying error
}
```

### Error Constructors

- `NewInvalidTreeDepthError(depth int) *SMTError`
- `NewOutOfRangeError(index *big.Int, maxIndex *big.Int) *SMTError`
- `NewKeyExistsError(index *big.Int) *SMTError`
- `NewKeyNotFoundError(index *big.Int) *SMTError`
- `NewInvalidProofError(leaf string, index *big.Int) *SMTError`
- `NewInvalidHexFormatError(input string) *SMTError`
- `NewMalformedHexError(input string) *SMTError`

### Error Codes

- `1001`: InvalidTreeDepth
- `1002`: OutOfRange
- `1003`: KeyExists
- `1004`: KeyNotFound
- `1005`: InvalidProof
- `1006`: InvalidHexFormat
- `1007`: MalformedHex

## Performance Optimizations

### Memory Pools

The library includes memory pools for reduced allocations:

- `bigIntPool`: Reusable big.Int instances
- `stringSlicePool`: Reusable string slices
- `bigIntSlicePool`: Reusable big.Int slices
- `byteSlicePool`: Reusable byte slices

### Batch Operations

```go
type BatchProcessor struct {
    tree   *SparseMerkleTree
    treeKV *SparseMerkleTreeKV
    mu     sync.RWMutex
}

func NewBatchProcessor(tree *SparseMerkleTree) *BatchProcessor
func NewBatchProcessorKV(treeKV *SparseMerkleTreeKV) *BatchProcessor
func (bp *BatchProcessor) ProcessBatch(operations []BatchOperation) error
```

**Example:**

```go
processor := smt.NewBatchProcessor(tree)

operations := []smt.BatchOperation{
    {Type: "insert", Index: big.NewInt(1), Value: "0x1234..."},
    {Type: "insert", Index: big.NewInt(2), Value: "0x5678..."},
    {Type: "get", Index: big.NewInt(1)},
}

err := processor.ProcessBatch(operations)
if err != nil {
    log.Fatal("Batch processing failed:", err)
}

// Check results
for i, op := range operations {
    if op.Error != nil {
        fmt.Printf("Operation %d failed: %v\n", i, op.Error)
    } else {
        fmt.Printf("Operation %d result: %v\n", i, op.Result)
    }
}
```
