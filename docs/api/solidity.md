# Solidity API Reference

Complete API documentation for the Solidity SMT library and deployable contract.

## Table of Contents

- [SparseMerkleTree Library](#sparsemerkletree-library)
- [SparseMerkleTreeContract](#sparsemerkletreecontract)
- [Data Structures](#data-structures)
- [Events](#events)
- [Errors](#errors)
- [Gas Optimization](#gas-optimization)

## SparseMerkleTree Library

Core library providing SMT functionality for use in other contracts.

### Usage

```solidity
import "./SparseMerkleTree.sol";

contract MyContract {
    using SparseMerkleTree for SparseMerkleTree.SMTStorage;
    SparseMerkleTree.SMTStorage private smt;

    constructor() {
        smt.initialize(256);
    }
}
```

### Functions

#### initialize

```solidity
function initialize(SMTStorage storage smt, uint16 treeDepth) internal
```

Initializes SMT storage with specified depth.

**Parameters:**

- `smt`: SMT storage reference
- `treeDepth`: Depth of the tree (must be ≤ 256)

**Reverts:**

- `InvalidTreeDepth(treeDepth)` if depth > 256

**Example:**

```solidity
SparseMerkleTree.SMTStorage private smt;

constructor(uint16 depth) {
    smt.initialize(depth);
}
```

#### getRoot

```solidity
function getRoot(SMTStorage storage smt) internal view returns (bytes32)
```

Returns the current root hash of the tree.

**Parameters:**

- `smt`: SMT storage reference

**Returns:** bytes32: Root hash

**Example:**

```solidity
bytes32 currentRoot = smt.getRoot();
```

#### exists

```solidity
function exists(SMTStorage storage smt, uint256 index) internal view returns (bool)
```

Checks if a key exists at the specified index.

**Parameters:**

- `smt`: SMT storage reference
- `index`: Index to check

**Returns:** bool: True if key exists

**Reverts:**

- `OutOfRange(index)` if index exceeds tree capacity

**Example:**

```solidity
bool keyExists = smt.exists(42);
if (keyExists) {
    // Key exists, proceed with operation
}
```

#### get

```solidity
function get(SMTStorage storage smt, uint256 index) internal returns (Proof memory)
```

Gets proof of membership/non-membership for a leaf with event emission.

**Parameters:**

- `smt`: SMT storage reference
- `index`: Index of leaf

**Returns:** Proof: Proof structure with membership information

**Reverts:**

- `OutOfRange(index)` if index exceeds tree capacity

**Events:**

- Emits `ProofGenerated` event

**Example:**

```solidity
SparseMerkleTree.Proof memory proof = smt.get(42);
if (proof.exists) {
    bytes32 value = proof.value;
    // Process existing value
}
```

#### getView

```solidity
function getView(SMTStorage storage smt, uint256 index) internal view returns (Proof memory)
```

Gets proof of membership/non-membership for a leaf (view-only, no events).

**Parameters:**

- `smt`: SMT storage reference
- `index`: Index of leaf

**Returns:** Proof: Proof structure with membership information

**Example:**

```solidity
SparseMerkleTree.Proof memory proof = smt.getView(42);
// Use for read-only operations
```

#### insert

```solidity
function insert(SMTStorage storage smt, uint256 index, bytes32 leaf) internal returns (UpdateProof memory)
```

Inserts a new leaf into the tree.

**Parameters:**

- `smt`: SMT storage reference
- `index`: Index where to insert
- `leaf`: Leaf hash to insert

**Returns:** UpdateProof: Proof of the insertion operation

**Reverts:**

- `KeyExists(index)` if key already exists
- `OutOfRange(index)` if index exceeds tree capacity

**Events:**

- Emits `TreeStateUpdated` event

**Example:**

```solidity
bytes32 leafHash = keccak256(abi.encodePacked(key, value, uint256(1)));
SparseMerkleTree.UpdateProof memory proof = smt.insert(42, leafHash);
```

#### update

```solidity
function update(SMTStorage storage smt, uint256 index, bytes32 newLeaf) internal returns (UpdateProof memory)
```

Updates an existing leaf in the tree.

**Parameters:**

- `smt`: SMT storage reference
- `index`: Index to update
- `newLeaf`: New leaf hash

**Returns:** UpdateProof: Proof of the update operation

**Reverts:**

- `KeyNotFound(index)` if key doesn't exist
- `OutOfRange(index)` if index exceeds tree capacity

**Events:**

- Emits `TreeStateUpdated` event

**Example:**

```solidity
bytes32 newLeafHash = keccak256(abi.encodePacked(key, newValue, uint256(1)));
SparseMerkleTree.UpdateProof memory proof = smt.update(42, newLeafHash);
```

#### verifyProof

```solidity
function verifyProof(
    SMTStorage storage smt,
    bytes32 leaf,
    uint256 index,
    uint256 enables,
    bytes32[] memory siblings
) internal view returns (bool)
```

Verifies a Merkle proof against the current tree state.

**Parameters:**

- `smt`: SMT storage reference
- `leaf`: Leaf hash to verify
- `index`: Index of the leaf
- `enables`: Bitmask indicating which siblings are non-zero
- `siblings`: Array of non-zero sibling hashes

**Returns:** bool: True if proof is valid

**Example:**

```solidity
bool isValid = smt.verifyProof(leafHash, index, enables, siblings);
require(isValid, "Invalid proof");
```

#### computeRoot

```solidity
function computeRoot(
    uint16 treeDepth,
    bytes32 leaf,
    uint256 index,
    uint256 enables,
    bytes32[] calldata siblings
) internal pure returns (bytes32)
```

Computes Merkle root from leaf and proof (utility function).

**Parameters:**

- `treeDepth`: Depth of Merkle tree
- `leaf`: Leaf hash
- `index`: Index of leaf in tree
- `enables`: Bitmask indicating which siblings are non-zero
- `siblings`: Array of non-zero sibling hashes

**Returns:** bytes32: Computed root hash

**Reverts:**

- `InvalidTreeDepth(treeDepth)` if depth > 256
- `OutOfRange(index)` if index exceeds tree capacity
- `InvalidProof(...)` if proof is malformed

**Example:**

```solidity
bytes32 computedRoot = SparseMerkleTree.computeRoot(
    256, leafHash, index, enables, siblings
);
```

### Hash Functions

#### hash

```solidity
function hash(bytes32 left, bytes32 right) internal pure returns (bytes32)
```

Keccak256 hash function with zero optimization.

**Parameters:**

- `left`: Left input
- `right`: Right input

**Returns:** bytes32: Hash result (0 if both inputs are 0)

#### hash3

```solidity
function hash3(bytes32 input1, bytes32 input2, bytes32 input3) internal pure returns (bytes32)
```

Hash function for three inputs (used in leaf creation).

**Parameters:**

- `input1`: First input
- `input2`: Second input
- `input3`: Third input

**Returns:** bytes32: Hash result

### Batch Operations

#### batchInsert

```solidity
function batchInsert(
    SMTStorage storage smt,
    uint256[] memory indices,
    bytes32[] memory leaves
) internal returns (UpdateProof[] memory)
```

Batch insert multiple leaves efficiently.

#### batchUpdate

```solidity
function batchUpdate(
    SMTStorage storage smt,
    uint256[] memory indices,
    bytes32[] memory newLeaves
) internal returns (UpdateProof[] memory)
```

Batch update multiple leaves efficiently.

#### batchGet

```solidity
function batchGet(
    SMTStorage storage smt,
    uint256[] memory indices
) internal view returns (Proof[] memory)
```

Batch get proofs for multiple indices efficiently.

## SparseMerkleTreeContract

Deployable contract with access control and comprehensive functionality.

### Constructor

```solidity
constructor(
    uint16 treeDepth,
    string memory contractName,
    string memory contractVersion
)
```

Initializes the SMT contract.

**Parameters:**

- `treeDepth`: Depth of the tree (must be ≤ 256)
- `contractName`: Name of the contract instance
- `contractVersion`: Version of the contract

**Example:**

```solidity
SparseMerkleTreeContract smt = new SparseMerkleTreeContract(
    256,
    "MyApp SMT",
    "1.0.0"
);
```

### View Functions

#### root

```solidity
function root() external view returns (bytes32)
```

Returns the current root hash.

#### depth

```solidity
function depth() external view returns (uint16)
```

Returns the tree depth.

#### exists

```solidity
function exists(uint256 index) external view returns (bool)
```

Checks if a key exists.

#### get

```solidity
function get(uint256 index) external view returns (SparseMerkleTree.Proof memory)
```

Gets proof without event emission (view-only).

#### getWithEvents

```solidity
function getWithEvents(uint256 index) external returns (SparseMerkleTree.Proof memory)
```

Gets proof with event emission.

#### verifyProof

```solidity
function verifyProof(
    bytes32 leaf,
    uint256 index,
    uint256 enables,
    bytes32[] calldata siblings
) external view returns (bool)
```

Verifies a Merkle proof.

#### computeRoot

```solidity
function computeRoot(
    bytes32 leaf,
    uint256 index,
    uint256 enables,
    bytes32[] calldata siblings
) external view returns (bytes32)
```

Computes root from proof.

### State-Changing Functions

#### insert

```solidity
function insert(uint256 index, bytes32 leaf) external onlyOperator whenNotPaused returns (SparseMerkleTree.UpdateProof memory)
```

Inserts a new leaf (requires operator role).

#### update

```solidity
function update(uint256 index, bytes32 newLeaf) external onlyOperator whenNotPaused returns (SparseMerkleTree.UpdateProof memory)
```

Updates an existing leaf (requires operator role).

#### batchInsert

```solidity
function batchInsert(
    uint256[] calldata indices,
    bytes32[] calldata leaves
) external onlyOperator whenNotPaused returns (SparseMerkleTree.UpdateProof[] memory)
```

Batch insert multiple leaves.

#### batchUpdate

```solidity
function batchUpdate(
    uint256[] calldata indices,
    bytes32[] calldata newLeaves
) external onlyOperator whenNotPaused returns (SparseMerkleTree.UpdateProof[] memory)
```

Batch update multiple leaves.

### Access Control Functions

#### transferOwnership

```solidity
function transferOwnership(address newOwner) external onlyOwner
```

Transfers contract ownership.

#### addOperator

```solidity
function addOperator(address operator) external onlyOwner
```

Adds an operator who can perform tree operations.

#### removeOperator

```solidity
function removeOperator(address operator) external onlyOwner
```

Removes an operator.

#### isOperator

```solidity
function isOperator(address operator) external view returns (bool)
```

Checks if an address is an operator.

### Emergency Functions

#### pause

```solidity
function pause() external onlyOwner
```

Pauses the contract (emergency function).

#### unpause

```solidity
function unpause() external onlyOwner
```

Unpauses the contract.

### Information Functions

#### getContractInfo

```solidity
function getContractInfo() external view returns (
    string memory contractName,
    string memory contractVersion,
    address contractOwner,
    bool contractPaused,
    uint16 treeDepth,
    bytes32 treeRoot,
    uint256 totalOps,
    uint256 deployBlock
)
```

Returns comprehensive contract information.

#### getIndexOperationCount

```solidity
function getIndexOperationCount(uint256 index) external view returns (uint256)
```

Returns operation count for a specific index.

## Data Structures

### SMTStorage

```solidity
struct SMTStorage {
    mapping(bytes32 => bytes32[2]) db;           // Internal node storage
    mapping(bytes32 => bytes32) leaves;          // Leaf storage
    mapping(bytes32 => uint256) leafIndices;     // Leaf index mapping
    bytes32 root;                                // Current root
    uint16 depth;                                // Tree depth
}
```

### Proof

```solidity
struct Proof {
    bool exists;                // Whether entry exists
    bytes32 leaf;              // Leaf hash
    bytes32 value;             // Leaf value (if exists)
    uint256 index;             // Tree index
    uint256 enables;           // Sibling enable bitmask
    bytes32[] siblings;        // Non-zero sibling hashes
}
```

### UpdateProof

```solidity
struct UpdateProof {
    bool exists;               // Whether entry existed before
    bytes32 leaf;             // Old leaf hash
    bytes32 value;            // Old leaf value (if existed)
    uint256 index;            // Tree index
    uint256 enables;          // Sibling enable bitmask
    bytes32[] siblings;       // Non-zero sibling hashes
    bytes32 newLeaf;          // New leaf hash
}
```

## Events

### Library Events

#### ProofGenerated

```solidity
event ProofGenerated(
    uint256 indexed index,
    bool exists,
    bytes32 indexed leaf,
    bytes32 value,
    uint256 enables,
    bytes32[] siblings
);
```

Emitted when a proof is generated.

#### TreeStateUpdated

```solidity
event TreeStateUpdated(
    uint256 indexed index,
    bytes32 indexed oldLeaf,
    bytes32 indexed newLeaf,
    bytes32 oldRoot,
    bytes32 newRoot,
    uint256 enables,
    bytes32[] siblings
);
```

Emitted when tree state is updated.

### Contract Events

#### TreeUpdated

```solidity
event TreeUpdated(
    uint256 indexed index,
    bytes32 indexed oldLeaf,
    bytes32 indexed newLeaf,
    bytes32 newRoot,
    address operator,
    uint256 blockNumber,
    uint256 operationId
);
```

#### LeafInserted

```solidity
event LeafInserted(
    uint256 indexed index,
    bytes32 indexed leaf,
    bytes32 indexed operator,
    bytes32 newRoot,
    uint256 blockNumber,
    uint256 operationId
);
```

#### LeafUpdated

```solidity
event LeafUpdated(
    uint256 indexed index,
    bytes32 indexed oldLeaf,
    bytes32 indexed newLeaf,
    bytes32 newRoot,
    address operator,
    uint256 blockNumber,
    uint256 operationId
);
```

#### Access Control Events

```solidity
event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
event OperatorAdded(address indexed operator, address indexed addedBy);
event OperatorRemoved(address indexed operator, address indexed removedBy);
event ContractPaused(address indexed pausedBy);
event ContractUnpaused(address indexed unpausedBy);
```

## Errors

### Library Errors

```solidity
error InvalidTreeDepth(uint16 treeDepth);
error OutOfRange(uint256 index);
error InvalidProof(bytes32 leaf, uint256 index, uint256 enables, bytes32[] siblings);
error KeyExists(uint256 index);
error KeyNotFound(uint256 index);
```

### Contract Errors

```solidity
error Unauthorized(address caller, string operation);
error ContractIsPaused();
error ZeroAddress();
error SelfTransfer();
error InvalidOperationError(address caller, uint256 index, string reason);
```

## Gas Optimization

### Assembly Optimizations

The library uses inline assembly for critical operations:

- **Hash Function**: Optimized zero-check and memory usage
- **Bit Operations**: Gas-efficient bit extraction and manipulation
- **Root Computation**: Assembly-optimized proof verification

### Batch Operations

Batch functions reduce gas costs for multiple operations:

```solidity
// Instead of multiple single operations
smt.insert(1, leaf1);
smt.insert(2, leaf2);
smt.insert(3, leaf3);

// Use batch operation
uint256[] memory indices = [1, 2, 3];
bytes32[] memory leaves = [leaf1, leaf2, leaf3];
smt.batchInsert(indices, leaves);
```

### Memory Management

- Exact array allocation to minimize memory usage
- Reuse of memory slots where possible
- Optimized struct packing

### Storage Optimization

- Efficient mapping structures
- Minimal storage writes
- Gas-optimized data layouts

## Best Practices

### Error Handling

Always handle errors appropriately:

```solidity
try smt.insert(index, leaf) returns (SparseMerkleTree.UpdateProof memory proof) {
    // Handle successful insertion
    emit LeafInserted(index, leaf);
} catch Error(string memory reason) {
    // Handle known errors
    emit InsertionFailed(index, reason);
} catch {
    // Handle unknown errors
    emit InsertionFailed(index, "Unknown error");
}
```

### Access Control

Use proper access control patterns:

```solidity
modifier onlyAuthorized() {
    require(isOperator(msg.sender) || msg.sender == owner, "Unauthorized");
    _;
}

function secureInsert(uint256 index, bytes32 leaf) external onlyAuthorized {
    smt.insert(index, leaf);
}
```

### Event Emission

Emit events for off-chain indexing:

```solidity
function insertWithTracking(uint256 index, bytes32 leaf) external {
    SparseMerkleTree.UpdateProof memory proof = smt.insert(index, leaf);

    emit DetailedInsertion(
        index,
        leaf,
        proof.newLeaf,
        smt.getRoot(),
        block.timestamp,
        msg.sender
    );
}
```
