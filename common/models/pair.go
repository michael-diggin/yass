package models

// Pair type is the base type for key/value pairs
type Pair struct {
	Key   string      `json:"key"`
	Hash  uint32      `json:",omitempty"`
	Value interface{} `json:"value"`
}
