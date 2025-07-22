// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @title SMTLeafHash Library
/// @notice Computes leaf hashes compatible with Go SMT implementation
/// @dev Go uses Keccak256(index || value || 1) for leaf hashes
library SMTLeafHash {
    /// @notice Compute leaf hash using Go's formula: Keccak256(index || value || 1)
    /// @param index The index in the tree (as bytes32)
    /// @param value The value to hash
    /// @return The computed leaf hash
    function computeLeafHash(uint256 index, bytes32 value) internal pure returns (bytes32) {
        // Convert index to bytes32 (big-endian, matching Go's approach)
        bytes32 indexBytes = bytes32(index);
        
        // Compute Keccak256(index || value || 1)
        return keccak256(abi.encodePacked(indexBytes, value, uint8(1)));
    }
    
    /// @notice Verify that a leaf hash matches the expected computation
    /// @param leafHash The provided leaf hash
    /// @param index The index in the tree
    /// @param value The original value
    /// @return True if the leaf hash matches the computation
    function verifyLeafHash(bytes32 leafHash, uint256 index, bytes32 value) internal pure returns (bool) {
        return leafHash == computeLeafHash(index, value);
    }
}