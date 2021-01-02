package hashring

// Nodes is an array of nodes
type Nodes []Node

// Node is the struct holding the info of the servers
type Node struct {
	ID     string
	VID    string
	HashID uint32
}

func (n Nodes) Len() int           { return len(n) }
func (n Nodes) Less(i, j int) bool { return n[i].HashID < n[j].HashID }
func (n Nodes) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

// NewNode returns a new instane of Node
func NewNode(id string, idx int) *Node {
	return &Node{
		ID:     id,
		VID:    vNodeID(id, idx),
		HashID: Hash(vNodeID(id, idx)),
	}
}
