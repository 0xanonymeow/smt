// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "../src/SparseMerkleTreeContract.sol";

contract SparseMerkleTreeContractTest is Test {
    SparseMerkleTreeContract private smtContract;

    event TreeUpdated(
        uint256 indexed index,
        bytes32 indexed oldLeaf,
        bytes32 indexed newLeaf,
        bytes32 newRoot,
        address operator,
        uint256 blockNumber,
        uint256 operationId
    );

    event LeafInserted(
        uint256 indexed index,
        bytes32 indexed leaf,
        bytes32 indexed operator,
        bytes32 newRoot,
        uint256 blockNumber,
        uint256 operationId
    );

    event LeafUpdated(
        uint256 indexed index,
        bytes32 indexed oldLeaf,
        bytes32 indexed newLeaf,
        bytes32 newRoot,
        address operator,
        uint256 blockNumber,
        uint256 operationId
    );

    function setUp() public {
        smtContract = new SparseMerkleTreeContract(8, "TestSMT", "1.0.0"); // Small tree for testing
    }

    function testContractInitialization() public view {
        assertEq(smtContract.depth(), 8);
        assertEq(smtContract.root(), bytes32(0));
    }

    function testContractInsertAndExists() public {
        uint256 index = 5;
        bytes32 leaf = keccak256("test_leaf");

        // Should not exist initially
        assertFalse(smtContract.exists(index));

        // Insert leaf and check events
        vm.expectEmit(true, true, true, false);
        emit LeafInserted(
            index,
            leaf,
            bytes32(uint256(uint160(address(this)))),
            bytes32(0),
            0,
            0
        ); // We don't know the exact values

        vm.expectEmit(true, true, true, false);
        emit TreeUpdated(
            index,
            bytes32(0),
            leaf,
            bytes32(0),
            address(this),
            0,
            0
        ); // We don't know the exact values

        ISparseMerkleTree.UpdateProof memory proof = smtContract.insert(
            index,
            leaf
        );
        assertFalse(proof.exists); // Should be false for new insertion
        assertEq(proof.newLeaf, leaf);

        // Should exist now
        assertTrue(smtContract.exists(index));
        assertNotEq(smtContract.root(), bytes32(0));
    }

    function testContractGet() public {
        uint256 index = 10;
        bytes32 leaf = keccak256("another_test_leaf");

        // Get non-existent leaf
        ISparseMerkleTree.Proof memory proof = smtContract.get(index);
        assertFalse(proof.exists);
        assertEq(proof.index, index);

        // Insert and get existing leaf
        smtContract.insert(index, leaf);
        proof = smtContract.get(index);
        assertTrue(proof.exists);
        assertEq(proof.leaf, leaf);
        assertEq(proof.index, index);
    }

    function testContractUpdate() public {
        uint256 index = 15;
        bytes32 oldLeaf = keccak256("old_leaf");
        bytes32 newLeaf = keccak256("new_leaf");

        // Insert initial leaf
        smtContract.insert(index, oldLeaf);
        assertTrue(smtContract.exists(index));

        // Update leaf and check events
        vm.expectEmit(true, true, true, false);
        emit LeafUpdated(
            index,
            oldLeaf,
            newLeaf,
            bytes32(0),
            address(this),
            0,
            0
        ); // We don't know the exact values

        vm.expectEmit(true, true, true, false);
        emit TreeUpdated(
            index,
            oldLeaf,
            newLeaf,
            bytes32(0),
            address(this),
            0,
            0
        ); // We don't know the exact values

        ISparseMerkleTree.UpdateProof memory proof = smtContract.update(
            index,
            newLeaf
        );
        assertTrue(proof.exists); // Should be true for update
        assertEq(proof.leaf, oldLeaf);
        assertEq(proof.newLeaf, newLeaf);

        // Verify update
        ISparseMerkleTree.Proof memory getProof = smtContract.get(index);
        assertTrue(getProof.exists);
        assertEq(getProof.leaf, newLeaf);
    }

    function testContractProofVerification() public {
        uint256 index = 7;
        bytes32 leaf = keccak256("proof_test_leaf");

        // Insert leaf
        smtContract.insert(index, leaf);

        // Get proof
        ISparseMerkleTree.Proof memory proof = smtContract.get(index);

        // Verify proof
        bool isValid = smtContract.verifyProof(
            proof.leaf,
            proof.index,
            proof.enables,
            proof.siblings
        );
        assertTrue(isValid);

        // Test computeRoot utility
        bytes32 computedRoot = smtContract.computeRoot(
            proof.leaf,
            proof.index,
            proof.enables,
            proof.siblings
        );
        assertEq(computedRoot, smtContract.root());
    }

    function testContractMultipleOperations() public {
        // Test multiple insertions and updates
        uint256[] memory indices = new uint256[](3);
        bytes32[] memory leaves = new bytes32[](3);

        indices[0] = 1;
        indices[1] = 10;
        indices[2] = 100;

        leaves[0] = keccak256("leaf1");
        leaves[1] = keccak256("leaf2");
        leaves[2] = keccak256("leaf3");

        // Insert all leaves
        for (uint256 i = 0; i < 3; i++) {
            smtContract.insert(indices[i], leaves[i]);
            assertTrue(smtContract.exists(indices[i]));
        }

        // Verify all leaves exist and have correct values
        for (uint256 i = 0; i < 3; i++) {
            ISparseMerkleTree.Proof memory proof = smtContract.get(indices[i]);
            assertTrue(proof.exists);
            assertEq(proof.leaf, leaves[i]);
            assertEq(proof.index, indices[i]);

            // Verify proof
            bool isValid = smtContract.verifyProof(
                proof.leaf,
                proof.index,
                proof.enables,
                proof.siblings
            );
            assertTrue(isValid);
        }

        // Update one leaf
        bytes32 newLeaf = keccak256("updated_leaf2");
        smtContract.update(indices[1], newLeaf);

        // Verify the update
        ISparseMerkleTree.Proof memory updatedProof = smtContract.get(
            indices[1]
        );
        assertTrue(updatedProof.exists);
        assertEq(updatedProof.leaf, newLeaf);

        // Verify all proofs are still valid
        for (uint256 i = 0; i < 3; i++) {
            ISparseMerkleTree.Proof memory proof = smtContract.get(indices[i]);
            bool isValid = smtContract.verifyProof(
                proof.leaf,
                proof.index,
                proof.enables,
                proof.siblings
            );
            assertTrue(isValid);
        }
    }

    function testEventEmission() public {
        uint256 index = 42;
        bytes32 leaf = keccak256("event_test_leaf");

        // Test getWithEvents function works (without checking specific events for now)
        ISparseMerkleTree.Proof memory proof = smtContract.getWithEvents(index);
        assertFalse(proof.exists);
        assertEq(proof.index, index);

        // Insert leaf
        smtContract.insert(index, leaf);

        // Test getWithEvents on existing leaf
        proof = smtContract.getWithEvents(index);
        assertTrue(proof.exists);
        assertEq(proof.leaf, leaf);
        assertEq(proof.index, index);

        // Test update
        bytes32 newLeaf = keccak256("updated_event_test_leaf");
        smtContract.update(index, newLeaf);

        // Verify update worked
        proof = smtContract.getWithEvents(index);
        assertTrue(proof.exists);
        assertEq(proof.leaf, newLeaf);
        assertEq(proof.index, index);
    }
}
