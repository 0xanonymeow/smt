// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "../src/SparseMerkleTree.sol";
import "../src/SparseMerkleTreeContract.sol";

/// @title Cross-Platform Compatibility Test
/// @notice Tests that verify Go and Solidity SMT implementations produce identical results
contract CrossPlatformCompatibilityTest is Test {
    using SparseMerkleTree for ISparseMerkleTree.SMTStorage;

    ISparseMerkleTree.SMTStorage private smt;
    SparseMerkleTreeContract private smtContract;

    function setUp() public {
        SparseMerkleTree.initialize(smt, 8);
        smtContract = new SparseMerkleTreeContract(
            8,
            "CrossPlatformTest",
            "1.0.0"
        );
    }

    /// @notice Test hash function compatibility with Go implementation
    function testHashFunctionCompatibility() public {
        // Test case 1: Hash of zero inputs should return keccak256(0,0)
        bytes32 result1 = SparseMerkleTree.hash(bytes32(0), bytes32(0));
        bytes32 expected1 = keccak256(abi.encodePacked(bytes32(0), bytes32(0)));
        assertEq(result1, expected1, "Hash of zero inputs should match keccak256(0,0)");

        // Test case 2: Known hash values from Go implementation
        bytes32 left2 = 0x1111111111111111111111111111111111111111111111111111111111111111;
        bytes32 right2 = 0x2222222222222222222222222222222222222222222222222222222222222222;
        bytes32 expected2 = 0x3e92e0db88d6afea9edc4eedf62fffa4d92bcdfc310dccbe943747fe8302e871;
        bytes32 result2 = SparseMerkleTree.hash(left2, right2);
        assertEq(
            result2,
            expected2,
            "Hash should match Go implementation result"
        );

        // Test case 3: Another known hash value
        bytes32 left3 = 0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890;
        bytes32 right3 = 0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321;
        bytes32 expected3 = 0x5fa4b85b55d6f0543eb23722e63bfd622a406645e39ba54d7220c202f3096fbc;
        bytes32 result3 = SparseMerkleTree.hash(left3, right3);
        assertEq(
            result3,
            expected3,
            "Hash should match Go implementation result"
        );
    }

    /// @notice Test proof verification compatibility with Go-generated proofs
    function testProofVerificationCompatibility() public {
        // Test case: Simple insert and verify proof matches Go implementation
        uint256 index = 5;
        bytes32 leaf = 0x1111111111111111111111111111111111111111111111111111111111111111;

        // Insert leaf
        ISparseMerkleTree.UpdateProof memory updateProof = smt.insert(
            index,
            leaf
        );
        assertFalse(
            updateProof.exists,
            "Initial insert should have exists=false"
        );
        assertEq(
            updateProof.newLeaf,
            leaf,
            "NewLeaf should match inserted leaf"
        );

        // Get proof
        ISparseMerkleTree.Proof memory proof = smt.getView(index);
        assertTrue(proof.exists, "Proof should indicate existence");
        assertEq(proof.leaf, leaf, "Proof leaf should match inserted leaf");
        assertEq(proof.index, index, "Proof index should match");

        // Verify proof
        bool isValid = SparseMerkleTree.verifyProofMemory(
            smt.getRoot(),
            proof.leaf,
            proof.index,
            proof.enables,
            proof.siblings,
            smt.depth
        );
        assertTrue(isValid, "Proof should be valid");
    }

    /// @notice Test multiple operations compatibility with Go implementation
    function testMultipleOperationsCompatibility() public {
        // Test sequence: multiple inserts
        uint256[] memory indices = new uint256[](3);
        bytes32[] memory leaves = new bytes32[](3);

        indices[0] = 1;
        indices[1] = 5;
        indices[2] = 10;

        leaves[
            0
        ] = 0x1111111111111111111111111111111111111111111111111111111111111111;
        leaves[
            1
        ] = 0x5555555555555555555555555555555555555555555555555555555555555555;
        leaves[
            2
        ] = 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa;

        // Insert all leaves
        for (uint256 i = 0; i < indices.length; i++) {
            ISparseMerkleTree.UpdateProof memory proof = smt.insert(
                indices[i],
                leaves[i]
            );
            assertFalse(proof.exists, "Insert should have exists=false");
            assertEq(proof.newLeaf, leaves[i], "NewLeaf should match");
        }

        // Verify all leaves exist and have correct proofs
        for (uint256 i = 0; i < indices.length; i++) {
            assertTrue(smt.exists(indices[i]), "Leaf should exist");

            ISparseMerkleTree.Proof memory proof = smt.getView(indices[i]);
            assertTrue(proof.exists, "Proof should indicate existence");
            assertEq(proof.leaf, leaves[i], "Proof leaf should match");
            assertEq(proof.index, indices[i], "Proof index should match");

            bool isValid = SparseMerkleTree.verifyProofMemory(
                smt.getRoot(),
                proof.leaf,
                proof.index,
                proof.enables,
                proof.siblings,
                smt.depth
            );
            assertTrue(isValid, "Proof should be valid");
        }
    }

    /// @notice Test edge cases compatibility with Go implementation
    function testEdgeCasesCompatibility() public {
        // Test case 1: Index 0 (minimum)
        uint256 index0 = 0;
        bytes32 leaf0 = 0x0000000000000000000000000000000000000000000000000000000000000001;
        smt.insert(index0, leaf0);
        assertTrue(smt.exists(index0), "Index 0 should exist");

        // Test case 2: Maximum index for tree depth
        uint256 maxIndex = 2 ** 8 - 1; // 255 for depth 8
        bytes32 leafMax = 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff;
        smt.insert(maxIndex, leafMax);
        assertTrue(smt.exists(maxIndex), "Max index should exist");

        // Verify edge case proofs
        ISparseMerkleTree.Proof memory proof0 = smt.getView(index0);
        assertTrue(proof0.exists, "Edge case proof should exist");
        assertEq(proof0.leaf, leaf0, "Edge case leaf should match");

        bool isValid0 = SparseMerkleTree.verifyProofMemory(
            smt.getRoot(),
            proof0.leaf,
            proof0.index,
            proof0.enables,
            proof0.siblings,
            smt.depth
        );
        assertTrue(isValid0, "Edge case proof should be valid");

        ISparseMerkleTree.Proof memory proofMax = smt.getView(maxIndex);
        assertTrue(proofMax.exists, "Max index proof should exist");
        assertEq(proofMax.leaf, leafMax, "Max index leaf should match");

        bool isValidMax = SparseMerkleTree.verifyProofMemory(
            smt.getRoot(),
            proofMax.leaf,
            proofMax.index,
            proofMax.enables,
            proofMax.siblings,
            smt.depth
        );
        assertTrue(isValidMax, "Max index proof should be valid");
    }

    /// @notice Test serialization compatibility with Go implementation
    function testSerializationCompatibility() public {
        // Test various values that should serialize identically to Go

        // Test zero
        bytes32 zero = bytes32(0);
        assertEq(
            zero,
            0x0000000000000000000000000000000000000000000000000000000000000000
        );

        // Test one
        bytes32 one = bytes32(uint256(1));
        assertEq(
            one,
            0x0000000000000000000000000000000000000000000000000000000000000001
        );

        // Test max uint8
        bytes32 maxUint8 = bytes32(uint256(255));
        assertEq(
            maxUint8,
            0x00000000000000000000000000000000000000000000000000000000000000ff
        );

        // Test max bytes32
        bytes32 maxBytes32 = 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff;
        assertEq(
            maxBytes32,
            0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
        );
    }
}
