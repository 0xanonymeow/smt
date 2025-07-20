package vectors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadHashVectors loads hash test vectors from JSON file
func LoadHashVectors(filename string) ([]HashTestVector, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read hash vectors file %s: %w", filename, err)
	}
	
	var vectors []HashTestVector
	if err := json.Unmarshal(data, &vectors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hash vectors: %w", err)
	}
	
	return vectors, nil
}

// LoadProofVectors loads proof test vectors from JSON file
func LoadProofVectors(filename string) ([]ProofTestVector, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read proof vectors file %s: %w", filename, err)
	}
	
	var vectors []ProofTestVector
	if err := json.Unmarshal(data, &vectors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proof vectors: %w", err)
	}
	
	return vectors, nil
}

// LoadAddressVectors loads address test vectors from JSON file
func LoadAddressVectors(filename string) ([]AddressTestVector, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read address vectors file %s: %w", filename, err)
	}
	
	var vectors []AddressTestVector
	if err := json.Unmarshal(data, &vectors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal address vectors: %w", err)
	}
	
	return vectors, nil
}

// LoadRootComputationVectors loads root computation test vectors from JSON file
func LoadRootComputationVectors(filename string) ([]RootComputationTestVector, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read root computation vectors file %s: %w", filename, err)
	}
	
	var vectors []RootComputationTestVector
	if err := json.Unmarshal(data, &vectors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal root computation vectors: %w", err)
	}
	
	return vectors, nil
}

// SaveHashVectors saves hash test vectors to JSON file
func SaveHashVectors(filename string, vectors []HashTestVector) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	data, err := json.MarshalIndent(vectors, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal hash vectors: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write hash vectors file: %w", err)
	}
	
	return nil
}

// SaveProofVectors saves proof test vectors to JSON file
func SaveProofVectors(filename string, vectors []ProofTestVector) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	data, err := json.MarshalIndent(vectors, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal proof vectors: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write proof vectors file: %w", err)
	}
	
	return nil
}

// SaveAddressVectors saves address test vectors to JSON file
func SaveAddressVectors(filename string, vectors []AddressTestVector) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	data, err := json.MarshalIndent(vectors, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal address vectors: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write address vectors file: %w", err)
	}
	
	return nil
}

// SaveRootComputationVectors saves root computation test vectors to JSON file
func SaveRootComputationVectors(filename string, vectors []RootComputationTestVector) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	data, err := json.MarshalIndent(vectors, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal root computation vectors: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write root computation vectors file: %w", err)
	}
	
	return nil
}