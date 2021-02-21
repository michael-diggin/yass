package models

// Node is the struct holding the info of the servers
type Node struct {
	ID     string
	Idx    int
	HashID uint32
}

// Instruction is the struct that holds the
// hash range information used for rebalacing data from a node
type Instruction struct {
	FromNode string
	FromIdx  int
	ToIdx    int
	LowHash  uint32
	HighHash uint32
}

//go:generate mockgen -destination=../mocks/mock_hash_ring.go -package=mocks . HashRing

// HashRing is the interface for the Consistent Hash Ring
type HashRing interface {
	AddNode(id string)
	RemoveNode(id string) error
	Get(uint32) Node
	GetN(uint32, int) ([]Node, error)
	RebalanceInstructions(string) []Instruction
	Hash(string) uint32
}
