# Cross-Platform Integration Guide

This guide demonstrates how to use the Go and Solidity SMT libraries together for seamless cross-platform operations.

## Table of Contents

- [Overview](#overview)
- [Architecture Patterns](#architecture-patterns)
- [Proof Compatibility](#proof-compatibility)
- [Hybrid Applications](#hybrid-applications)
- [State Synchronization](#state-synchronization)
- [Best Practices](#best-practices)

## Overview

The SMT libraries are designed for perfect cross-platform compatibility:

- **Identical Hash Functions**: Both libraries produce the same hash outputs
- **Compatible Proof Formats**: Proofs generated in Go verify in Solidity and vice versa
- **Consistent Serialization**: Data formats are identical across platforms
- **Synchronized State**: Trees with the same operations produce identical roots

## Architecture Patterns

### Pattern 1: Off-Chain Computation, On-Chain Verification

Use Go for heavy computation off-chain, then verify results on-chain.

```
┌─────────────────┐    ┌─────────────────┐
│   Go Backend    │    │ Solidity Contract│
│                 │    │                 │
│ Heavy SMT Ops   │───►│ Proof Verification│
│ Batch Processing│    │ State Updates   │
│ Complex Logic   │    │ Access Control  │
└─────────────────┘    └─────────────────┘
```

**Go Backend:**

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "math/big"
    "github.com/0xanonymeow/smt"
)

type ProofData struct {
    Leaf     string   `json:"leaf"`
    Index    string   `json:"index"`
    Enables  string   `json:"enables"`
    Siblings []string `json:"siblings"`
    Root     string   `json:"root"`
}

func processUserData(userData []UserRecord) (*ProofData, error) {
    // Create SMT and process all user data
    tree := smt.NewSparseMerkleTree(256, nil)

    for _, user := range userData {
        index := big.NewInt(int64(user.ID))
        leaf := smt.Serialize(hashUserData(user))

        _, err := tree.Insert(index, leaf)
        if err != nil {
            return nil, fmt.Errorf("failed to insert user %d: %w", user.ID, err)
        }
    }

    // Generate proof for specific user
    targetIndex := big.NewInt(int64(userData[0].ID))
    proof := tree.Get(targetIndex)

    return &ProofData{
        Leaf:     proof.Leaf,
        Index:    proof.Index.String(),
        Enables:  proof.Enables.String(),
        Siblings: proof.Siblings,
        Root:     tree.Root(),
    }, nil
}

func hashUserData(user UserRecord) *big.Int {
    // Hash user data consistently
    data := fmt.Sprintf("%d:%s:%s", user.ID, user.Name, user.Email)
    hash := smt.Keccak(smt.Deserialize("0x"+hex.EncodeToString([]byte(data))))
    return hash
}
```

**Solidity Contract:**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./SparseMerkleTree.sol";

contract UserRegistry {
    using SparseMerkleTree for SparseMerkleTree.SMTStorage;

    SparseMerkleTree.SMTStorage private userSMT;
    bytes32 public currentRoot;

    event RootUpdated(bytes32 indexed oldRoot, bytes32 indexed newRoot);
    event UserVerified(uint256 indexed userId, bytes32 indexed userHash);

    constructor() {
        userSMT.initialize(256);
    }

    function updateRoot(bytes32 newRoot) external {
        bytes32 oldRoot = currentRoot;
        currentRoot = newRoot;
        emit RootUpdated(oldRoot, newRoot);
    }

    function verifyUser(
        uint256 userId,
        bytes32 userHash,
        uint256 enables,
        bytes32[] calldata siblings
    ) external view returns (bool) {
        return SparseMerkleTree._verifyProofAgainstRoot(
            currentRoot,
            256,
            userHash,
            userId,
            enables,
            siblings
        );
    }

    function verifyAndReward(
        uint256 userId,
        bytes32 userHash,
        uint256 enables,
        bytes32[] calldata siblings
    ) external {
        require(
            verifyUser(userId, userHash, enables, siblings),
            "Invalid user proof"
        );

        emit UserVerified(userId, userHash);
        // Distribute rewards, update balances, etc.
    }
}
```

### Pattern 2: Hybrid State Management

Maintain synchronized state across both platforms.

```go
// Go: State Manager
type StateManager struct {
    tree     *smt.SparseMerkleTree
    contract *ethclient.Client
    address  common.Address
}

func (sm *StateManager) SyncInsert(index *big.Int, leaf string) error {
    // Insert in Go tree
    proof, err := sm.tree.Insert(index, leaf)
    if err != nil {
        return err
    }

    // Insert in Solidity contract
    tx, err := sm.callContract("insert", index, common.HexToHash(leaf))
    if err != nil {
        // Rollback Go operation if needed
        return err
    }

    // Verify both trees have same root
    goRoot := sm.tree.Root()
    solidityRoot, err := sm.getContractRoot()
    if err != nil {
        return err
    }

    if goRoot != solidityRoot.Hex() {
        return fmt.Errorf("root mismatch: Go=%s, Solidity=%s", goRoot, solidityRoot.Hex())
    }

    return nil
}
```

### Pattern 3: Proof Relay System

Generate proofs in one system and verify in another.

```go
// Go: Proof Generator Service
type ProofService struct {
    tree *smt.SparseMerkleTree
}

func (ps *ProofService) GenerateProof(index *big.Int) (*smt.Proof, error) {
    return ps.tree.Get(index), nil
}

func (ps *ProofService) SerializeProof(proof *smt.Proof) ([]byte, error) {
    return json.Marshal(map[string]interface{}{
        "exists":   proof.Exists,
        "leaf":     proof.Leaf,
        "value":    proof.Value,
        "index":    proof.Index.String(),
        "enables":  proof.Enables.String(),
        "siblings": proof.Siblings,
    })
}
```

```solidity
// Solidity: Proof Verifier
contract ProofVerifier {
    using SparseMerkleTree for SparseMerkleTree.SMTStorage;

    struct ExternalProof {
        bytes32 leaf;
        uint256 index;
        uint256 enables;
        bytes32[] siblings;
    }

    function verifyExternalProof(
        ExternalProof calldata proof,
        bytes32 expectedRoot
    ) external pure returns (bool) {
        return SparseMerkleTree._verifyProofAgainstRoot(
            expectedRoot,
            256,
            proof.leaf,
            proof.index,
            proof.enables,
            proof.siblings
        );
    }
}
```

## Proof Compatibility

### Data Format Conversion

Both libraries use identical data formats, but you may need to convert between language-specific types:

```go
// Go: Convert proof to JSON for API
func ProofToJSON(proof *smt.Proof) ([]byte, error) {
    data := map[string]interface{}{
        "exists":   proof.Exists,
        "leaf":     proof.Leaf,
        "value":    proof.Value,
        "index":    proof.Index.String(),
        "enables":  proof.Enables.String(),
        "siblings": proof.Siblings,
    }
    return json.Marshal(data)
}

// Go: Convert JSON to proof for verification
func JSONToProof(data []byte) (*smt.Proof, error) {
    var raw map[string]interface{}
    if err := json.Unmarshal(data, &raw); err != nil {
        return nil, err
    }

    index, _ := new(big.Int).SetString(raw["index"].(string), 10)
    enables, _ := new(big.Int).SetString(raw["enables"].(string), 10)

    siblings := make([]string, len(raw["siblings"].([]interface{})))
    for i, s := range raw["siblings"].([]interface{}) {
        siblings[i] = s.(string)
    }

    var value *string
    if raw["value"] != nil {
        v := raw["value"].(string)
        value = &v
    }

    return &smt.Proof{
        Exists:   raw["exists"].(bool),
        Leaf:     raw["leaf"].(string),
        Value:    value,
        Index:    index,
        Enables:  enables,
        Siblings: siblings,
    }, nil
}
```

### JavaScript/TypeScript Integration

For web applications, you can use JavaScript to bridge Go and Solidity:

```javascript
// Convert Go proof format to Solidity parameters
function convertProofForSolidity(goProof) {
  return {
    leaf: goProof.leaf,
    index: BigInt(goProof.index),
    enables: BigInt(goProof.enables),
    siblings: goProof.siblings,
  };
}

// Verify proof in smart contract
async function verifyProofOnChain(contract, goProof, expectedRoot) {
  const solidityProof = convertProofForSolidity(goProof);

  const isValid = await contract.verifyProof(
    solidityProof.leaf,
    solidityProof.index,
    solidityProof.enables,
    solidityProof.siblings
  );

  return isValid;
}

// Example usage with ethers.js
const goProofJSON = await fetch("/api/proof/42").then((r) => r.json());
const isValid = await verifyProofOnChain(contract, goProofJSON, expectedRoot);
```

## Hybrid Applications

### Example: Decentralized Identity System

**Architecture:**

- Go backend manages identity data and generates proofs
- Solidity contract verifies identities and manages permissions
- Web frontend provides user interface

**Go Backend:**

```go
type IdentityManager struct {
    tree *smt.SparseMerkleTree
    db   *sql.DB
}

func (im *IdentityManager) RegisterIdentity(userID uint64, identity Identity) (*smt.Proof, error) {
    // Hash identity data
    identityHash := im.hashIdentity(identity)
    leaf := smt.Serialize(identityHash)

    // Insert into SMT
    index := big.NewInt(int64(userID))
    proof, err := im.tree.Insert(index, leaf)
    if err != nil {
        return nil, err
    }

    // Store in database
    err = im.storeIdentity(userID, identity, proof)
    if err != nil {
        return nil, err
    }

    return &proof.Proof, nil
}

func (im *IdentityManager) GetIdentityProof(userID uint64) (*smt.Proof, error) {
    index := big.NewInt(int64(userID))
    proof := im.tree.Get(index)

    if !proof.Exists {
        return nil, fmt.Errorf("identity not found for user %d", userID)
    }

    return proof, nil
}

func (im *IdentityManager) hashIdentity(identity Identity) *big.Int {
    data := fmt.Sprintf("%s:%s:%d", identity.Name, identity.Email, identity.Timestamp)
    return smt.Keccak(smt.Deserialize("0x" + hex.EncodeToString([]byte(data))))
}
```

**Solidity Contract:**

```solidity
contract IdentityVerifier {
    using SparseMerkleTree for SparseMerkleTree.SMTStorage;

    SparseMerkleTree.SMTStorage private identitySMT;
    mapping(address => uint256) public userIds;
    mapping(uint256 => bool) public verifiedIdentities;

    event IdentityVerified(address indexed user, uint256 indexed userId);
    event AccessGranted(address indexed user, string resource);

    constructor() {
        identitySMT.initialize(256);
    }

    function verifyIdentity(
        uint256 userId,
        bytes32 identityHash,
        uint256 enables,
        bytes32[] calldata siblings,
        bytes32 expectedRoot
    ) external {
        // Verify proof against expected root
        bool isValid = SparseMerkleTree._verifyProofAgainstRoot(
            expectedRoot,
            256,
            identityHash,
            userId,
            enables,
            siblings
        );

        require(isValid, "Invalid identity proof");

        // Mark identity as verified
        userIds[msg.sender] = userId;
        verifiedIdentities[userId] = true;

        emit IdentityVerified(msg.sender, userId);
    }

    function accessResource(string calldata resource) external {
        uint256 userId = userIds[msg.sender];
        require(userId != 0, "User not registered");
        require(verifiedIdentities[userId], "Identity not verified");

        emit AccessGranted(msg.sender, resource);
        // Grant access to resource
    }
}
```

### Example: Supply Chain Tracking

**Go Backend (Logistics System):**

```go
type SupplyChainTracker struct {
    tree *smt.SparseMerkleTree
}

func (sct *SupplyChainTracker) TrackItem(itemID uint64, location string, timestamp int64) error {
    // Create tracking record
    record := TrackingRecord{
        ItemID:    itemID,
        Location:  location,
        Timestamp: timestamp,
    }

    // Hash the record
    recordHash := sct.hashRecord(record)
    leaf := smt.Serialize(recordHash)

    // Insert into SMT
    index := big.NewInt(int64(itemID))
    _, err := sct.tree.Insert(index, leaf)
    return err
}

func (sct *SupplyChainTracker) GenerateTrackingProof(itemID uint64) (*smt.Proof, error) {
    index := big.NewInt(int64(itemID))
    proof := sct.tree.Get(index)

    if !proof.Exists {
        return nil, fmt.Errorf("item %d not found", itemID)
    }

    return proof, nil
}
```

**Solidity Contract (Verification System):**

```solidity
contract SupplyChainVerifier {
    struct TrackingProof {
        uint256 itemId;
        bytes32 recordHash;
        uint256 enables;
        bytes32[] siblings;
    }

    mapping(uint256 => bytes32) public verifiedItems;

    event ItemVerified(uint256 indexed itemId, bytes32 recordHash);

    function verifyTracking(
        TrackingProof calldata proof,
        bytes32 expectedRoot
    ) external {
        bool isValid = SparseMerkleTree._verifyProofAgainstRoot(
            expectedRoot,
            256,
            proof.recordHash,
            proof.itemId,
            proof.enables,
            proof.siblings
        );

        require(isValid, "Invalid tracking proof");

        verifiedItems[proof.itemId] = proof.recordHash;
        emit ItemVerified(proof.itemId, proof.recordHash);
    }

    function isItemVerified(uint256 itemId) external view returns (bool) {
        return verifiedItems[itemId] != bytes32(0);
    }
}
```

## State Synchronization

### Synchronization Strategies

#### 1. Event-Driven Synchronization

```go
// Go: Listen to Solidity events and sync
func (sm *StateManager) SyncFromContract() error {
    // Listen to contract events
    logs := make(chan types.Log)
    sub, err := sm.client.SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{
        Addresses: []common.Address{sm.contractAddress},
    }, logs)
    if err != nil {
        return err
    }

    for {
        select {
        case err := <-sub.Err():
            return err
        case vLog := <-logs:
            // Parse event and sync to Go tree
            err := sm.handleContractEvent(vLog)
            if err != nil {
                log.Printf("Error handling event: %v", err)
            }
        }
    }
}

func (sm *StateManager) handleContractEvent(vLog types.Log) error {
    // Parse event data
    event, err := sm.parseEvent(vLog)
    if err != nil {
        return err
    }

    // Apply to Go tree
    switch event.Type {
    case "Insert":
        _, err = sm.tree.Insert(event.Index, event.Leaf)
    case "Update":
        _, err = sm.tree.Update(event.Index, event.Leaf)
    }

    return err
}
```

#### 2. Periodic Root Verification

```go
func (sm *StateManager) VerifySync() error {
    goRoot := sm.tree.Root()
    contractRoot, err := sm.getContractRoot()
    if err != nil {
        return err
    }

    if goRoot != contractRoot.Hex() {
        return fmt.Errorf("sync error: Go root %s != Contract root %s",
            goRoot, contractRoot.Hex())
    }

    return nil
}

// Run periodic verification
func (sm *StateManager) StartSyncVerification(interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            if err := sm.VerifySync(); err != nil {
                log.Printf("Sync verification failed: %v", err)
                // Trigger resync process
                sm.Resync()
            }
        }
    }()
}
```

#### 3. Checkpoint-Based Synchronization

```go
type Checkpoint struct {
    BlockNumber uint64    `json:"block_number"`
    Root        string    `json:"root"`
    Timestamp   time.Time `json:"timestamp"`
}

func (sm *StateManager) CreateCheckpoint() (*Checkpoint, error) {
    blockNumber, err := sm.client.BlockNumber(context.Background())
    if err != nil {
        return nil, err
    }

    return &Checkpoint{
        BlockNumber: blockNumber,
        Root:        sm.tree.Root(),
        Timestamp:   time.Now(),
    }, nil
}

func (sm *StateManager) RestoreFromCheckpoint(checkpoint *Checkpoint) error {
    // Rebuild tree from checkpoint
    // This would involve replaying events from the checkpoint block
    return sm.replayFromBlock(checkpoint.BlockNumber)
}
```

## Best Practices

### 1. Consistent Hashing

Always use the same hash function and input format across platforms:

```go
// Go
func HashUserData(userID uint64, name, email string) *big.Int {
    data := fmt.Sprintf("%d:%s:%s", userID, name, email)
    return smt.Keccak(smt.Deserialize("0x" + hex.EncodeToString([]byte(data))))
}
```

```solidity
// Solidity
function hashUserData(uint256 userId, string memory name, string memory email)
    internal pure returns (bytes32) {
    return keccak256(abi.encodePacked(userId, ":", name, ":", email));
}
```

### 2. Error Handling

Handle cross-platform errors gracefully:

```go
func (sm *StateManager) SafeInsert(index *big.Int, leaf string) error {
    // Try Go insertion first
    goProof, err := sm.tree.Insert(index, leaf)
    if err != nil {
        return fmt.Errorf("Go insertion failed: %w", err)
    }

    // Try Solidity insertion
    tx, err := sm.insertInContract(index, leaf)
    if err != nil {
        // Rollback Go insertion if possible
        // (This is complex and depends on your specific needs)
        return fmt.Errorf("Solidity insertion failed: %w", err)
    }

    // Verify consistency
    return sm.VerifySync()
}
```

### 3. Gas Optimization

Use batch operations when possible:

```solidity
function batchVerifyAndProcess(
    uint256[] calldata indices,
    bytes32[] calldata leaves,
    uint256[] calldata enables,
    bytes32[][] calldata siblings
) external {
    require(indices.length == leaves.length, "Length mismatch");

    for (uint256 i = 0; i < indices.length; i++) {
        bool isValid = SparseMerkleTree._verifyProofAgainstRoot(
            currentRoot,
            256,
            leaves[i],
            indices[i],
            enables[i],
            siblings[i]
        );

        if (isValid) {
            processVerifiedItem(indices[i], leaves[i]);
        }
    }
}
```

### 4. Testing Cross-Platform Compatibility

```go
func TestCrossPlatformCompatibility(t *testing.T) {
    // Create identical trees in Go and simulate Solidity
    goTree := smt.NewSparseMerkleTree(256, nil)

    // Insert same data
    index := big.NewInt(42)
    leaf := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

    goProof, err := goTree.Insert(index, leaf)
    require.NoError(t, err)

    // Verify proof would work in Solidity
    isValid := goTree.VerifyProof(goProof.Leaf, goProof.Index, goProof.Enables, goProof.Siblings)
    require.True(t, isValid)

    // Test serialization compatibility
    proofJSON, err := json.Marshal(goProof)
    require.NoError(t, err)

    var deserializedProof smt.UpdateProof
    err = json.Unmarshal(proofJSON, &deserializedProof)
    require.NoError(t, err)

    // Verify deserialized proof
    isValid = goTree.VerifyProof(
        deserializedProof.Leaf,
        deserializedProof.Index,
        deserializedProof.Enables,
        deserializedProof.Siblings,
    )
    require.True(t, isValid)
}
```

### 5. Monitoring and Alerting

```go
type SyncMonitor struct {
    goTree       *smt.SparseMerkleTree
    contractAddr common.Address
    client       *ethclient.Client
    alertChan    chan SyncAlert
}

type SyncAlert struct {
    Type        string
    Message     string
    GoRoot      string
    SolidityRoot string
    Timestamp   time.Time
}

func (sm *SyncMonitor) MonitorSync() {
    ticker := time.NewTicker(30 * time.Second)

    for range ticker.C {
        goRoot := sm.goTree.Root()
        solidityRoot, err := sm.getContractRoot()

        if err != nil {
            sm.alertChan <- SyncAlert{
                Type:      "ERROR",
                Message:   fmt.Sprintf("Failed to get contract root: %v", err),
                Timestamp: time.Now(),
            }
            continue
        }

        if goRoot != solidityRoot.Hex() {
            sm.alertChan <- SyncAlert{
                Type:         "SYNC_MISMATCH",
                Message:      "Root mismatch detected",
                GoRoot:       goRoot,
                SolidityRoot: solidityRoot.Hex(),
                Timestamp:    time.Now(),
            }
        }
    }
}
```

This comprehensive cross-platform integration approach ensures that your Go and Solidity SMT implementations work seamlessly together, providing the flexibility to leverage the strengths of both platforms while maintaining data consistency and proof compatibility.
