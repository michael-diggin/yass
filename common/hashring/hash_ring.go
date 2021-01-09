package hashring

import (
	"errors"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"

	"github.com/michael-diggin/yass/common/models"
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

// Hash returns the hash of the key
func (r *Ring) Hash(key string) uint32 {
	return Hash(key)
}

// Hash returns the fnv hash of key
func Hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// vNodeID generates a string key for a virtual node.
func vNodeID(id string, idx int) string {
	return id + "-" + strconv.Itoa(int(Hash(strconv.Itoa(idx))))
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

func handleOOB(i, len int) int {
	if i < 0 {
		return i + len
	}
	if i >= len {
		return i - len
	}
	return i
}

// RebalanceInstructions returns the instructions for where to find data
// to add to a new node
func (r *Ring) RebalanceInstructions(id string) []models.Instruction {
	r.Lock()
	defer r.Unlock()

	instr := []models.Instruction{}

	len := r.Nodes.Len()

	for i := 0; i < len; i++ {
		if r.Nodes[i].ID != id {
			continue
		}

		m := handleOOB(i-2, len)
		j := handleOOB(i-1, len)
		k := handleOOB(i+1, len)

		// FromNode (k) has to be different to the new node
		for {
			if r.Nodes[k].ID != id {
				break
			}
			k = handleOOB(k+1, len)
		}

		n := handleOOB(k+1, len)

		// FromNode (n) has to be different to the new node
		for {
			if r.Nodes[n].ID != id {
				break
			}
			n = handleOOB(n+1, len)
		}

		// hash range just before this vNode, from different Node 2 after it
		// this is what this node is mainly for
		instructOne := models.Instruction{
			FromNode: r.Nodes[n].ID,
			FromIdx:  r.Nodes[n].Idx,
			ToIdx:    r.Nodes[i].Idx,
			LowHash:  r.Nodes[j].HashID,
			HighHash: r.Nodes[i].HashID,
		}

		// hash range two before this vNode, from different Node 1 after it
		// this is the range that gets replicated to this node
		instructTwo := models.Instruction{
			FromNode: r.Nodes[k].ID,
			FromIdx:  r.Nodes[k].Idx,
			ToIdx:    r.Nodes[i].Idx,
			LowHash:  r.Nodes[m].HashID,
			HighHash: r.Nodes[j].HashID,
		}

		instr = append(instr, instructOne, instructTwo)
	}

	return instr
}

// RemoveNode will remove a node from the hash ring, including it's virtual nodes
func (r *Ring) RemoveNode(id string) error {
	r.Lock()
	defer r.Unlock()

	for idx := 0; idx < r.Weight; idx++ {
		hash := Hash(vNodeID(id, idx))
		i := r.binSearch(hash)
		if i >= r.Nodes.Len() || r.Nodes[i].ID != id {
			return ErrNodeNotFound
		}

		r.Nodes = append(r.Nodes[:i], r.Nodes[i+1:]...)
	}

	return nil
}

// Get returns the nearest node greater than the hash of key
func (r *Ring) Get(hashkey uint32) models.Node {
	i := r.binSearch(hashkey)
	if i >= r.Nodes.Len() {
		i = 0
	}
	return r.Nodes[i]
}

// GetN returns the N nearest unique nodes greater than the hash of the key
func (r *Ring) GetN(hashkey uint32, n int) ([]models.Node, error) {
	if n == 1 {
		return []models.Node{r.Get(hashkey)}, nil
	}
	if r.Nodes.Len()/r.Weight < n {
		return nil, ErrNotEnoughNodes
	}
	var i int
	nodeIDs := make(map[string]struct{})
	output := []models.Node{}
	i = r.binSearch(hashkey)
	if i >= r.Nodes.Len() {
		i = 0
	}
	nodeIDs[r.Nodes[i].ID] = struct{}{}
	output = append(output, r.Nodes[i])
	left := n - 1
	for left > 0 {
		i = i + 1
		if i >= r.Nodes.Len() {
			i = 0
		}
		if _, ok := nodeIDs[r.Nodes[i].ID]; !ok {
			nodeIDs[r.Nodes[i].ID] = struct{}{}
			output = append(output, r.Nodes[i])
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
