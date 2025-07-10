package smt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// SerializeProof converts a Proof to its serialized format
func SerializeProof(proof *Proof) *SerializedProof {
	siblings := make([]string, len(proof.Siblings)) // coverage-ignore
	for i, sibling := range proof.Siblings { // coverage-ignore
		siblings[i] = Bytes32ToHex(sibling)
	}
	
	exists := uint8(0)
	if proof.Exists {
		exists = 1
	}
	
	return &SerializedProof{
		Exists:   exists,
		Index:    proof.Index,
		Leaf:     Bytes32ToHex(proof.Leaf),
		Value:    Bytes32ToHex(proof.Value),
		Enables:  fmt.Sprintf("0x%x", proof.Enables),
		Siblings: siblings,
	}
}

// DeserializeProof converts a SerializedProof back to Proof
func DeserializeProof(sp *SerializedProof) (*Proof, error) {
	leaf, err := HexToBytes32(sp.Leaf)
	if err != nil {
		return nil, fmt.Errorf("invalid leaf hex: %w", err)
	}
	
	value, err := HexToBytes32(sp.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid value hex: %w", err)
	}
	
	// Parse enables
	enablesHex := strings.TrimPrefix(sp.Enables, "0x")
	enables := new(big.Int)
	if _, ok := enables.SetString(enablesHex, 16); !ok {
		return nil, fmt.Errorf("invalid enables hex: %s", sp.Enables)
	}
	
	// Parse siblings
	siblings := make([]Bytes32, len(sp.Siblings))
	for i, siblingHex := range sp.Siblings {
		sibling, err := HexToBytes32(siblingHex)
		if err != nil {
			return nil, fmt.Errorf("invalid sibling hex at index %d: %w", i, err)
		}
		siblings[i] = sibling
	}
	
	return &Proof{
		Exists:   sp.Exists != 0,
		Index:    sp.Index,
		Leaf:     leaf,
		Value:    value,
		Enables:  enables,
		Siblings: siblings,
	}, nil
}

// SerializeUpdateProof converts an UpdateProof to its serialized format
func SerializeUpdateProof(proof *UpdateProof) *SerializedUpdateProof {
	siblings := make([]string, len(proof.Siblings)) // coverage-ignore
	for i, sibling := range proof.Siblings { // coverage-ignore
		siblings[i] = Bytes32ToHex(sibling)
	}
	
	exists := uint8(0)
	if proof.Exists {
		exists = 1
	}
	
	return &SerializedUpdateProof{
		Exists:   exists,
		Index:    proof.Index,
		Leaf:     Bytes32ToHex(proof.Leaf),
		Value:    Bytes32ToHex(proof.Value),
		Enables:  fmt.Sprintf("0x%x", proof.Enables),
		Siblings: siblings,
		NewLeaf:  Bytes32ToHex(proof.NewLeaf),
	}
}

// DeserializeUpdateProof converts a SerializedUpdateProof back to UpdateProof
func DeserializeUpdateProof(sup *SerializedUpdateProof) (*UpdateProof, error) {
	// First deserialize the base proof
	baseProof := &SerializedProof{
		Exists:   sup.Exists,
		Index:    sup.Index,
		Leaf:     sup.Leaf,
		Value:    sup.Value,
		Enables:  sup.Enables,
		Siblings: sup.Siblings,
	}
	
	proof, err := DeserializeProof(baseProof)
	if err != nil { // coverage-ignore
		return nil, err
	}
	
	// Parse new leaf
	newLeaf, err := HexToBytes32(sup.NewLeaf)
	if err != nil {
		return nil, fmt.Errorf("invalid new leaf hex: %w", err)
	}
	
	return &UpdateProof{
		Exists:   proof.Exists,
		Index:    proof.Index,
		Leaf:     proof.Leaf,
		Value:    proof.Value,
		Enables:  proof.Enables,
		Siblings: proof.Siblings,
		NewLeaf:  newLeaf,
	}, nil
}

// ProofToJSON converts a proof to a JSON-friendly format
func ProofToJSON(proof *Proof) map[string]interface{} {
	siblings := make([]string, len(proof.Siblings)) // coverage-ignore
	for i, s := range proof.Siblings { // coverage-ignore
		siblings[i] = s.String()
	}
	
	return map[string]interface{}{
		"exists":   proof.Exists,
		"index":    proof.Index.String(),
		"leaf":     proof.Leaf.String(),
		"value":    proof.Value.String(),
		"enables":  fmt.Sprintf("0x%x", proof.Enables),
		"siblings": siblings,
	}
}

// UpdateProofToJSON converts an update proof to a JSON-friendly format
func UpdateProofToJSON(proof *UpdateProof) map[string]interface{} {
	base := ProofToJSON(&Proof{
		Exists:   proof.Exists,
		Index:    proof.Index,
		Leaf:     proof.Leaf,
		Value:    proof.Value,
		Enables:  proof.Enables,
		Siblings: proof.Siblings,
	})
	
	base["newLeaf"] = proof.NewLeaf.String()
	return base
}

// ParseHex parses a hex string with or without 0x prefix
func ParseHex(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	if len(s)%2 != 0 {
		s = "0" + s
	}
	return hex.DecodeString(s)
}

// FormatHex formats bytes as hex string with 0x prefix
func FormatHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

// SerializeBigInt serializes a big.Int to hex string
func SerializeBigInt(value *big.Int) string {
	if value == nil {
		return "0x0"
	}
	return fmt.Sprintf("0x%x", value)
}

// DeserializeBigInt deserializes a hex string to big.Int
func DeserializeBigInt(s string) (*big.Int, error) {
	s = strings.TrimPrefix(s, "0x")
	if s == "" {
		return big.NewInt(0), nil
	}
	
	value := new(big.Int)
	if _, ok := value.SetString(s, 16); !ok {
		return nil, fmt.Errorf("invalid big int hex: 0x%s", s)
	}
	return value, nil
}