package hashring

import "github.com/michael-diggin/yass/common/models"

// Nodes is an array of nodes
type Nodes []models.Node

func (n Nodes) Len() int           { return len(n) }
func (n Nodes) Less(i, j int) bool { return n[i].HashID < n[j].HashID }
func (n Nodes) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

// Get returns the node at the given index
func (n Nodes) Get(i int) models.Node {
	if i >= n.Len() {
		return n[0]
	}
	return n[i]
}

// NewNode returns a new instane of Node
func NewNode(id string, idx int) *models.Node {
	return &models.Node{
		ID:     id,
		Idx:    idx,
		HashID: Hash(vNodeID(id, idx)),
	}
}
