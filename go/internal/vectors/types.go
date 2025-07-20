package vectors

// HashTestVector represents a test case for hash function compatibility
type HashTestVector struct {
	Left     string `json:"left"`
	Right    string `json:"right"`
	Expected string `json:"expected"`
}

// ProofTestVector represents a test case for proof verification
type ProofTestVector struct {
	TreeDepth uint16   `json:"treeDepth"`
	Leaf      string   `json:"leaf"`
	Index     string   `json:"index"`
	Enables   string   `json:"enables"`
	Siblings  []string `json:"siblings"`
	Expected  string   `json:"expected"`
}

// AddressTestVector represents a test case for AddressKeyedSMT
type AddressTestVector struct {
	Address   string   `json:"address"`
	Value     string   `json:"value"`
	OldLeaf   string   `json:"oldLeaf"`
	Enables   string   `json:"enables"`
	Siblings  []string `json:"siblings"`
	Expected  string   `json:"expected"`
}

// RootComputationTestVector represents a test case for root computation
type RootComputationTestVector struct {
	TreeDepth uint16   `json:"treeDepth"`
	Leaf      string   `json:"leaf"`
	Index     string   `json:"index"`
	Enables   string   `json:"enables"`
	Siblings  []string `json:"siblings"`
	Expected  string   `json:"expected"`
}