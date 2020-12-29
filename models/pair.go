package models

// Pair type is the base type for key/value pairs
type Pair struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
