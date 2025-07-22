// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../interfaces/ISparseMerkleTree.sol";
import "./SMTHash.sol";

/// @title SMTCore Library
/// @notice Core operations for Sparse Merkle Trees
library SMTCore {
    using SMTHash for bytes32;
    using SMTHash for uint256;

    /// @notice Initialize SMT storage with specified depth
    /// @param smt SMT storage reference
    /// @param treeDepth Depth of the tree (must be <= 256)
    function initialize(ISparseMerkleTree.SMTStorage storage smt, uint16 treeDepth) internal {
        if (treeDepth > 256) revert ISparseMerkleTree.InvalidTreeDepth(treeDepth);
        smt.depth = treeDepth;
        smt.root = bytes32(0);
    }

    /// @notice Get current root of the tree
    /// @param smt SMT storage reference
    /// @return Root hash of the tree
    function getRoot(ISparseMerkleTree.SMTStorage storage smt) internal view returns (bytes32) {
        return smt.root;
    }

    /// @notice Check if a key exists in the tree
    /// @param smt SMT storage reference
    /// @param index Index to check
    /// @return True if key exists
    function exists(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256 index
    ) internal view returns (bool) {
        if (smt.depth < 256 && index >= SMTHash.pow2(smt.depth))
            revert ISparseMerkleTree.OutOfRange(index);

        bytes32 current = smt.root;

        // Walk down the tree
        for (uint256 i = 0; i < smt.depth; i++) {
            // MSB->LSB bit extraction using optimized function
            uint256 bit = SMTHash.getBit(index, smt.depth - i - 1);

            // Check if current node exists in internal storage
            bytes32[2] storage children = smt.db[current];
            if (children[0] == bytes32(0) && children[1] == bytes32(0)) {
                // No children found, check if this is a leaf
                return smt.leaves[current] != bytes32(0);
            }

            current = children[bit];
            if (current == bytes32(0)) {
                return false;
            }
        }

        // Check if final node is a valid leaf
        return smt.leaves[current] != bytes32(0);
    }

    /// @notice Get proof of membership (or non-membership) for a leaf
    /// @param smt SMT storage reference
    /// @param index Index of leaf
    /// @return Proof structure with membership information
    function get(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256 index
    ) internal returns (ISparseMerkleTree.Proof memory) {
        if (smt.depth < 256 && index >= SMTHash.pow2(smt.depth))
            revert ISparseMerkleTree.OutOfRange(index);

        uint256 enables = 0;
        uint256 siblingCount = 0;
        bytes32 current = smt.root;

        // First pass: count siblings to allocate exact array size
        bytes32 tempCurrent = current;
        for (uint256 i = 0; i < smt.depth; i++) {
            uint256 bit = SMTHash.getBit(index, smt.depth - i - 1);
            uint256 siblingBit = bit ^ 1;

            bytes32[2] storage children = smt.db[tempCurrent];
            bytes32 sibling = children[siblingBit];

            if (sibling != bytes32(0)) {
                siblingCount++;
                enables |= (1 << (smt.depth - i - 1));
            }

            tempCurrent = children[bit];
            if (tempCurrent == bytes32(0)) break;
        }

        // Second pass: collect siblings with exact allocation
        bytes32[] memory siblings = new bytes32[](siblingCount);
        uint256 siblingIndex = siblingCount; // Start from end for prepending
        current = smt.root;

        for (uint256 i = 0; i < smt.depth; i++) {
            uint256 bit = SMTHash.getBit(index, smt.depth - i - 1);
            uint256 siblingBit = bit ^ 1;

            bytes32[2] storage children = smt.db[current];
            bytes32 sibling = children[siblingBit];

            if (sibling != bytes32(0)) {
                siblingIndex--;
                siblings[siblingIndex] = sibling;
            }

            current = children[bit];
            if (current == bytes32(0)) break;
        }

        // Check if current is a valid leaf
        bool leafExists = smt.leaves[current] != bytes32(0);
        bytes32 leafValue = leafExists ? smt.leaves[current] : bytes32(0);

        return ISparseMerkleTree.Proof({
            exists: leafExists,
            leaf: current,
            value: leafValue,
            index: index,
            enables: enables,
            siblings: siblings
        });
    }

    /// @notice Get proof of membership (or non-membership) for a leaf (view-only, no events)
    /// @param smt SMT storage reference
    /// @param index Index of leaf
    /// @return Proof structure with membership information
    function getView(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256 index
    ) internal view returns (ISparseMerkleTree.Proof memory) {
        if (smt.depth < 256 && index >= SMTHash.pow2(smt.depth))
            revert ISparseMerkleTree.OutOfRange(index);

        uint256 enables = 0;
        uint256 siblingCount = 0;
        bytes32 current = smt.root;

        // First pass: count siblings to allocate exact array size
        bytes32 tempCurrent = current;
        for (uint256 i = 0; i < smt.depth; i++) {
            uint256 bit = SMTHash.getBit(index, smt.depth - i - 1);
            uint256 siblingBit = bit ^ 1;

            bytes32[2] storage children = smt.db[tempCurrent];
            bytes32 sibling = children[siblingBit];

            if (sibling != bytes32(0)) {
                siblingCount++;
                enables |= (1 << (smt.depth - i - 1));
            }

            tempCurrent = children[bit];
            if (tempCurrent == bytes32(0)) break;
        }

        // Second pass: collect siblings with exact allocation
        bytes32[] memory siblings = new bytes32[](siblingCount);
        uint256 siblingIndex = siblingCount; // Start from end for prepending
        current = smt.root;

        for (uint256 i = 0; i < smt.depth; i++) {
            uint256 bit = SMTHash.getBit(index, smt.depth - i - 1);
            uint256 siblingBit = bit ^ 1;

            bytes32[2] storage children = smt.db[current];
            bytes32 sibling = children[siblingBit];

            if (sibling != bytes32(0)) {
                siblingIndex--;
                siblings[siblingIndex] = sibling;
            }

            current = children[bit];
            if (current == bytes32(0)) break;
        }

        // Check if current is a valid leaf
        bool leafExists = smt.leaves[current] != bytes32(0);
        bytes32 leafValue = leafExists ? smt.leaves[current] : bytes32(0);

        return
            ISparseMerkleTree.Proof({
                exists: leafExists,
                leaf: current,
                value: leafValue,
                index: index,
                enables: enables,
                siblings: siblings
            });
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
        if (exists(smt, index)) revert ISparseMerkleTree.KeyExists(index);
        return _upsert(smt, index, leaf);
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
        if (!exists(smt, index)) revert ISparseMerkleTree.KeyNotFound(index);
        return _upsert(smt, index, newLeaf);
    }

    /// @notice Internal upsert function that handles both insert and update
    /// @param smt SMT storage reference
    /// @param index Index to upsert
    /// @param newLeaf New leaf hash
    /// @return UpdateProof with old and new leaf information
    function _upsert(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256 index,
        bytes32 newLeaf
    ) private returns (ISparseMerkleTree.UpdateProof memory) {
        if (smt.depth < 256 && index >= 2 ** smt.depth)
            revert ISparseMerkleTree.OutOfRange(index);

        // Get proof of current state
        ISparseMerkleTree.Proof memory oldProof = get(smt, index);

        // Collect siblings while walking root->leaf and delete old path
        bytes32[] memory siblings = new bytes32[](smt.depth);
        bytes32 current = smt.root;

        for (uint256 i = 0; i < smt.depth; i++) {
            // MSB->LSB bit extraction
            uint256 bit = (index >> (smt.depth - i - 1)) & 1;
            uint256 siblingBit = bit ^ 1;

            bytes32[2] storage children = smt.db[current];
            siblings[i] = children[siblingBit]; // Store sibling for reconstruction

            bytes32 nextCurrent = children[bit];

            // Delete current internal node
            delete smt.db[current];

            current = nextCurrent;
        }

        // Delete old leaf if it existed
        if (oldProof.exists) {
            delete smt.leaves[oldProof.leaf];
            delete smt.leafIndices[oldProof.leaf];
        }

        // Insert new leaf
        smt.leaves[newLeaf] = newLeaf; // Store leaf hash as value
        smt.leafIndices[newLeaf] = index;

        // Rebuild path from leaf->root
        current = newLeaf;
        for (uint256 i = 0; i < smt.depth; i++) {
            // LSB->MSB for parent reconstruction
            uint256 bit = (index >> i) & 1;
            bytes32 sibling = siblings[smt.depth - 1 - i];

            bytes32 parent;
            if (bit == 1) {
                // Current is right child
                parent = sibling.hash(current);
                smt.db[parent][0] = sibling;
                smt.db[parent][1] = current;
            } else {
                // Current is left child
                parent = current.hash(sibling);
                smt.db[parent][0] = current;
                smt.db[parent][1] = sibling;
            }

            current = parent;
        }

        // Store old root for event emission
        bytes32 oldRoot = smt.root;

        // Update root
        smt.root = current;

        ISparseMerkleTree.UpdateProof memory updateProof = ISparseMerkleTree.UpdateProof({
            exists: oldProof.exists,
            leaf: oldProof.leaf,
            value: oldProof.value,
            index: oldProof.index,
            enables: oldProof.enables,
            siblings: oldProof.siblings,
            newLeaf: newLeaf
        });


        return updateProof;
    }
}