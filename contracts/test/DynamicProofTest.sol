// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "forge-std/console.sol";
import "../src/OrderedSMTVerifier.sol";

/// @title DynamicProofTest
/// @notice Tests that dynamically read Go-generated proof data via FFI
contract DynamicProofTest is Test {
    OrderedSMTVerifier private verifier;

    function setUp() public {
        verifier = new OrderedSMTVerifier();
    }

    /// @notice Test that reads Go-generated proof data via FFI
    function testGoGeneratedProofs() public {
        // Use FFI to read the JSON file
        string[] memory inputs = new string[](2);
        inputs[0] = "cat";
        inputs[1] = "test_data.json";
        bytes memory ffiResult = vm.ffi(inputs);
        string memory json = string(ffiResult);

        // Parse the root hash
        bytes32 expectedRoot = vm.parseJsonBytes32(json, ".root");

        // Parse tree depth
        uint16 treeDepth = uint16(vm.parseJsonUint(json, ".depth"));

        console.log("=== Dynamic Go-Generated Proof Test ===");
        console.log("Root from Go:", vm.toString(expectedRoot));
        console.log("Tree Depth:", treeDepth);

        // Parse first 4 proofs for testing
        OrderedSMTVerifier.OrderedProof[]
            memory proofs = new OrderedSMTVerifier.OrderedProof[](4);

        for (uint256 i = 0; i < 4; i++) {
            string memory indexPath = string.concat(
                ".proofs[",
                vm.toString(i),
                "]"
            );

            // Parse proof data
            uint256 index = vm.parseJsonUint(
                json,
                string.concat(indexPath, ".index")
            );
            bytes32 leaf = vm.parseJsonBytes32(
                json,
                string.concat(indexPath, ".leaf")
            );
            bytes32 value = vm.parseJsonBytes32(
                json,
                string.concat(indexPath, ".value")
            );
            uint256 enables = vm.parseJsonUint(
                json,
                string.concat(indexPath, ".enables")
            );

            // Parse siblings array
            bytes32[] memory siblings = vm.parseJsonBytes32Array(
                json,
                string.concat(indexPath, ".siblings")
            );

            proofs[i] = OrderedSMTVerifier.OrderedProof({
                index: index,
                leaf: leaf,
                value: value,
                enables: enables,
                siblings: siblings
            });

            console.log("Proof", i, "- Index:", index);
            console.log("  Leaf:", vm.toString(leaf));
        }

        // Create tree data
        OrderedSMTVerifier.OrderedTreeData memory treeData = OrderedSMTVerifier
            .OrderedTreeData({
                root: expectedRoot,
                depth: treeDepth,
                length: 4,
                proofs: proofs
            });

        // Verify the proofs
        OrderedSMTVerifier.VerificationResult memory result = verifier
            .verifyOrderedTree(treeData);

        // Assertions
        assertTrue(
            result.success,
            "Go-generated proofs should verify successfully"
        );
        assertEq(result.verifiedCount, 4, "Should verify all 4 proofs");
        assertEq(
            result.computedRoot,
            expectedRoot,
            "Computed root should match Go-generated root"
        );

        console.log("=== Verification Results ===");
        console.log("Success:", result.success);
        console.log("Verified Count:", result.verifiedCount);
        console.log("Expected Root: ", vm.toString(expectedRoot));
        console.log("Computed Root: ", vm.toString(result.computedRoot));
        console.log("Gas Used:", result.gasUsed);

        emit log("CROSS-PLATFORM VERIFICATION SUCCESSFUL!");
        emit log("Go-generated proofs verified in Solidity!");
    }
}