package batch

import (
	"fmt"
	"math/big"
	"sync"

	smt "github.com/0xanonymeow/smt/go"
	"github.com/0xanonymeow/smt/go/internal/pool"
)

// BatchOperation represents a single operation in a batch
type BatchOperation struct {
	Type  OperationType
	Index *big.Int
	Value smt.Bytes32
}

// OperationType defines the type of batch operation
type OperationType int

const (
	Insert OperationType = iota
	Update
	Delete
)

// BatchResult contains the result of a batch operation
type BatchResult struct {
	Index       *big.Int
	Success     bool
	Error       error
	Proof       *smt.Proof
	UpdateProof *smt.UpdateProof
}

// BatchProcessor handles batch operations with optimizations
type BatchProcessor struct {
	tree     *smt.SparseMerkleTree
	pool     *pool.BigIntPool
	maxBatch int
	mu       sync.RWMutex
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(tree *smt.SparseMerkleTree, maxBatchSize int) *BatchProcessor {
	return &BatchProcessor{
		tree:     tree,
		pool:     pool.NewBigIntPool(),
		maxBatch: maxBatchSize,
	}
}

// ProcessBatch processes a batch of operations with optimizations
func (bp *BatchProcessor) ProcessBatch(operations []BatchOperation) ([]BatchResult, error) {
	if len(operations) == 0 {
		return nil, nil
	}

	// Split large batches into smaller chunks to manage memory
	if len(operations) > bp.maxBatch {
		return bp.processLargeBatch(operations)
	}

	bp.mu.Lock()
	defer bp.mu.Unlock()

	results := make([]BatchResult, len(operations))

	// Pre-allocate memory for better performance
	indices := make([]*big.Int, len(operations))
	for i, op := range operations {
		if op.Index != nil {
			indices[i] = bp.pool.GetCopy(op.Index)
		}
	}

	// Process operations
	for i, op := range operations {
		result := bp.processOperation(op, indices[i])
		results[i] = result
	}

	// Return indices to pool
	for _, idx := range indices {
		if idx != nil {
			bp.pool.Put(idx)
		}
	}

	return results, nil
}

// processLargeBatch splits large batches into smaller chunks
func (bp *BatchProcessor) processLargeBatch(operations []BatchOperation) ([]BatchResult, error) {
	var allResults []BatchResult

	for i := 0; i < len(operations); i += bp.maxBatch {
		end := i + bp.maxBatch
		if end > len(operations) {
			end = len(operations)
		}

		chunk := operations[i:end]
		results, err := bp.ProcessBatch(chunk)
		if err != nil {
			return allResults, err
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// processOperation processes a single operation
func (bp *BatchProcessor) processOperation(op BatchOperation, index *big.Int) BatchResult {
	result := BatchResult{
		Index:   index,
		Success: false,
	}

	switch op.Type {
	case Insert:
		updateProof, err := bp.tree.Insert(index, op.Value)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
			result.UpdateProof = updateProof
		}

	case Update:
		updateProof, err := bp.tree.Update(index, op.Value)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
			result.UpdateProof = updateProof
		}

	default:
		result.Error = fmt.Errorf("unsupported operation type: %d", op.Type)
	}

	return result
}

// BatchInsert performs batch insert operations
func (bp *BatchProcessor) BatchInsert(indices []*big.Int, values []smt.Bytes32) ([]BatchResult, error) {
	if len(indices) != len(values) {
		return nil, fmt.Errorf("indices and values length mismatch: %d != %d", len(indices), len(values))
	}

	operations := make([]BatchOperation, len(indices))
	for i, idx := range indices {
		operations[i] = BatchOperation{
			Type:  Insert,
			Index: idx,
			Value: values[i],
		}
	}

	return bp.ProcessBatch(operations)
}

// BatchUpdate performs batch update operations
func (bp *BatchProcessor) BatchUpdate(indices []*big.Int, values []smt.Bytes32) ([]BatchResult, error) {
	if len(indices) != len(values) {
		return nil, fmt.Errorf("indices and values length mismatch: %d != %d", len(indices), len(values))
	}

	operations := make([]BatchOperation, len(indices))
	for i, idx := range indices {
		operations[i] = BatchOperation{
			Type:  Update,
			Index: idx,
			Value: values[i],
		}
	}

	return bp.ProcessBatch(operations)
}

// ParallelBatchProcessor handles concurrent batch operations
type ParallelBatchProcessor struct {
	processors []*BatchProcessor
	numWorkers int
}

// NewParallelBatchProcessor creates a new parallel batch processor
func NewParallelBatchProcessor(trees []*smt.SparseMerkleTree, numWorkers int) *ParallelBatchProcessor {
	processors := make([]*BatchProcessor, len(trees))
	for i, tree := range trees {
		processors[i] = NewBatchProcessor(tree, 100) // Default batch size
	}

	return &ParallelBatchProcessor{
		processors: processors,
		numWorkers: numWorkers,
	}
}

// ProcessParallelBatch processes operations across multiple trees in parallel
func (pbp *ParallelBatchProcessor) ProcessParallelBatch(operations []BatchOperation) ([][]BatchResult, error) {
	if len(operations) == 0 {
		return nil, nil
	}

	// Distribute operations across workers
	chunkSize := (len(operations) + pbp.numWorkers - 1) / pbp.numWorkers
	results := make([][]BatchResult, pbp.numWorkers)
	errors := make([]error, pbp.numWorkers)

	var wg sync.WaitGroup

	for i := 0; i < pbp.numWorkers && i < len(pbp.processors); i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			start := workerID * chunkSize
			end := start + chunkSize
			if end > len(operations) {
				end = len(operations)
			}

			if start < len(operations) {
				chunk := operations[start:end]
				workerResults, err := pbp.processors[workerID].ProcessBatch(chunk)
				results[workerID] = workerResults
				errors[workerID] = err
			}
		}(i)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("worker %d error: %v", i, err)
		}
	}

	return results, nil
}

