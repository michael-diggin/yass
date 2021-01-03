package hashring

// Instruction is the struct that holds the
// hash range information
type Instruction struct {
	FromNode string
	FromIdx  int
	ToIdx    int
	LowHash  uint32
	HighHash uint32
}
