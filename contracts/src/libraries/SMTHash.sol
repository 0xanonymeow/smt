// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @title SMTHash Library
/// @notice Provides hash functions and bit utilities for Sparse Merkle Trees
library SMTHash {
    /// @notice keccak hash, but returns 0 if both inputs are 0
    /// @param left Left value
    /// @param right Right value
    function hash(
        bytes32 left,
        bytes32 right
    ) internal pure returns (bytes32 ret) {
        assembly {
            // Optimized zero check: if both inputs are zero, return zero
            if iszero(and(iszero(left), iszero(right))) {
                // Use scratch space for hashing
                mstore(0x00, left)
                mstore(0x20, right)
                ret := keccak256(0x00, 0x40)
            }
            // ret is already 0 if both inputs are zero
        }
    }

    /// @notice Batch hash function for multiple pairs - gas optimized
    /// @param pairs Array of hash pairs (must be even length)
    /// @return results Array of hash results
    function batchHash(
        bytes32[] memory pairs
    ) internal pure returns (bytes32[] memory results) {
        require(pairs.length % 2 == 0, "Pairs array must have even length");

        uint256 resultCount = pairs.length / 2;
        results = new bytes32[](resultCount);

        assembly {
            let pairsPtr := add(pairs, 0x20)
            let resultsPtr := add(results, 0x20)

            for {
                let i := 0
            } lt(i, resultCount) {
                i := add(i, 1)
            } {
                let left := mload(add(pairsPtr, mul(mul(i, 2), 0x20)))
                let right := mload(add(pairsPtr, mul(add(mul(i, 2), 1), 0x20)))

                let result := 0
                // Optimized zero check
                if iszero(and(iszero(left), iszero(right))) {
                    mstore(0x00, left)
                    mstore(0x20, right)
                    result := keccak256(0x00, 0x40)
                }

                mstore(add(resultsPtr, mul(i, 0x20)), result)
            }
        }
    }

    /// @notice Optimized hash function for three inputs (used in leaf creation)
    /// @param input1 First input
    /// @param input2 Second input
    /// @param input3 Third input
    function hash3(
        bytes32 input1,
        bytes32 input2,
        bytes32 input3
    ) internal pure returns (bytes32 ret) {
        assembly {
            // Use scratch space for hashing
            mstore(0x00, input1)
            mstore(0x20, input2)
            mstore(0x40, input3)
            ret := keccak256(0x00, 0x60)
        }
    }

    /// @notice Gas-optimized bit extraction
    /// @param value Value to extract bit from
    /// @param position Bit position (0-based from right)
    function getBit(
        uint256 value,
        uint256 position
    ) internal pure returns (uint256 bit) {
        assembly {
            bit := and(shr(position, value), 1)
        }
    }

    /// @notice Gas-optimized power of 2 calculation
    /// @param exponent Exponent value
    function pow2(uint256 exponent) internal pure returns (uint256 result) {
        assembly {
            result := shl(exponent, 1)
        }
    }
}