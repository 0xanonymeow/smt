# Migration Guide

This guide helps you migrate from existing SMT implementations to the production-ready libraries.

## Table of Contents

- [Migration Overview](#migration-overview)
- [From TypeScript Reference Implementation](#from-typescript-reference-implementation)
- [From Other Go SMT Libraries](#from-other-go-smt-libraries)
- [From Basic Solidity Implementations](#from-basic-solidity-implementations)
- [Data Migration Strategies](#data-migration-strategies)
- [Testing Migration](#testing-migration)
- [Rollback Procedures](#rollback-procedures)

## Migration Overview

### What's New in Production Libraries

The production SMT libraries offer significant improvements over reference implementations:

**Enhanced Features:**

- Complete CRUD operations (Insert, Update, Get, Delete)
- Comprehensive error handling with structured error types
- Performance optimizations and memory management
- Cross-platform proof compatibility
- Batch operations for improved efficiency
- Production-ready access control and security features

**Breaking Changes:**

- Enhanced API with additional methods and parameters
- Structured error handling instead of panics
- Modified proof structures with additional fields
- Updated serialization formats for consistency

### Migration Checklist

- [ ] Backup existing data and state
- [ ] Update import statements and dependencies
- [ ] Modify API calls to use new method signatures
- [ ] Update error handling to use structured errors
- [ ] Test proof compatibility between old and new implementations
- [ ] Update deployment scripts and configurations
- [ ] Verify cross-platform compatibility if applicable
- [ ] Update monitoring and logging
- [ ] Plan rollback procedures

## From TypeScript Reference Implementation

### API Changes

#### Tree Creation

**Before (TypeScript):**

```typescript
import { SparseMerkleTree } from "./SparseMerkleTree";

const tree = new SparseMerkleTree(256, 0n, keccak, deserialize, serialize);
```

**After (Go):**

```go
import "github.com/0xanonymeow/smt"

tree := smt.NewSparseMerkleTree(256, &smt.SparseMerkleTreeKVOptions{
    ZeroElement:    big.NewInt(0),
    HashFn:         smt.Keccak,
    DeserializerFn: smt.Deserialize,
    SerializerFn:   smt.Serialize,
})
```

#### Insert Operations

**Before (TypeScript):**

```typescript
const proof = tree.insert(index, leaf);
// No error handling - throws on duplicate
```

**After (Go):**

```go
proof, err := tree.Insert(index, leaf)
if err != nil {
    // Handle error - returns error instead of throwing
    var smtErr *smt.SMTError
    if errors.As(err, &smtErr) {
        switch smtErr.Code {
        case 1003: // KeyExists
            // Handle duplicate key
        default:
            // Handle other errors
        }
    }
}
```

#### Key-Value Operations

**Before (TypeScript):**

```typescript
import { SparseMerkleTreeKV } from "./SparseMerkleTreeKV";

const treeKV = new SparseMerkleTreeKV();
const proof = treeKV.insert(key, value);
```

**After (Go):**

```go
treeKV := smt.NewSparseMerkleTreeKV(nil)
proof, err := treeKV.Insert(key, value)
if err != nil {
    log.Printf("Insert failed: %v", err)
}
```

### Data Structure Migration

#### Proof Structure Changes

**TypeScript Proof:**

```typescript
interface Proof {
  exists: boolean;
  leaf: bigint;
  value: bigint | null;
  index: bigint;
  enables: bigint;
  siblings: bigint[];
}
```

**Go Proof:**

```go
type Proof struct {
    Exists   bool     `json:"exists"`
    Leaf     string   `json:"leaf"`      // Hex string instead of bigint
    Value    *string  `json:"value"`     // Hex string pointer
    Index    *big.Int `json:"index"`     // big.Int instead of bigint
    Enables  *big.Int `json:"enables"`   // big.Int instead of bigint
    Siblings []string `json:"siblings"`  // Hex strings instead of bigints
}
```

#### Migration Helper Functions

```go
// Convert TypeScript proof format to Go format
func ConvertTSProofToGo(tsProof map[string]interface{}) (*smt.Proof, error) {
    exists := tsProof["exists"].(bool)
    leaf := fmt.Sprintf("0x%s", tsProof["leaf"].(string))

    var value *string
    if tsProof["value"] != nil {
        v := fmt.Sprintf("0x%s", tsProof["value"].(string))
        value = &v
    }

    index, _ := new(big.Int).SetString(tsProof["index"].(string), 10)
    enables, _ := new(big.Int).SetString(tsProof["enables"].(string), 10)

    siblingsRaw := tsProof["siblings"].([]interface{})
    siblings := make([]string, len(siblingsRaw))
    for i, s := range siblingsRaw {
        siblings[i] = fmt.Sprintf("0x%s", s.(string))
    }

    return &smt.Proof{
        Exists:   exists,
        Leaf:     leaf,
        Value:    value,
        Index:    index,
        Enables:  enables,
        Siblings: siblings,
    }, nil
}

// Convert Go proof to TypeScript-compatible format
func ConvertGoProofToTS(goProof *smt.Proof) map[string]interface{} {
    result := map[string]interface{}{
        "exists":   goProof.Exists,
        "leaf":     strings.TrimPrefix(goProof.Leaf, "0x"),
        "index":    goProof.Index.String(),
        "enables":  goProof.Enables.String(),
        "siblings": make([]string, len(goProof.Siblings)),
    }

    if goProof.Value != nil {
        result["value"] = strings.TrimPrefix(*goProof.Value, "0x")
    } else {
        result["value"] = nil
    }

    for i, s := range goProof.Siblings {
        result["siblings"].([]string)[i] = strings.TrimPrefix(s, "0x")
    }

    return result
}
```

### Migration Script Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "math/big"
    "github.com/0xanonymeow/smt"
)

type TSTreeState struct {
    Root    string                   `json:"root"`
    Entries []TSEntry               `json:"entries"`
}

type TSEntry struct {
    Index string `json:"index"`
    Leaf  string `json:"leaf"`
    Value string `json:"value"`
}

func MigrateFromTypeScript(tsStateFile string) (*smt.SparseMerkleTree, error) {
    // Read TypeScript tree state
    data, err := ioutil.ReadFile(tsStateFile)
    if err != nil {
        return nil, err
    }

    var tsState TSTreeState
    if err := json.Unmarshal(data, &tsState); err != nil {
        return nil, err
    }

    // Create new Go tree
    goTree := smt.NewSparseMerkleTree(256, nil)

    // Migrate entries
    for _, entry := range tsState.Entries {
        index, _ := new(big.Int).SetString(entry.Index, 10)
        leaf := fmt.Sprintf("0x%s", entry.Leaf)

        _, err := goTree.Insert(index, leaf)
        if err != nil {
            log.Printf("Failed to migrate entry %s: %v", entry.Index, err)
            continue
        }
    }

    // Verify root matches
    goRoot := goTree.Root()
    expectedRoot := fmt.Sprintf("0x%s", tsState.Root)

    if goRoot != expectedRoot {
        return nil, fmt.Errorf("root mismatch: expected %s, got %s", expectedRoot, goRoot)
    }

    log.Printf("Successfully migrated %d entries", len(tsState.Entries))
    return goTree, nil
}
```

## From Other Go SMT Libraries

### Common Migration Patterns

#### From github.com/iden3/go-merkletree

**Before:**

```go
import "github.com/iden3/go-merkletree"

mt, err := merkletree.NewMerkleTree(storage, 40)
err = mt.Add(key, value)
proof, err := mt.GenerateProof(key)
```

**After:**

```go
import "github.com/0xanonymeow/smt"

tree := smt.NewSparseMerkleTree(256, nil)
index := smt.Deserialize(key)
leaf := smt.Serialize(value)

proof, err := tree.Insert(index, leaf)
getProof := tree.Get(index)
```

#### Migration Helper

```go
func MigrateFromIden3(oldStorage merkletree.Storage) (*smt.SparseMerkleTree, error) {
    newTree := smt.NewSparseMerkleTree(256, nil)

    // Iterate through old storage
    iter := oldStorage.Iterate()
    for iter.Next() {
        key := iter.Key()
        value := iter.Value()

        // Convert key to index
        index := new(big.Int).SetBytes(key)

        // Convert value to leaf
        leaf := smt.Serialize(new(big.Int).SetBytes(value))

        _, err := newTree.Insert(index, leaf)
        if err != nil {
            return nil, fmt.Errorf("migration failed for key %x: %w", key, err)
        }
    }

    return newTree, nil
}
```

### Performance Comparison

```go
func BenchmarkMigrationPerformance(b *testing.B) {
    // Setup test data
    testData := generateTestData(1000)

    b.Run("OldLibrary", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            // Benchmark old library operations
            oldTree := setupOldTree()
            for _, data := range testData {
                oldTree.Add(data.Key, data.Value)
            }
        }
    })

    b.Run("NewLibrary", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            // Benchmark new library operations
            newTree := smt.NewSparseMerkleTree(256, nil)
            for _, data := range testData {
                index := smt.Deserialize(data.Key)
                leaf := smt.Serialize(data.Value)
                newTree.Insert(index, leaf)
            }
        }
    })
}
```

## From Basic Solidity Implementations

### Contract Migration

#### Before (Basic Implementation):

```solidity
contract BasicSMT {
    mapping(uint256 => bytes32) public leaves;
    bytes32 public root;

    function insert(uint256 index, bytes32 leaf) external {
        require(leaves[index] == bytes32(0), "Key exists");
        leaves[index] = leaf;
        // Simple root update (not secure)
        root = keccak256(abi.encodePacked(root, leaf));
    }
}
```

#### After (Production Implementation):

```solidity
import "./SparseMerkleTreeContract.sol";

contract ProductionSMT {
    SparseMerkleTreeContract public smt;

    constructor() {
        smt = new SparseMerkleTreeContract(256, "Production SMT", "1.0.0");
    }

    function insert(uint256 index, bytes32 leaf) external {
        smt.insert(index, leaf);
    }

    function get(uint256 index) external view returns (SparseMerkleTree.Proof memory) {
        return smt.get(index);
    }
}
```

### Data Migration Contract

```solidity
contract SMTMigrator {
    using SparseMerkleTree for SparseMerkleTree.SMTStorage;

    SparseMerkleTree.SMTStorage private newSMT;

    event MigrationProgress(uint256 migrated, uint256 total);
    event MigrationComplete(bytes32 finalRoot);

    constructor() {
        newSMT.initialize(256);
    }

    function migrateFromBasicSMT(
        address oldContract,
        uint256[] calldata indices,
        bytes32[] calldata leaves
    ) external {
        require(indices.length == leaves.length, "Length mismatch");

        BasicSMT oldSMT = BasicSMT(oldContract);

        for (uint256 i = 0; i < indices.length; i++) {
            // Verify data exists in old contract
            bytes32 oldLeaf = oldSMT.leaves(indices[i]);
            require(oldLeaf == leaves[i], "Data mismatch");

            // Insert into new SMT
            newSMT.insert(indices[i], leaves[i]);
        }

        emit MigrationProgress(indices.length, indices.length);
        emit MigrationComplete(newSMT.getRoot());
    }

    function batchMigrate(
        address oldContract,
        uint256[][] calldata batches,
        bytes32[][] calldata leaveBatches
    ) external {
        require(batches.length == leaveBatches.length, "Batch length mismatch");

        uint256 totalMigrated = 0;

        for (uint256 b = 0; b < batches.length; b++) {
            migrateFromBasicSMT(oldContract, batches[b], leaveBatches[b]);
            totalMigrated += batches[b].length;

            emit MigrationProgress(totalMigrated, getTotalEntries());
        }
    }

    function getTotalEntries() internal pure returns (uint256) {
        // Return total number of entries to migrate
        return 1000; // Example
    }
}
```

## Data Migration Strategies

### Strategy 1: Offline Migration

```go
func OfflineMigration(oldDataPath, newDataPath string) error {
    // Read old data
    oldData, err := loadOldData(oldDataPath)
    if err != nil {
        return err
    }

    // Create new tree
    newTree := smt.NewSparseMerkleTree(256, nil)

    // Migrate data in batches
    batchSize := 1000
    for i := 0; i < len(oldData); i += batchSize {
        end := i + batchSize
        if end > len(oldData) {
            end = len(oldData)
        }

        batch := oldData[i:end]
        if err := migrateBatch(newTree, batch); err != nil {
            return fmt.Errorf("batch migration failed at index %d: %w", i, err)
        }

        log.Printf("Migrated %d/%d entries", end, len(oldData))
    }

    // Save new data
    return saveNewData(newTree, newDataPath)
}

func migrateBatch(tree *smt.SparseMerkleTree, batch []OldDataEntry) error {
    operations := make([]smt.BatchOperation, len(batch))

    for i, entry := range batch {
        operations[i] = smt.BatchOperation{
            Type:  "insert",
            Index: convertOldIndex(entry.Index),
            Value: convertOldValue(entry.Value),
        }
    }

    processor := smt.NewBatchProcessor(tree)
    return processor.ProcessBatch(operations)
}
```

### Strategy 2: Online Migration with Dual Write

```go
type DualWriteSMT struct {
    oldTree OldSMTInterface
    newTree *smt.SparseMerkleTree
    migrationComplete bool
    mu sync.RWMutex
}

func (d *DualWriteSMT) Insert(index *big.Int, leaf string) error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Write to new tree
    _, err := d.newTree.Insert(index, leaf)
    if err != nil {
        return err
    }

    // Write to old tree if migration not complete
    if !d.migrationComplete {
        return d.oldTree.Insert(index, leaf)
    }

    return nil
}

func (d *DualWriteSMT) Get(index *big.Int) *smt.Proof {
    d.mu.RLock()
    defer d.mu.RUnlock()

    // Try new tree first
    proof := d.newTree.Get(index)
    if proof.Exists {
        return proof
    }

    // Fallback to old tree if migration not complete
    if !d.migrationComplete {
        oldProof := d.oldTree.Get(index)
        return convertOldProof(oldProof)
    }

    return proof
}

func (d *DualWriteSMT) CompleteMigration() error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Verify consistency
    if err := d.verifyConsistency(); err != nil {
        return err
    }

    d.migrationComplete = true
    return nil
}
```

### Strategy 3: Incremental Migration

```go
type IncrementalMigrator struct {
    oldTree     OldSMTInterface
    newTree     *smt.SparseMerkleTree
    migrated    map[string]bool
    batchSize   int
    mu          sync.RWMutex
}

func (im *IncrementalMigrator) MigrateNext() error {
    im.mu.Lock()
    defer im.mu.Unlock()

    // Get next batch of unmigrated entries
    entries := im.getNextBatch()
    if len(entries) == 0 {
        return nil // Migration complete
    }

    // Migrate batch
    for _, entry := range entries {
        index := convertOldIndex(entry.Index)
        leaf := convertOldValue(entry.Value)

        _, err := im.newTree.Insert(index, leaf)
        if err != nil {
            return err
        }

        im.migrated[entry.ID] = true
    }

    log.Printf("Migrated batch of %d entries", len(entries))
    return nil
}

func (im *IncrementalMigrator) GetMigrationProgress() (int, int) {
    im.mu.RLock()
    defer im.mu.RUnlock()

    total := im.oldTree.GetTotalEntries()
    migrated := len(im.migrated)

    return migrated, total
}
```

## Testing Migration

### Migration Test Suite

```go
func TestMigrationAccuracy(t *testing.T) {
    // Setup old tree with test data
    oldTree := setupOldTreeWithData(t)

    // Perform migration
    newTree, err := MigrateFromOldTree(oldTree)
    require.NoError(t, err)

    // Verify all data migrated correctly
    testEntries := getTestEntries()
    for _, entry := range testEntries {
        // Check old tree
        oldProof := oldTree.Get(entry.Index)
        require.True(t, oldProof.Exists)

        // Check new tree
        newProof := newTree.Get(entry.Index)
        require.True(t, newProof.Exists)

        // Verify values match
        require.Equal(t, convertOldValue(oldProof.Value), *newProof.Value)
    }

    // Verify roots match (if applicable)
    if oldTree.SupportsRootHash() {
        oldRoot := oldTree.GetRoot()
        newRoot := newTree.Root()
        require.Equal(t, oldRoot, newRoot)
    }
}

func TestMigrationPerformance(t *testing.T) {
    sizes := []int{100, 1000, 10000}

    for _, size := range sizes {
        t.Run(fmt.Sprintf("Size%d", size), func(t *testing.T) {
            oldTree := setupOldTreeWithSize(t, size)

            start := time.Now()
            newTree, err := MigrateFromOldTree(oldTree)
            duration := time.Since(start)

            require.NoError(t, err)
            require.NotNil(t, newTree)

            t.Logf("Migrated %d entries in %v (%.2f entries/sec)",
                size, duration, float64(size)/duration.Seconds())
        })
    }
}

func TestMigrationRollback(t *testing.T) {
    oldTree := setupOldTreeWithData(t)

    // Create backup
    backup := createBackup(t, oldTree)

    // Perform migration
    newTree, err := MigrateFromOldTree(oldTree)
    require.NoError(t, err)

    // Simulate migration failure
    simulateFailure()

    // Rollback
    restoredTree := restoreFromBackup(t, backup)

    // Verify rollback successful
    verifyTreesEqual(t, oldTree, restoredTree)
}
```

### Cross-Platform Migration Testing

```go
func TestCrossPlatformMigration(t *testing.T) {
    // Create Go tree with test data
    goTree := smt.NewSparseMerkleTree(256, nil)
    testData := generateTestData(100)

    for _, data := range testData {
        _, err := goTree.Insert(data.Index, data.Leaf)
        require.NoError(t, err)
    }

    // Export Go tree state
    goState := exportTreeState(goTree)

    // Import into Solidity (simulated)
    solidityState := importToSolidity(t, goState)

    // Verify states match
    require.Equal(t, goTree.Root(), solidityState.Root)

    // Verify individual entries
    for _, data := range testData {
        goProof := goTree.Get(data.Index)
        solidityProof := solidityState.Get(data.Index)

        require.Equal(t, goProof.Exists, solidityProof.Exists)
        require.Equal(t, goProof.Leaf, solidityProof.Leaf)
    }
}
```

## Rollback Procedures

### Automated Rollback System

```go
type MigrationManager struct {
    oldTree    OldSMTInterface
    newTree    *smt.SparseMerkleTree
    backup     *TreeBackup
    checkpoint *MigrationCheckpoint
}

type MigrationCheckpoint struct {
    Timestamp    time.Time
    EntriesDone  int
    OldRoot      string
    NewRoot      string
    BackupPath   string
}

func (mm *MigrationManager) CreateCheckpoint() error {
    mm.checkpoint = &MigrationCheckpoint{
        Timestamp:   time.Now(),
        EntriesDone: mm.getEntriesMigrated(),
        OldRoot:     mm.oldTree.GetRoot(),
        NewRoot:     mm.newTree.Root(),
        BackupPath:  mm.createBackup(),
    }

    return mm.saveCheckpoint()
}

func (mm *MigrationManager) Rollback() error {
    if mm.checkpoint == nil {
        return fmt.Errorf("no checkpoint available for rollback")
    }

    log.Printf("Rolling back migration to checkpoint at %v", mm.checkpoint.Timestamp)

    // Restore from backup
    if err := mm.restoreFromBackup(mm.checkpoint.BackupPath); err != nil {
        return fmt.Errorf("backup restoration failed: %w", err)
    }

    // Verify rollback
    currentRoot := mm.oldTree.GetRoot()
    if currentRoot != mm.checkpoint.OldRoot {
        return fmt.Errorf("rollback verification failed: expected root %s, got %s",
            mm.checkpoint.OldRoot, currentRoot)
    }

    log.Printf("Rollback completed successfully")
    return nil
}

func (mm *MigrationManager) SafeMigrate() error {
    // Create initial checkpoint
    if err := mm.CreateCheckpoint(); err != nil {
        return err
    }

    defer func() {
        if r := recover(); r != nil {
            log.Printf("Migration panic detected: %v", r)
            mm.Rollback()
        }
    }()

    // Perform migration with periodic checkpoints
    batchSize := 1000
    totalEntries := mm.oldTree.GetTotalEntries()

    for i := 0; i < totalEntries; i += batchSize {
        // Migrate batch
        if err := mm.migrateBatch(i, batchSize); err != nil {
            log.Printf("Migration failed at batch %d: %v", i/batchSize, err)
            return mm.Rollback()
        }

        // Create checkpoint every 10 batches
        if (i/batchSize)%10 == 0 {
            if err := mm.CreateCheckpoint(); err != nil {
                log.Printf("Checkpoint creation failed: %v", err)
                // Continue migration but log warning
            }
        }
    }

    // Final verification
    return mm.verifyMigration()
}
```

### Manual Rollback Procedures

```bash
#!/bin/bash
# rollback.sh - Manual rollback script

set -e

BACKUP_DIR="/path/to/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "Starting rollback procedure at $TIMESTAMP"

# 1. Stop services
echo "Stopping services..."
systemctl stop your-smt-service

# 2. Backup current state (in case rollback fails)
echo "Creating rollback backup..."
cp -r /path/to/current/data "$BACKUP_DIR/rollback_$TIMESTAMP"

# 3. Restore from migration backup
echo "Restoring from backup..."
LATEST_BACKUP=$(ls -t $BACKUP_DIR/pre_migration_* | head -1)
cp -r "$LATEST_BACKUP"/* /path/to/current/data/

# 4. Verify restoration
echo "Verifying restoration..."
go run verify_restoration.go

# 5. Restart services
echo "Restarting services..."
systemctl start your-smt-service

# 6. Run health checks
echo "Running health checks..."
curl -f http://localhost:8080/health || exit 1

echo "Rollback completed successfully"
```

### Rollback Verification

```go
func VerifyRollback(originalBackup, currentState string) error {
    // Load original backup
    originalData, err := loadBackupData(originalBackup)
    if err != nil {
        return fmt.Errorf("failed to load original backup: %w", err)
    }

    // Load current state
    currentData, err := loadCurrentData(currentState)
    if err != nil {
        return fmt.Errorf("failed to load current state: %w", err)
    }

    // Compare roots
    if originalData.Root != currentData.Root {
        return fmt.Errorf("root mismatch: original %s, current %s",
            originalData.Root, currentData.Root)
    }

    // Compare entry count
    if len(originalData.Entries) != len(currentData.Entries) {
        return fmt.Errorf("entry count mismatch: original %d, current %d",
            len(originalData.Entries), len(currentData.Entries))
    }

    // Spot check random entries
    sampleSize := min(100, len(originalData.Entries))
    for i := 0; i < sampleSize; i++ {
        idx := rand.Intn(len(originalData.Entries))
        original := originalData.Entries[idx]
        current := currentData.Entries[idx]

        if !entriesEqual(original, current) {
            return fmt.Errorf("entry mismatch at index %d", idx)
        }
    }

    log.Printf("Rollback verification successful: %d entries verified", sampleSize)
    return nil
}
```

This comprehensive migration guide provides step-by-step instructions for migrating from various existing implementations to the production-ready SMT libraries, including data migration strategies, testing procedures, and rollback mechanisms.
