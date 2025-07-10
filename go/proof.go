package smt

import (
	"math/big"
)

// VerifyProof verifies a proof against a given root
func VerifyProof(root Bytes32, depth uint16, proof *Proof) bool {
	computedRoot := ComputeRootFromProof(depth, proof)
	return computedRoot == root
}

// ComputeRootFromProof computes the root hash from a proof
func ComputeRootFromProof(depth uint16, proof *Proof) Bytes32 {
	if proof == nil {
		return Bytes32{}
	}
	
	// For non-existence proofs, we start with zero and build up the path
	// For existence proofs, we start with the computed leaf hash
	var current Bytes32
	if proof.Exists {
		current = ComputeLeafHash(proof.Index, proof.Value)
	} else {
		// For non-existence proofs, we don't compute a leaf hash
		// We start with zero (empty subtree)
		current = Bytes32{}
	}
	siblingIndex := 0
	
	// Rebuild root from leaf->root (LSB->MSB)
	// Only process levels where we have siblings or non-zero current
	for i := uint(0); i < uint(depth); i++ {
		bit := GetBit(proof.Index, i)
		var sibling Bytes32
		
		// Check if sibling is enabled (non-zero)
		if GetBit(proof.Enables, i) == 1 {
			if siblingIndex < len(proof.Siblings) {
				sibling = proof.Siblings[siblingIndex]
				siblingIndex++
			}
		}
		
		// For non-existence proofs, if both current and sibling are zero,
		// the result stays zero (empty subtree)
		if current.IsZero() && sibling.IsZero() {
			continue
		}
		
		// Compute parent hash
		if bit == 1 {
			// Current is right child
			current = HashBytes32(sibling, current)
		} else {
			// Current is left child
			current = HashBytes32(current, sibling)
		}
	}
	
	return current
}

// VerifyProofWithLeaf verifies a proof for a specific leaf value
func VerifyProofWithLeaf(root Bytes32, depth uint16, leaf Bytes32, index *big.Int, enables *big.Int, siblings []Bytes32) bool {
	proof := &Proof{
		Exists:   !leaf.IsZero(),
		Leaf:     leaf,  // This is already a computed leaf hash
		Value:    leaf,  // In this case, leaf and value are the same
		Index:    index,
		Enables:  enables,
		Siblings: siblings,
	}
	return VerifyProof(root, depth, proof)
}

// ComputeRootWithLeaf computes root from leaf and proof components
func ComputeRootWithLeaf(depth uint16, leaf Bytes32, index *big.Int, enables *big.Int, siblings []Bytes32) Bytes32 {
	proof := &Proof{
		Exists:   !leaf.IsZero(),
		Leaf:     leaf,  // This is already a computed leaf hash
		Value:    leaf,  // In this case, leaf and value are the same
		Index:    index,
		Enables:  enables,
		Siblings: siblings,
	}
	return ComputeRootFromProof(depth, proof)
}

// VerifyUpdateProof verifies an update proof
func VerifyUpdateProof(oldRoot, newRoot Bytes32, depth uint16, updateProof *UpdateProof) bool {
	// Verify old proof
	oldProof := &Proof{
		Exists:   updateProof.Exists,
		Leaf:     updateProof.Leaf,
		Value:    updateProof.Value,
		Index:    updateProof.Index,
		Enables:  updateProof.Enables,
		Siblings: updateProof.Siblings,
	}
	
	if !VerifyProof(oldRoot, depth, oldProof) {
		return false
	}
	
	// Compute new root with new leaf
	newProof := &Proof{
		Exists:   true,
		Leaf:     updateProof.NewLeaf,  // This is already a computed leaf hash
		Value:    updateProof.NewLeaf,  // In update proofs, NewLeaf is the hash
		Index:    updateProof.Index,
		Enables:  updateProof.Enables,
		Siblings: updateProof.Siblings,
	}
	
	computedNewRoot := ComputeRootFromProof(depth, newProof)
	return computedNewRoot == newRoot
}

// BatchVerifyProof verifies multiple proofs efficiently
func BatchVerifyProof(root Bytes32, depth uint16, proofs []*Proof) []bool {
	results := make([]bool, len(proofs))
	for i, proof := range proofs {
		results[i] = VerifyProof(root, depth, proof)
	}
	return results
}

// BatchComputeRoot computes roots for multiple proofs
func BatchComputeRoot(depth uint16, proofs []*Proof) []Bytes32 {
	roots := make([]Bytes32, len(proofs))
	for i, proof := range proofs {
		roots[i] = ComputeRootFromProof(depth, proof)
	}
	return roots
}

// ProofPath represents the path taken during proof generation
type ProofPath struct {
	Nodes    []Bytes32
	Bits     []uint
	Siblings []Bytes32
}

// GenerateProofPath generates the complete path for a proof
func GenerateProofPath(smt *SparseMerkleTree, index *big.Int) (*ProofPath, error) {
	path := &ProofPath{
		Nodes:    make([]Bytes32, 0, smt.depth),
		Bits:     make([]uint, 0, smt.depth),
		Siblings: make([]Bytes32, 0, smt.depth),
	}
	
	current := smt.root
	
	for i := smt.depth - 1; i >= 0 && i < smt.depth; i-- {
		path.Nodes = append(path.Nodes, current)
		
		if current.IsZero() {
			break
		}
		
		node, err := smt.getNode(current)
		if err != nil { // coverage-ignore
			return nil, err
		}
		
		if node.IsEmpty() { // coverage-ignore
			break
		}
		
		bit := GetBit(index, uint(i))
		path.Bits = append(path.Bits, bit)
		
		if bit == 0 {
			path.Siblings = append(path.Siblings, node.Right)
			current = node.Left
		} else {
			path.Siblings = append(path.Siblings, node.Left)
			current = node.Right
		}
	}
	
	return path, nil
}