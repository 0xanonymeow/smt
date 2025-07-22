// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../interfaces/ISparseMerkleTree.sol";
import "./SMTHash.sol";

/// @title SMTProof Library
/// @notice Provides proof verification functions for Sparse Merkle Trees
library SMTProof {
    using SMTHash for bytes32;

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
        if (treeDepth > 256) revert ISparseMerkleTree.InvalidTreeDepth(treeDepth);
        if (treeDepth < 256 && index >= 2 ** treeDepth)
            revert ISparseMerkleTree.OutOfRange(index);

        bytes32 computedRoot = leaf;
        uint256 siblingIndex = 0;

        // Rebuild root from leaf->root (LSB->MSB)
        for (uint256 i = 0; i < treeDepth; i++) {
            uint256 bit = (index >> i) & 1;
            bytes32 sibling;

            // Check if sibling is enabled (non-zero)
            if ((enables >> i) & 1 == 1) {
                if (siblingIndex >= siblings.length) {
                    revert ISparseMerkleTree.InvalidProof(leaf, index, enables, siblings);
                }
                sibling = siblings[siblingIndex];
                siblingIndex++;
            } else {
                sibling = bytes32(0);
            }

            // Compute parent hash
            if (bit == 1) {
                // Current is right child
                computedRoot = sibling.hash(computedRoot);
            } else {
                // Current is left child
                computedRoot = computedRoot.hash(sibling);
            }
        }

        return computedRoot == root;
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
        if (treeDepth > 256) revert ISparseMerkleTree.InvalidTreeDepth(treeDepth);
        if (treeDepth < 256 && index >= 2 ** treeDepth)
            revert ISparseMerkleTree.OutOfRange(index);

        bytes32 computedRoot = leaf;
        uint256 siblingIndex = 0;

        // Rebuild root from leaf->root (LSB->MSB)
        for (uint256 i = 0; i < treeDepth; i++) {
            uint256 bit = (index >> i) & 1;
            bytes32 sibling;

            // Check if sibling is enabled (non-zero)
            if ((enables >> i) & 1 == 1) {
                if (siblingIndex >= siblings.length) {
                    // For memory arrays, we need to create a temporary calldata-compatible error
                    bytes32[] memory tempSiblings = new bytes32[](0);
                    revert ISparseMerkleTree.InvalidProof(leaf, index, enables, tempSiblings);
                }
                sibling = siblings[siblingIndex];
                siblingIndex++;
            } else {
                sibling = bytes32(0);
            }

            // Compute parent hash
            if (bit == 1) {
                // Current is right child
                computedRoot = sibling.hash(computedRoot);
            } else {
                // Current is left child
                computedRoot = computedRoot.hash(sibling);
            }
        }

        return computedRoot == root;
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
        if (treeDepth > 256) revert ISparseMerkleTree.InvalidTreeDepth(treeDepth);
        if (treeDepth < 256 && index >= 2 ** treeDepth)
            revert ISparseMerkleTree.OutOfRange(index);

        bytes32 computedRoot = leaf;

        assembly {
            let siblingsPtr := siblings.offset
            let siblingsLen := siblings.length
            let siblingIndex := 0

            for {
                let i := 0
            } lt(i, treeDepth) {
                i := add(i, 1)
            } {
                let bit := and(shr(i, index), 1)
                let sibling := 0

                // Check if sibling is enabled
                if and(shr(i, enables), 1) {
                    if iszero(lt(siblingIndex, siblingsLen)) {
                        // Revert with InvalidProof
                        mstore(0x00, 0x8e72ea7a) // InvalidProof selector
                        mstore(0x04, leaf)
                        mstore(0x24, index)
                        mstore(0x44, enables)
                        mstore(0x64, 0x80) // siblings offset
                        mstore(0x84, siblingsLen)
                        revert(0x00, 0xa4)
                    }
                    sibling := calldataload(add(siblingsPtr, mul(siblingIndex, 0x20)))
                    siblingIndex := add(siblingIndex, 1)
                }

                // Compute parent hash with zero optimization
                switch bit
                case 1 {
                    // Current is right child: hash(sibling, current)
                    if iszero(and(iszero(sibling), iszero(computedRoot))) {
                        mstore(0x00, sibling)
                        mstore(0x20, computedRoot)
                        computedRoot := keccak256(0x00, 0x40)
                    }
                }
                default {
                    // Current is left child: hash(current, sibling)
                    if iszero(and(iszero(computedRoot), iszero(sibling))) {
                        mstore(0x00, computedRoot)
                        mstore(0x20, sibling)
                        computedRoot := keccak256(0x00, 0x40)
                    }
                }
            }
        }

        return computedRoot;
    }
}