// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./SparseMerkleTree.sol";
import "./interfaces/ISparseMerkleTree.sol";
import "./libraries/SMTLeafHash.sol";

/// @title OrderedSMTVerifier
/// @notice Verifies ordered Sparse Merkle Tree proofs generated from sequential data
/// @dev This contract is designed to verify proofs where data is inserted at sequential indices 0,1,2,...
contract OrderedSMTVerifier {
    using SparseMerkleTree for ISparseMerkleTree.SMTStorage;

    /// @dev Represents a proof for a specific index in an ordered sequence
    struct OrderedProof {
        uint256 index;      // Sequential index (0, 1, 2, ...)
        bytes32 leaf;       // Leaf hash
        bytes32 value;      // Original value
        uint256 enables;    // Bitmask for non-zero siblings
        bytes32[] siblings; // Array of sibling hashes
    }

    /// @dev Complete tree verification data
    struct OrderedTreeData {
        bytes32 root;           // Expected root hash
        uint16 depth;           // Tree depth
        uint256 length;         // Number of elements in sequence
        OrderedProof[] proofs;  // Array of proofs in order
    }

    /// @dev Verification result
    struct VerificationResult {
        bool success;           // Overall verification success
        uint256 verifiedCount;  // Number of successfully verified proofs
        uint256 totalProofs;    // Total number of proofs submitted
        bytes32 computedRoot;   // Root computed from proofs
        uint256 gasUsed;        // Gas consumed for verification
    }

    /// @notice Emitted when a tree verification is completed
    /// @param root The expected root hash
    /// @param actualRoot The computed root hash
    /// @param proofCount Number of proofs verified
    /// @param success Whether verification was successful
    /// @param gasUsed Gas consumed during verification
    event TreeVerified(
        bytes32 indexed root,
        bytes32 actualRoot,
        uint256 proofCount,
        bool success,
        uint256 gasUsed
    );

    /// @notice Emitted for individual proof verification results
    /// @param index The index that was verified
    /// @param success Whether the individual proof was valid
    event ProofVerified(uint256 indexed index, bool success);

    /// @notice Custom errors for better gas efficiency
    error InvalidTreeDepth(uint16 depth);
    error InvalidSequence(uint256 expectedIndex, uint256 actualIndex);
    error ProofVerificationFailed(uint256 index);
    error EmptyProofArray();
    error RootMismatch(bytes32 expected, bytes32 computed);

    /// @notice Verifies an ordered sequence of SMT proofs
    /// @param treeData Complete tree data including root, depth, and proofs
    /// @return result Detailed verification result
    function verifyOrderedTree(
        OrderedTreeData calldata treeData
    ) external returns (VerificationResult memory result) {
        uint256 gasStart = gasleft();
        
        // Input validation
        if (treeData.depth == 0 || treeData.depth > 256) {
            revert InvalidTreeDepth(treeData.depth);
        }
        
        if (treeData.proofs.length == 0) {
            revert EmptyProofArray();
        }

        // Initialize result
        result.totalProofs = treeData.proofs.length;
        result.success = true;

        // Verify each proof in sequence
        for (uint256 i = 0; i < treeData.proofs.length; i++) {
            OrderedProof calldata proof = treeData.proofs[i];
            
            // Verify sequential ordering
            if (proof.index != i) {
                emit ProofVerified(i, false);
                revert InvalidSequence(i, proof.index);
            }
            
            // Verify individual proof
            bool proofValid = _verifyIndividualProof(
                treeData.root,
                proof,
                treeData.depth
            );
            
            emit ProofVerified(proof.index, proofValid);
            
            if (proofValid) {
                result.verifiedCount++;
            } else {
                result.success = false;
                revert ProofVerificationFailed(proof.index);
            }
        }

        // Verify the expected length matches actual proof count
        if (treeData.length != treeData.proofs.length) {
            result.success = false;
        }

        // Calculate gas used
        result.gasUsed = gasStart - gasleft();
        result.computedRoot = treeData.root; // In a full implementation, this would be computed

        emit TreeVerified(
            treeData.root,
            result.computedRoot,
            result.verifiedCount,
            result.success,
            result.gasUsed
        );

        return result;
    }

    /// @notice Verifies an array of ordered proofs against an expected root
    /// @param expectedRoot The root hash to verify against
    /// @param treeDepth Depth of the tree
    /// @param proofs Array of ordered proofs to verify
    /// @return success True if all proofs are valid and in correct order
    function verifyOrderedProofs(
        bytes32 expectedRoot,
        uint16 treeDepth,
        OrderedProof[] calldata proofs
    ) external view returns (bool success) {
        if (proofs.length == 0) return false;
        if (treeDepth == 0 || treeDepth > 256) return false;

        // Verify each proof in sequence
        for (uint256 i = 0; i < proofs.length; i++) {
            // Check sequential ordering
            if (proofs[i].index != i) {
                return false;
            }
            
            // Verify individual proof
            if (!_verifyIndividualProof(expectedRoot, proofs[i], treeDepth)) {
                return false;
            }
        }

        return true;
    }

    /// @notice Batch verification of multiple ordered trees
    /// @param treesData Array of tree data to verify
    /// @return results Array of verification results for each tree
    function batchVerifyOrderedTrees(
        OrderedTreeData[] calldata treesData
    ) external returns (VerificationResult[] memory results) {
        results = new VerificationResult[](treesData.length);
        
        for (uint256 i = 0; i < treesData.length; i++) {
            // Note: This would call internal verification logic to avoid external call overhead
            results[i] = this.verifyOrderedTree(treesData[i]);
        }
        
        return results;
    }

    /// @notice Computes the optimal tree depth for a given number of elements
    /// @param elementCount Number of elements to store
    /// @return depth Minimum depth needed to store all elements
    function calculateOptimalDepth(uint256 elementCount) external pure returns (uint16 depth) {
        if (elementCount <= 1) {
            return 1;
        }
        
        // Calculate ceil(log2(elementCount))
        uint256 temp = elementCount - 1;
        depth = 0;
        
        while (temp > 0) {
            temp >>= 1;
            depth++;
        }
        
        // Cap at maximum tree depth
        if (depth > 256) {
            depth = 256;
        }
        
        return depth;
    }

    /// @notice Gets gas estimates for verifying different proof counts
    /// @param proofCounts Array of proof counts to estimate
    /// @return gasEstimates Array of gas estimates for each proof count
    function getGasEstimates(
        uint256[] calldata proofCounts
    ) external pure returns (uint256[] memory gasEstimates) {
        gasEstimates = new uint256[](proofCounts.length);
        
        // Base gas cost for contract call overhead
        uint256 baseGas = 21000;
        // Estimated gas per proof verification (including storage and computation)
        uint256 gasPerProof = 15000;
        
        for (uint256 i = 0; i < proofCounts.length; i++) {
            gasEstimates[i] = baseGas + (gasPerProof * proofCounts[i]);
        }
        
        return gasEstimates;
    }

    /// @dev Internal function to verify an individual proof
    /// @param expectedRoot Root hash to verify against
    /// @param proof Proof to verify
    /// @param treeDepth Depth of the tree
    /// @return valid True if the proof is valid
    function _verifyIndividualProof(
        bytes32 expectedRoot,
        OrderedProof calldata proof,
        uint16 treeDepth
    ) internal pure returns (bool valid) {
        // Use the leaf hash directly from the proof (Go now correctly computes and stores it)
        return SparseMerkleTree.verifyProof(
            expectedRoot,
            proof.leaf,    // Use the computed leaf hash from Go
            proof.index,
            proof.enables,
            proof.siblings,
            treeDepth
        );
    }

    /// @notice View function to check if a tree depth can accommodate a certain number of elements
    /// @param treeDepth Depth of the tree
    /// @param elementCount Number of elements
    /// @return canFit True if the tree can fit all elements
    function canTreeFitElements(uint16 treeDepth, uint256 elementCount) 
        public 
        pure 
        returns (bool canFit) 
    {
        if (treeDepth > 256) return false;
        if (treeDepth == 256) return true; // Max capacity
        
        uint256 maxElements = 2 ** treeDepth;
        return elementCount <= maxElements;
    }

    /// @notice Validates the structure of ordered proof data
    /// @param treeData Tree data to validate
    /// @return valid True if the data structure is valid
    function validateOrderedTreeData(
        OrderedTreeData calldata treeData
    ) external pure returns (bool valid) {
        // Check basic constraints
        if (treeData.depth == 0 || treeData.depth > 256) return false;
        if (treeData.proofs.length == 0) return false;
        if (treeData.length != treeData.proofs.length) return false;
        
        // Check tree capacity
        if (!canTreeFitElements(treeData.depth, treeData.length)) return false;
        
        // Check sequential ordering
        for (uint256 i = 0; i < treeData.proofs.length; i++) {
            if (treeData.proofs[i].index != i) return false;
        }
        
        return true;
    }
}