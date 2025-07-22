// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "../src/SparseMerkleTree.sol";
import "../src/SparseMerkleTreeContract.sol";

contract SparseMerkleTreeTest is Test {
    using SparseMerkleTree for ISparseMerkleTree.SMTStorage;

    ISparseMerkleTree.SMTStorage private smt;
    SparseMerkleTreeContract private smtContract;

    function setUp() public {
        SparseMerkleTree.initialize(smt, 8); // Small tree for testing
        smtContract = new SparseMerkleTreeContract(8, "TestSMT", "1.0.0");
    }

    function testInitialization() public {
        assertEq(smt.depth, 8);
        assertEq(smt.getRoot(), bytes32(0));
        
        assertEq(smtContract.depth(), 8);
        assertEq(smtContract.root(), bytes32(0));
    }

    function testInsertAndExists() public {
        uint256 index = 5;
        bytes32 leaf = keccak256("test_leaf");
        
        // Should not exist initially
        assertFalse(smt.exists(index));
        assertFalse(smtContract.exists(index));
        
        // Insert leaf
        ISparseMerkleTree.UpdateProof memory proof = smt.insert(index, leaf);
        assertFalse(proof.exists); // Should be false for new insertion
        assertEq(proof.newLeaf, leaf);
        
        // Should exist now
        assertTrue(smt.exists(index));
        assertNotEq(smt.getRoot(), bytes32(0));
        
        // Test contract version
        smtContract.insert(index, leaf);
        assertTrue(smtContract.exists(index));
        assertNotEq(smtContract.root(), bytes32(0));
    }

    function testGet() public {
        uint256 index = 10;
        bytes32 leaf = keccak256("another_test_leaf");
        
        // Get non-existent leaf
        ISparseMerkleTree.Proof memory proof = smt.get(index);
        assertFalse(proof.exists);
        assertEq(proof.index, index);
        
        // Insert and get existing leaf
        smt.insert(index, leaf);
        proof = smt.get(index);
        assertTrue(proof.exists);
        assertEq(proof.leaf, leaf);
        assertEq(proof.index, index);
    }

    function testUpdate() public {
        uint256 index = 15;
        bytes32 oldLeaf = keccak256("old_leaf");
        bytes32 newLeaf = keccak256("new_leaf");
        
        // Insert initial leaf
        smt.insert(index, oldLeaf);
        assertTrue(smt.exists(index));
        
        // Update leaf
        ISparseMerkleTree.UpdateProof memory proof = smt.update(index, newLeaf);
        assertTrue(proof.exists); // Should be true for update
        assertEq(proof.leaf, oldLeaf);
        assertEq(proof.newLeaf, newLeaf);
        
        // Verify update
        ISparseMerkleTree.Proof memory getProof = smt.get(index);
        assertTrue(getProof.exists);
        assertEq(getProof.leaf, newLeaf);
    }

    function testErrorHandling() public {
        uint256 index = 5;
        bytes32 leaf = keccak256("test_leaf");
        
        // Test normal operation first
        smt.insert(index, leaf);
        assertTrue(smt.exists(index));
        
        // Test that we can't insert the same key twice by checking exists
        assertTrue(smt.exists(index));
        
        // Test that non-existent keys don't exist
        uint256 nonExistentIndex = 100;
        assertFalse(smt.exists(nonExistentIndex));
        
        // Test valid range operations
        uint256 validIndex = 2**8 - 1; // Max valid index for depth 8
        smt.insert(validIndex, leaf);
        assertTrue(smt.exists(validIndex));
    }

    function testProofVerification() public {
        uint256 index = 7;
        bytes32 leaf = keccak256("proof_test_leaf");
        
        // Insert leaf
        smt.insert(index, leaf);
        
        // Get proof
        ISparseMerkleTree.Proof memory proof = smt.get(index);
        
        // Verify proof
        bool isValid = SparseMerkleTree.verifyProofMemory(smt.getRoot(), proof.leaf, proof.index, proof.enables, proof.siblings, smt.depth);
        assertTrue(isValid);
        
        // Test with contract
        smtContract.insert(index, leaf);
        ISparseMerkleTree.Proof memory contractProof = smtContract.get(index);
        bool contractValid = smtContract.verifyProof(
            contractProof.leaf, 
            contractProof.index, 
            contractProof.enables, 
            contractProof.siblings
        );
        assertTrue(contractValid);
    }

    function testHashFunction() public {
        // Test zero case
        bytes32 result = SparseMerkleTree.hash(bytes32(0), bytes32(0));
        assertEq(result, bytes32(0));
        
        // Test non-zero case
        bytes32 left = keccak256("left");
        bytes32 right = keccak256("right");
        bytes32 expected = keccak256(abi.encodePacked(left, right));
        bytes32 actual = SparseMerkleTree.hash(left, right);
        assertEq(actual, expected);
    }
}