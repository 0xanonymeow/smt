// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../interfaces/ISparseMerkleTree.sol";
import "./SMTCore.sol";
import "./SMTProof.sol";

/// @title SMTBatch Library
/// @notice Batch operations for Sparse Merkle Trees
library SMTBatch {
    using SMTCore for ISparseMerkleTree.SMTStorage;

    /// @notice Batch insert multiple leaves efficiently
    /// @param smt SMT storage reference
    /// @param indices Array of indices to insert
    /// @param leaves Array of leaf hashes to insert
    /// @return updateProofs Array of update proofs for each insertion
    function batchInsert(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256[] memory indices,
        bytes32[] memory leaves
    ) internal returns (ISparseMerkleTree.UpdateProof[] memory updateProofs) {
        require(indices.length == leaves.length, "Arrays length mismatch");

        updateProofs = new ISparseMerkleTree.UpdateProof[](indices.length);

        for (uint256 i = 0; i < indices.length; i++) {
            if (!smt.exists(indices[i])) {
                updateProofs[i] = smt.insert(indices[i], leaves[i]);
            }
        }
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
    ) internal returns (ISparseMerkleTree.UpdateProof[] memory updateProofs) {
        require(indices.length == newLeaves.length, "Arrays length mismatch");

        updateProofs = new ISparseMerkleTree.UpdateProof[](indices.length);

        for (uint256 i = 0; i < indices.length; i++) {
            if (smt.exists(indices[i])) {
                updateProofs[i] = smt.update(indices[i], newLeaves[i]);
            }
        }
    }

    /// @notice Batch get proofs for multiple indices efficiently
    /// @param smt SMT storage reference
    /// @param indices Array of indices to get proofs for
    /// @return proofs Array of proofs for each index
    function batchGet(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256[] memory indices
    ) internal view returns (ISparseMerkleTree.Proof[] memory proofs) {
        proofs = new ISparseMerkleTree.Proof[](indices.length);

        for (uint256 i = 0; i < indices.length; i++) {
            proofs[i] = smt.getView(indices[i]);
        }
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
    ) internal view returns (bool[] memory results) {
        require(
            leaves.length == indices.length &&
                indices.length == enablesArray.length &&
                enablesArray.length == siblingsArray.length,
            "Arrays length mismatch"
        );

        results = new bool[](leaves.length);

        for (uint256 i = 0; i < leaves.length; i++) {
            results[i] = SMTProof.verifyProofMemory(
                smt.root,
                leaves[i],
                indices[i],
                enablesArray[i],
                siblingsArray[i],
                smt.depth
            );
        }
    }

    /// @notice Gas-optimized existence check for multiple indices
    /// @param smt SMT storage reference
    /// @param indices Array of indices to check
    /// @return results Array of existence results
    function batchExists(
        ISparseMerkleTree.SMTStorage storage smt,
        uint256[] memory indices
    ) internal view returns (bool[] memory results) {
        results = new bool[](indices.length);

        for (uint256 i = 0; i < indices.length; i++) {
            results[i] = smt.exists(indices[i]);
        }
    }
}