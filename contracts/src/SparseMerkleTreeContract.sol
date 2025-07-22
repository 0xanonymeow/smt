// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./SparseMerkleTree.sol";
import "./interfaces/ISparseMerkleTree.sol";

/// @title Deployable Sparse Merkle Tree Contract
/// @notice Production-ready SMT contract with access control and security features
/// @dev Uses OpenZeppelin-style access control patterns for production deployment
contract SparseMerkleTreeContract {
    ISparseMerkleTree.SMTStorage private smt;
    
    // Access control
    address public owner;
    mapping(address => bool) public operators;
    bool public paused;
    
    // Contract metadata
    string public name;
    string public version;
    uint256 public immutable deploymentBlock;
    
    // Operation tracking
    uint256 public totalOperations;
    mapping(uint256 => uint256) public indexOperationCount;

    // Events for state changes and comprehensive indexing
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
    
    event ProofRequested(
        uint256 indexed index,
        bool exists,
        bytes32 indexed leaf,
        address indexed requester,
        uint256 blockNumber
    );
    
    // Access control events
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    event OperatorAdded(address indexed operator, address indexed addedBy);
    event OperatorRemoved(address indexed operator, address indexed removedBy);
    event ContractPaused(address indexed pausedBy);
    event ContractUnpaused(address indexed unpausedBy);
    
    // Security events
    event UnauthorizedAccess(address indexed caller, string operation);
    event InvalidOperation(address indexed caller, uint256 index, string reason);

    // Custom errors for enhanced error handling
    error Unauthorized(address caller, string operation);
    error ContractIsPaused();
    error InvalidTreeDepth(uint16 treeDepth);
    error ZeroAddress();
    error SelfTransfer();
    error InvalidOperationError(address caller, uint256 index, string reason);

    // Modifiers for access control and security
    modifier onlyOwner() {
        if (msg.sender != owner) {
            emit UnauthorizedAccess(msg.sender, "onlyOwner");
            revert Unauthorized(msg.sender, "onlyOwner");
        }
        _;
    }

    modifier onlyOperator() {
        if (!operators[msg.sender] && msg.sender != owner) {
            emit UnauthorizedAccess(msg.sender, "onlyOperator");
            revert Unauthorized(msg.sender, "onlyOperator");
        }
        _;
    }

    modifier whenNotPaused() {
        if (paused) {
            revert ContractIsPaused();
        }
        _;
    }

    modifier validAddress(address addr) {
        if (addr == address(0)) {
            revert ZeroAddress();
        }
        _;
    }

    /// @notice Initialize the SMT contract with specified depth and metadata
    /// @param treeDepth Depth of the tree (must be <= 256)
    /// @param contractName Name of the contract instance
    /// @param contractVersion Version of the contract
    constructor(
        uint16 treeDepth,
        string memory contractName,
        string memory contractVersion
    ) {
        if (treeDepth > 256) revert InvalidTreeDepth(treeDepth);
        
        SparseMerkleTree.initialize(smt, treeDepth);
        owner = msg.sender;
        operators[msg.sender] = true; // Owner is automatically an operator
        name = contractName;
        version = contractVersion;
        deploymentBlock = block.number;
        
        emit OwnershipTransferred(address(0), msg.sender);
        emit OperatorAdded(msg.sender, msg.sender);
    }

    /// @notice Get the current root of the tree
    /// @return Root hash
    function root() external view returns (bytes32) {
        return SparseMerkleTree.getRoot(smt);
    }

    /// @notice Get the depth of the tree
    /// @return Tree depth
    function depth() external view returns (uint16) {
        return smt.depth;
    }

    /// @notice Check if a key exists in the tree
    /// @param index Index to check
    /// @return True if key exists
    function exists(uint256 index) external view returns (bool) {
        return SparseMerkleTree.exists(smt, index);
    }

    /// @notice Get proof of membership (or non-membership) for a leaf (view-only)
    /// @param index Index of leaf
    /// @return Proof structure
    function get(uint256 index) external view returns (ISparseMerkleTree.Proof memory) {
        return SparseMerkleTree.getView(smt, index);
    }

    /// @notice Get proof of membership (or non-membership) for a leaf with event emission
    /// @param index Index of leaf
    /// @return Proof structure
    function getWithEvents(uint256 index) external returns (ISparseMerkleTree.Proof memory) {
        return SparseMerkleTree.get(smt, index);
    }

    /// @notice Insert a new leaf into the tree
    /// @param index Index where to insert
    /// @param leaf Leaf hash to insert
    /// @return UpdateProof with operation details
    function insert(uint256 index, bytes32 leaf) 
        external 
        onlyOperator 
        whenNotPaused 
        returns (ISparseMerkleTree.UpdateProof memory) 
    {
        ISparseMerkleTree.UpdateProof memory proof = SparseMerkleTree.insert(smt, index, leaf);
        
        // Increment operation counters
        totalOperations++;
        indexOperationCount[index]++;
        
        // Emit comprehensive events
        emit LeafInserted(index, leaf, bytes32(uint256(uint160(msg.sender))), SparseMerkleTree.getRoot(smt), block.number, totalOperations);
        emit TreeUpdated(index, proof.leaf, leaf, SparseMerkleTree.getRoot(smt), msg.sender, block.number, totalOperations);
        
        return proof;
    }

    /// @notice Update an existing leaf in the tree
    /// @param index Index to update
    /// @param newLeaf New leaf hash
    /// @return UpdateProof with operation details
    function update(uint256 index, bytes32 newLeaf) 
        external 
        onlyOperator 
        whenNotPaused 
        returns (ISparseMerkleTree.UpdateProof memory) 
    {
        ISparseMerkleTree.UpdateProof memory proof = SparseMerkleTree.update(smt, index, newLeaf);
        
        // Increment operation counters
        totalOperations++;
        indexOperationCount[index]++;
        
        // Emit comprehensive events
        emit LeafUpdated(index, proof.leaf, newLeaf, SparseMerkleTree.getRoot(smt), msg.sender, block.number, totalOperations);
        emit TreeUpdated(index, proof.leaf, newLeaf, SparseMerkleTree.getRoot(smt), msg.sender, block.number, totalOperations);
        
        return proof;
    }

    /// @notice Verify a Merkle proof against the current tree state
    /// @param leaf Leaf hash to verify
    /// @param index Index of the leaf
    /// @param enables Bitmask indicating which siblings are non-zero
    /// @param siblings Array of non-zero sibling hashes
    /// @return True if proof is valid
    function verifyProof(
        bytes32 leaf,
        uint256 index,
        uint256 enables,
        bytes32[] calldata siblings
    ) external view returns (bool) {
        return SparseMerkleTree.verifyProof(SparseMerkleTree.getRoot(smt), leaf, index, enables, siblings, smt.depth);
    }

    /// @notice Compute root hash from leaf and proof (utility function)
    /// @param leaf Leaf hash
    /// @param index Index of leaf
    /// @param enables Bitmask indicating which siblings are non-zero
    /// @param siblings Array of non-zero sibling hashes
    /// @return Computed root hash
    function computeRoot(
        bytes32 leaf,
        uint256 index,
        uint256 enables,
        bytes32[] calldata siblings
    ) external view returns (bytes32) {
        return SparseMerkleTree.computeRoot(smt.depth, leaf, index, enables, siblings);
    }

    // ============ ACCESS CONTROL FUNCTIONS ============

    /// @notice Transfer ownership of the contract
    /// @param newOwner Address of the new owner
    function transferOwnership(address newOwner) external onlyOwner validAddress(newOwner) {
        if (newOwner == owner) revert SelfTransfer();
        
        address previousOwner = owner;
        owner = newOwner;
        
        // Ensure new owner is an operator
        operators[newOwner] = true;
        
        emit OwnershipTransferred(previousOwner, newOwner);
        emit OperatorAdded(newOwner, msg.sender);
    }

    /// @notice Add an operator who can perform tree operations
    /// @param operator Address to add as operator
    function addOperator(address operator) external onlyOwner validAddress(operator) {
        if (!operators[operator]) {
            operators[operator] = true;
            emit OperatorAdded(operator, msg.sender);
        }
    }

    /// @notice Remove an operator
    /// @param operator Address to remove as operator
    function removeOperator(address operator) external onlyOwner validAddress(operator) {
        if (operator == owner) revert Unauthorized(operator, "Cannot remove owner as operator");
        
        if (operators[operator]) {
            operators[operator] = false;
            emit OperatorRemoved(operator, msg.sender);
        }
    }

    /// @notice Check if an address is an operator
    /// @param operator Address to check
    /// @return True if address is an operator
    function isOperator(address operator) external view returns (bool) {
        return operators[operator];
    }

    // ============ EMERGENCY FUNCTIONS ============

    /// @notice Pause the contract (emergency function)
    function pause() external onlyOwner {
        if (!paused) {
            paused = true;
            emit ContractPaused(msg.sender);
        }
    }

    /// @notice Unpause the contract
    function unpause() external onlyOwner {
        if (paused) {
            paused = false;
            emit ContractUnpaused(msg.sender);
        }
    }

    // ============ ENHANCED QUERY FUNCTIONS ============

    /// @notice Get comprehensive contract information
    /// @return contractName Name of the contract
    /// @return contractVersion Version of the contract
    /// @return contractOwner Owner address
    /// @return contractPaused Whether contract is paused
    /// @return treeDepth Depth of the SMT
    /// @return treeRoot Current root hash
    /// @return totalOps Total operations performed
    /// @return deployBlock Block number when deployed
    function getContractInfo() external view returns (
        string memory contractName,
        string memory contractVersion,
        address contractOwner,
        bool contractPaused,
        uint16 treeDepth,
        bytes32 treeRoot,
        uint256 totalOps,
        uint256 deployBlock
    ) {
        return (
            name,
            version,
            owner,
            paused,
            smt.depth,
            SparseMerkleTree.getRoot(smt),
            totalOperations,
            deploymentBlock
        );
    }

    /// @notice Get operation count for a specific index
    /// @param index Index to query
    /// @return Number of operations performed on this index
    function getIndexOperationCount(uint256 index) external view returns (uint256) {
        return indexOperationCount[index];
    }

    /// @notice Enhanced get function with event emission and access tracking
    /// @param index Index of leaf
    /// @return Proof structure
    function getWithTracking(uint256 index) external returns (ISparseMerkleTree.Proof memory) {
        ISparseMerkleTree.Proof memory proof = SparseMerkleTree.get(smt, index);
        
        // Emit tracking event
        emit ProofRequested(index, proof.exists, proof.leaf, msg.sender, block.number);
        
        return proof;
    }

    // ============ BATCH OPERATIONS ============

    /// @notice Batch insert multiple leaves (gas-optimized)
    /// @param indices Array of indices to insert
    /// @param leaves Array of leaf hashes to insert
    /// @return proofs Array of UpdateProof structures
    function batchInsert(
        uint256[] calldata indices,
        bytes32[] calldata leaves
    ) external onlyOperator whenNotPaused returns (ISparseMerkleTree.UpdateProof[] memory proofs) {
        if (indices.length != leaves.length) {
            revert InvalidOperationError(msg.sender, 0, "Array length mismatch");
        }
        
        proofs = new ISparseMerkleTree.UpdateProof[](indices.length);
        
        for (uint256 i = 0; i < indices.length; i++) {
            ISparseMerkleTree.UpdateProof memory proof = SparseMerkleTree.insert(smt, indices[i], leaves[i]);
            proofs[i] = proof;
            
            // Increment operation counters
            totalOperations++;
            indexOperationCount[indices[i]]++;
            
            // Emit events
            emit LeafInserted(indices[i], leaves[i], bytes32(uint256(uint160(msg.sender))), SparseMerkleTree.getRoot(smt), block.number, totalOperations);
            emit TreeUpdated(indices[i], proof.leaf, leaves[i], SparseMerkleTree.getRoot(smt), msg.sender, block.number, totalOperations);
        }
    }

    /// @notice Batch update multiple leaves (gas-optimized)
    /// @param indices Array of indices to update
    /// @param newLeaves Array of new leaf hashes
    /// @return proofs Array of UpdateProof structures
    function batchUpdate(
        uint256[] calldata indices,
        bytes32[] calldata newLeaves
    ) external onlyOperator whenNotPaused returns (ISparseMerkleTree.UpdateProof[] memory proofs) {
        if (indices.length != newLeaves.length) {
            revert InvalidOperationError(msg.sender, 0, "Array length mismatch");
        }
        
        proofs = new ISparseMerkleTree.UpdateProof[](indices.length);
        
        for (uint256 i = 0; i < indices.length; i++) {
            ISparseMerkleTree.UpdateProof memory proof = SparseMerkleTree.update(smt, indices[i], newLeaves[i]);
            proofs[i] = proof;
            
            // Increment operation counters
            totalOperations++;
            indexOperationCount[indices[i]]++;
            
            // Emit events
            emit LeafUpdated(indices[i], proof.leaf, newLeaves[i], SparseMerkleTree.getRoot(smt), msg.sender, block.number, totalOperations);
            emit TreeUpdated(indices[i], proof.leaf, newLeaves[i], SparseMerkleTree.getRoot(smt), msg.sender, block.number, totalOperations);
        }
    }
}