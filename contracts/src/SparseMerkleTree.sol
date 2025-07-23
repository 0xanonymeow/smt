// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./interfaces/ISparseMerkleTree.sol";
import "./libraries/SMTCore.sol";
import "./libraries/SMTProof.sol";
import "./libraries/SMTBatch.sol";
import "./libraries/SMTHash.sol";

/// @title SparseMerkleTree
/// @notice Main entry point for Sparse Merkle Tree functionality
/// @dev This contract aggregates all SMT libraries and provides a unified interface
library SparseMerkleTree {
    using SMTCore for ISparseMerkleTree.SMTStorage;
    using SMTBatch for ISparseMerkleTree.SMTStorage;
    
    // Re-export types for backward compatibility
    struct SMTStorage {
        mapping(bytes32 => bytes32[2]) db;
        mapping(bytes32 => bytes32) leaves;
        mapping(bytes32 => uint256) leafIndices;
        bytes32 root;
        uint16 depth;
    }
    
    struct Proof {
        bool exists;
        bytes32 leaf;
        bytes32 value;
        uint256 index;
        uint256 enables;
        bytes32[] siblings;
    }
    
    struct UpdateProof {
        bool exists;
        bytes32 leaf;
        bytes32 value;
        uint256 index;
        uint256 enables;
        bytes32[] siblings;
        bytes32 newLeaf;
    }
    
    // Re-export errors for backward compatibility
    error InvalidTreeDepth(uint16 treeDepth);
    error OutOfRange(uint256 index);
    error InvalidProof(bytes32 leaf, uint256 index, uint256 enables, bytes32[] siblings);
    error KeyExists(uint256 index);
    error KeyNotFound(uint256 index);
    
    // Re-export events for backward compatibility
    event ProofGenerated(
        uint256 indexed index,
        bool exists,
        bytes32 leaf,
        bytes32 value,
        uint256 enables,
        bytes32[] siblings
    );
    
    event TreeStateUpdated(
        uint256 indexed index,
        bytes32 oldLeaf,
        bytes32 newLeaf,
        bytes32 oldRoot,
        bytes32 newRoot,
        uint256 enables,
        bytes32[] siblings
    );
    
    /// @notice Initialize SMT storage with specified depth
    /// @param smt SMT storage reference
    /// @param treeDepth Depth of the tree (must be <= 256)
    function initialize(ISparseMerkleTree.SMTStorage storage smt, uint16 treeDepth) internal {
        smt.initialize(treeDepth);
    }
    
    /// @notice Get current root of the tree
    /// @param smt SMT storage reference
    /// @return Root hash of the tree
    function getRoot(ISparseMerkleTree.SMTStorage storage smt) internal view returns (bytes32) {
        return smt.getRoot();
    }
    
    /// @notice Check if a key exists in the tree
    /// @param smt SMT storage reference
    /// @param index Index to check
    /// @return True if key exists
    function exists(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256 index
    ) internal view returns (bool) {
        return smt.exists(index);
    }
    
    /// @notice Get proof of membership (or non-membership) for a leaf
    /// @param smt SMT storage reference
    /// @param index Index of leaf
    /// @return Proof structure with membership information
    function get(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256 index
    ) internal returns (ISparseMerkleTree.Proof memory) {
        return smt.get(index);
    }
    
    /// @notice Get proof of membership (or non-membership) for a leaf (view-only, no events)
    /// @param smt SMT storage reference
    /// @param index Index of leaf
    /// @return Proof structure with membership information
    function getView(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256 index
    ) internal view returns (ISparseMerkleTree.Proof memory) {
        return smt.getView(index);
    }
    
    /// @notice Insert a new leaf into the tree
    /// @param smt SMT storage reference
    /// @param index Index where to insert
    /// @param leaf Leaf hash to insert
    /// @return UpdateProof with old and new leaf information
    function insert(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256 index,
        bytes32 leaf
    ) internal returns (ISparseMerkleTree.UpdateProof memory) {
        return smt.insert(index, leaf);
    }
    
    /// @notice Update an existing leaf in the tree
    /// @param smt SMT storage reference
    /// @param index Index to update
    /// @param newLeaf New leaf hash
    /// @return UpdateProof with old and new leaf information
    function update(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256 index,
        bytes32 newLeaf
    ) internal returns (ISparseMerkleTree.UpdateProof memory) {
        return smt.update(index, newLeaf);
    }
    
    /// @notice Verify a Merkle proof
    /// @param root Root hash to verify against
    /// @param leaf Leaf hash to verify
    /// @param index Index of the leaf
    /// @param enables Bitmask indicating which siblings are non-zero
    /// @param siblings Array of non-zero sibling hashes
    /// @param treeDepth Depth of the tree
    /// @return True if proof is valid
    function verifyProof(
        bytes32 root,
        bytes32 leaf,
        uint256 index,
        uint256 enables,
        bytes32[] calldata siblings,
        uint16 treeDepth
    ) internal pure returns (bool) {
        return SMTProof.verifyProof(root, leaf, index, enables, siblings, treeDepth);
    }
    
    /// @notice Verify a Merkle proof (memory version)
    /// @param root Root hash to verify against
    /// @param leaf Leaf hash to verify
    /// @param index Index of the leaf
    /// @param enables Bitmask indicating which siblings are non-zero
    /// @param siblings Array of non-zero sibling hashes
    /// @param treeDepth Depth of the tree
    /// @return True if proof is valid
    function verifyProofMemory(
        bytes32 root,
        bytes32 leaf,
        uint256 index,
        uint256 enables,
        bytes32[] memory siblings,
        uint16 treeDepth
    ) internal pure returns (bool) {
        return SMTProof.verifyProofMemory(root, leaf, index, enables, siblings, treeDepth);
    }
    
    /// @notice Compute Merkle root from leaf and proof
    /// @param treeDepth Depth of Merkle tree
    /// @param leaf Leaf hash
    /// @param index Index of leaf in tree
    /// @param enables Bitmask indicating which siblings are non-zero
    /// @param siblings Array of non-zero sibling hashes
    /// @return Computed root hash
    function computeRoot(
        uint16 treeDepth,
        bytes32 leaf,
        uint256 index,
        uint256 enables,
        bytes32[] calldata siblings
    ) internal pure returns (bytes32) {
        return SMTProof.computeRoot(treeDepth, leaf, index, enables, siblings);
    }
    
    /// @notice Batch insert multiple leaves efficiently
    /// @param smt SMT storage reference
    /// @param indices Array of indices to insert
    /// @param leaves Array of leaf hashes to insert
    /// @return updateProofs Array of update proofs for each insertion
    function batchInsert(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256[] memory indices,
        bytes32[] memory leaves
    ) internal returns (ISparseMerkleTree.UpdateProof[] memory) {
        return smt.batchInsert(indices, leaves);
    }
    
    /// @notice Batch update multiple leaves efficiently
    /// @param smt SMT storage reference
    /// @param indices Array of indices to update
    /// @param newLeaves Array of new leaf hashes
    /// @return updateProofs Array of update proofs for each update
    function batchUpdate(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256[] memory indices,
        bytes32[] memory newLeaves
    ) internal returns (ISparseMerkleTree.UpdateProof[] memory) {
        return smt.batchUpdate(indices, newLeaves);
    }
    
    /// @notice Batch get proofs for multiple indices efficiently
    /// @param smt SMT storage reference
    /// @param indices Array of indices to get proofs for
    /// @return proofs Array of proofs for each index
    function batchGet(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256[] memory indices
    ) internal view returns (ISparseMerkleTree.Proof[] memory) {
        return smt.batchGet(indices);
    }
    
    /// @notice Batch verify multiple proofs efficiently
    /// @param smt SMT storage reference
    /// @param leaves Array of leaf hashes to verify
    /// @param indices Array of indices
    /// @param enablesArray Array of enable bitmasks
    /// @param siblingsArray Array of sibling arrays
    /// @return results Array of verification results
    function batchVerifyProof(
        ISparseMerkleTree.SMTStorage storage smt,
        bytes32[] memory leaves,
        uint256[] memory indices,
        uint256[] memory enablesArray,
        bytes32[][] memory siblingsArray
    ) internal view returns (bool[] memory) {
        return smt.batchVerifyProof(leaves, indices, enablesArray, siblingsArray);
    }
    
    /// @notice Gas-optimized existence check for multiple indices
    /// @param smt SMT storage reference
    /// @param indices Array of indices to check
    /// @return results Array of existence results
    function batchExists(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256[] memory indices
    ) internal view returns (bool[] memory) {
        return smt.batchExists(indices);
    }
    
    // ============ EXPORT HASH AND UTILITY FUNCTIONS ============
    
    /// @notice keccak256 hash function
    /// @param left Left value
    /// @param right Right value
    function hash(bytes32 left, bytes32 right) internal pure returns (bytes32) {
        return SMTHash.hash(left, right);
    }
    
    /// @notice Batch hash function for multiple pairs
    /// @param pairs Array of hash pairs (must be even length)
    /// @return results Array of hash results
    function batchHash(bytes32[] memory pairs) internal pure returns (bytes32[] memory) {
        return SMTHash.batchHash(pairs);
    }
    
    /// @notice Hash function for three inputs
    /// @param a First value
    /// @param b Second value
    /// @param c Third value
    function hash3(bytes32 a, bytes32 b, bytes32 c) internal pure returns (bytes32) {
        return SMTHash.hash3(a, b, c);
    }
    
    /// @notice Get the bit at position i in a uint256
    /// @param intVal The integer value
    /// @param index The bit position
    /// @return The bit value (0 or 1)
    function getBit(uint256 intVal, uint256 index) internal pure returns (uint256) {
        return SMTHash.getBit(intVal, index);
    }
    
    /// @notice Calculate 2^n
    /// @param n The exponent
    /// @return result 2^n
    function pow2(uint256 n) internal pure returns (uint256) {
        return SMTHash.pow2(n);
    }
}