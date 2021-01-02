package hashring

import (
	"errors"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

// ErrNodeNotFound is the error returned if a node does not exist
var ErrNodeNotFound error = errors.New("node not found")

// ErrNotEnoughNodes is the error returned if a node does not exist
var ErrNotEnoughNodes error = errors.New("not enough nodes on the hash ring")

// Ring is the consistent hash ring
type Ring struct {
	Nodes  Nodes
	Weight int
	sync.Mutex
}

// New returns a new instance of the hash ring
func New(weight int) *Ring {
	return &Ring{Nodes: Nodes{},
		Weight: weight,
	}
}

func getFNVHash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// vNodeID generates a string key for a virtual node.
func vNodeID(id string, idx int) string {
	return strconv.Itoa(idx) + id
}

// AddNode adds a node to the hash ring, including `Weight` virtual nodes
func (r *Ring) AddNode(id string) {
	newNodes := Nodes{}
	for i := 0; i < r.Weight; i++ {
		newNodes = append(newNodes, *NewNode(id, i))
	}
	r.Lock()
	defer r.Unlock()
	r.Nodes = append(r.Nodes, newNodes...)
	sort.Sort(r.Nodes) //TODO: is a tree implementation more effecient?
}

// RebalanceInstructions returns the instructions for where to find data
// to add to a new node
func (r *Ring) RebalanceInstructions(id string) ([]Instruction, error) {
	r.Lock()
	defer r.Unlock()

	instr := []Instruction{}

	for idx := 0; idx < r.Weight; idx++ {
		hash := getFNVHash(vNodeID(id, idx))
		i := r.binSearch(hash)
		if i >= r.Nodes.Len() || r.Nodes[i].ID != id {
			return nil, ErrNodeNotFound
		}
		j := i - 1
		if j == -1 {
			j = r.Nodes.Len() - 1
		}
		ins := Instruction{
			FromNode: r.Nodes[j].ID,
			LowHash:  r.Nodes[j].HashID,
			HighHash: r.Nodes[i].HashID,
		}
		instr = append(instr, ins)
	}

	return instr, nil
}

// RemoveNode will remove a node from the hash ring, including it's virtual nodes
func (r *Ring) RemoveNode(id string) error {
	r.Lock()
	defer r.Unlock()

	for idx := 0; idx < r.Weight; idx++ {
		hash := getFNVHash(vNodeID(id, idx))
		i := r.binSearch(hash)
		if i >= r.Nodes.Len() || r.Nodes[i].ID != id {
			return ErrNodeNotFound
		}

		r.Nodes = append(r.Nodes[:i], r.Nodes[i+1:]...)
	}

	return nil
}

// Get returns the nearest node greater than the hash of key
func (r *Ring) Get(key string) string {
	hashKey := getFNVHash(key)
	i := r.binSearch(hashKey)
	if i >= r.Nodes.Len() {
		i = 0
	}
	return r.Nodes[i].ID
}

// GetN returns the N nearest unique nodes greater than the hash of the key
func (r *Ring) GetN(key string, n int) ([]string, error) {
	if n == 1 {
		return []string{r.Get(key)}, nil
	}
	if r.Nodes.Len()/r.Weight < n {
		return nil, ErrNotEnoughNodes
	}
	var i int
	nodeIDs := make(map[string]struct{})
	output := []string{}
	hashKey := getFNVHash(key)
	i = r.binSearch(hashKey)
	if i >= r.Nodes.Len() {
		i = 0
	}
	nodeIDs[r.Nodes[i].ID] = struct{}{}
	output = append(output, r.Nodes[i].ID)
	left := n - 1
	for left > 0 {
		i = i + 1
		if i >= r.Nodes.Len() {
			i = 0
		}
		if _, ok := nodeIDs[r.Nodes[i].ID]; !ok {
			nodeIDs[r.Nodes[i].ID] = struct{}{}
			output = append(output, r.Nodes[i].ID)
			left--
		}
	}

	return output, nil
}

func (r *Ring) binSearch(hash uint32) int {
	searchfn := func(i int) bool {
		return r.Nodes[i].HashID >= hash
	}

	return sort.Search(r.Nodes.Len(), searchfn)
}
