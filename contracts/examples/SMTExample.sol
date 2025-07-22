// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../src/SparseMerkleTreeContract.sol";

/// @title SMT Usage Example
/// @notice Demonstrates how to use the enhanced Sparse Merkle Tree library
contract SMTExample {
    SparseMerkleTreeContract public smt;
    
    event ExampleOperation(string operation, uint256 index, bytes32 leaf, bytes32 root);
    
    constructor() {
        // Create SMT with depth 8 for this example
        smt = new SparseMerkleTreeContract(8, "ExampleSMT", "1.0.0");
    }
    
    /// @notice Example: Insert multiple leaves and verify proofs
    function demonstrateBasicOperations() external {
        // Insert some example data
        uint256 index1 = 5;
        uint256 index2 = 10;
        uint256 index3 = 15;
        
        bytes32 leaf1 = keccak256("Alice's data");
        bytes32 leaf2 = keccak256("Bob's data");
        bytes32 leaf3 = keccak256("Charlie's data");
        
        // Insert leaves
        smt.insert(index1, leaf1);
        emit ExampleOperation("INSERT", index1, leaf1, smt.root());
        
        smt.insert(index2, leaf2);
        emit ExampleOperation("INSERT", index2, leaf2, smt.root());
        
        smt.insert(index3, leaf3);
        emit ExampleOperation("INSERT", index3, leaf3, smt.root());
        
        // Update one leaf
        bytes32 newLeaf2 = keccak256("Bob's updated data");
        smt.update(index2, newLeaf2);
        emit ExampleOperation("UPDATE", index2, newLeaf2, smt.root());
    }
    
    /// @notice Example: Generate and verify proofs
    function demonstrateProofVerification(uint256 index) external view returns (bool) {
        // Get proof for the specified index
        ISparseMerkleTree.Proof memory proof = smt.get(index);
        
        if (!proof.exists) {
            return false; // No proof to verify for non-existent leaf
        }
        
        // Verify the proof
        return smt.verifyProof(proof.leaf, proof.index, proof.enables, proof.siblings);
    }
    
    /// @notice Example: Batch operations
    function demonstrateBatchOperations(
        uint256[] calldata indices,
        bytes32[] calldata leaves
    ) external {
        require(indices.length == leaves.length, "Arrays must have same length");
        
        for (uint256 i = 0; i < indices.length; i++) {
            if (!smt.exists(indices[i])) {
                smt.insert(indices[i], leaves[i]);
                emit ExampleOperation("BATCH_INSERT", indices[i], leaves[i], smt.root());
            } else {
                smt.update(indices[i], leaves[i]);
                emit ExampleOperation("BATCH_UPDATE", indices[i], leaves[i], smt.root());
            }
        }
    }
    
    /// @notice Get tree information
    function getTreeInfo() external view returns (
        bytes32 root,
        uint16 depth,
        uint256 exampleLeafCount
    ) {
        root = smt.root();
        depth = smt.depth();
        
        // Count some example leaves (indices 0-15)
        exampleLeafCount = 0;
        for (uint256 i = 0; i < 16; i++) {
            if (smt.exists(i)) {
                exampleLeafCount++;
            }
        }
    }
    
    /// @notice Verify external proof (useful for cross-platform verification)
    function verifyExternalProof(
        bytes32 leaf,
        uint256 index,
        uint256 enables,
        bytes32[] calldata siblings
    ) external view returns (bool) {
        return smt.verifyProof(leaf, index, enables, siblings);
    }
    
    /// @notice Compute root from external proof data
    function computeRootFromProof(
        bytes32 leaf,
        uint256 index,
        uint256 enables,
        bytes32[] calldata siblings
    ) external view returns (bytes32) {
        return smt.computeRoot(leaf, index, enables, siblings);
    }
}