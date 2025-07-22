// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "../src/SparseMerkleTreeContract.sol";

contract AccessControlTest is Test {
    SparseMerkleTreeContract private smt;
    address private owner;
    address private operator;
    address private unauthorized;

    function setUp() public {
        owner = address(this);
        operator = address(0x1);
        unauthorized = address(0x2);

        smt = new SparseMerkleTreeContract(8, "AccessTestSMT", "1.0.0");
    }

    function testOwnershipAndOperators() public {
        // Test initial state
        assertEq(smt.owner(), owner);
        assertTrue(smt.isOperator(owner)); // Owner should be operator by default
        assertFalse(smt.isOperator(operator));
        assertFalse(smt.isOperator(unauthorized));

        // Add operator
        smt.addOperator(operator);
        assertTrue(smt.isOperator(operator));

        // Test unauthorized access
        vm.prank(unauthorized);
        vm.expectRevert();
        smt.insert(1, keccak256("test"));

        // Test operator access
        vm.prank(operator);
        smt.insert(1, keccak256("test"));
        assertTrue(smt.exists(1));
    }

    function testPauseUnpause() public {
        // Add operator first
        smt.addOperator(operator);

        // Test normal operation
        vm.prank(operator);
        smt.insert(1, keccak256("test"));

        // Pause contract
        smt.pause();

        // Test that operations fail when paused
        vm.prank(operator);
        vm.expectRevert();
        smt.insert(2, keccak256("test2"));

        // Unpause contract
        smt.unpause();

        // Test that operations work again
        vm.prank(operator);
        smt.insert(2, keccak256("test2"));
        assertTrue(smt.exists(2));
    }

    function testBatchOperations() public {
        // Add operator
        smt.addOperator(operator);

        uint256[] memory indices = new uint256[](3);
        bytes32[] memory leaves = new bytes32[](3);

        indices[0] = 1;
        indices[1] = 2;
        indices[2] = 3;

        leaves[0] = keccak256("leaf1");
        leaves[1] = keccak256("leaf2");
        leaves[2] = keccak256("leaf3");

        // Test batch insert
        vm.prank(operator);
        ISparseMerkleTree.UpdateProof[] memory proofs = smt.batchInsert(
            indices,
            leaves
        );

        assertEq(proofs.length, 3);
        for (uint256 i = 0; i < 3; i++) {
            assertTrue(smt.exists(indices[i]));
            assertEq(proofs[i].newLeaf, leaves[i]);
        }

        // Test batch update
        bytes32[] memory newLeaves = new bytes32[](3);
        newLeaves[0] = keccak256("newleaf1");
        newLeaves[1] = keccak256("newleaf2");
        newLeaves[2] = keccak256("newleaf3");

        vm.prank(operator);
        ISparseMerkleTree.UpdateProof[] memory updateProofs = smt.batchUpdate(
            indices,
            newLeaves
        );

        assertEq(updateProofs.length, 3);
        for (uint256 i = 0; i < 3; i++) {
            assertTrue(smt.exists(indices[i]));
            assertEq(updateProofs[i].newLeaf, newLeaves[i]);
        }
    }

    function testContractInfo() public {
        (
            string memory contractName,
            string memory contractVersion,
            address contractOwner,
            bool contractPaused,
            uint16 treeDepth,
            bytes32 treeRoot,
            uint256 totalOps,
            uint256 deployBlock
        ) = smt.getContractInfo();

        assertEq(contractName, "AccessTestSMT");
        assertEq(contractVersion, "1.0.0");
        assertEq(contractOwner, owner);
        assertFalse(contractPaused);
        assertEq(treeDepth, 8);
        assertEq(treeRoot, bytes32(0)); // Empty tree
        assertEq(totalOps, 0); // No operations yet
        assertGt(deployBlock, 0); // Should be greater than 0
    }
}
