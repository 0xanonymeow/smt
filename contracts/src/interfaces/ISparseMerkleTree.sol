// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @title ISparseMerkleTree Interface
/// @notice Defines types, errors, and events for Sparse Merkle Tree implementation
interface ISparseMerkleTree {
    // Custom errors for enhanced error handling
    error InvalidTreeDepth(uint16 treeDepth);
    error OutOfRange(uint256 index);
    error InvalidProof(
        bytes32 leaf,
        uint256 index,
        uint256 enables,
        bytes32[] siblings
    );
    error KeyExists(uint256 index);
    error KeyNotFound(uint256 index);

    // Events for off-chain indexing and proof data tracking
    event ProofGenerated(
        uint256 indexed index,
        bool exists,
        bytes32 indexed leaf,
        bytes32 value,
        uint256 enables,
        bytes32[] siblings
    );

    event TreeStateUpdated(
        uint256 indexed index,
        bytes32 indexed oldLeaf,
        bytes32 indexed newLeaf,
        bytes32 oldRoot,
        bytes32 newRoot,
        uint256 enables,
        bytes32[] siblings
    );

    /// @notice SMT storage structure for state management
    struct SMTStorage {
        mapping(bytes32 => bytes32[2]) db; // Internal node storage: hash -> [left_child, right_child]
        mapping(bytes32 => bytes32) leaves; // Leaf storage: leaf_hash -> value
        mapping(bytes32 => uint256) leafIndices; // Leaf storage: leaf_hash -> index
        bytes32 root;
        uint16 depth;
    }

    /// @notice Proof structure matching Go/TypeScript interface
    /// 
    /// Field semantics:
    ///   - leaf:     The computed leaf hash (Keccak256(index || value || 1))
    ///   - value:    The original raw value stored at the index
    ///   - index:    The tree index where the value is stored
    ///   - exists:   Whether the leaf exists in the tree
    ///   - enables:  Bitmask indicating which siblings are non-zero
    ///   - siblings: Array of non-zero sibling hashes for proof verification
    struct Proof {
        bool exists;        // Whether entry exists
        bytes32 leaf;       // Computed leaf hash (Keccak256(index || value || 1))
        bytes32 value;      // Raw value stored at the index
        uint256 index;      // Tree index
        uint256 enables;    // Sibling enable bitmask
        bytes32[] siblings; // Non-zero sibling hashes
    }

    /// @notice Enhanced proof structure for update operations
    struct UpdateProof {
        bool exists; // Whether entry exists
        bytes32 leaf; // Old leaf hash
        bytes32 value; // Old leaf value (if exists)
        uint256 index; // Tree index
        uint256 enables; // Sibling enable bitmask
        bytes32[] siblings; // Non-zero sibling hashes
        bytes32 newLeaf; // New leaf hash
    }
}