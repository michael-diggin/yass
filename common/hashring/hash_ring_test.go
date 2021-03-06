package hashring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashRingSetUpNodes(t *testing.T) {
	t.Run("with weight of one", func(t *testing.T) {
		r := New(1)
		r.AddNode("server1")

		require.Equal(t, 1, r.Nodes.Len())
		require.Equal(t, "server1", r.Nodes[0].ID)
		require.Equal(t, uint32(0x3f6e411e), r.Nodes[0].HashID)
	})

	t.Run("with weight greater than one", func(t *testing.T) {
		r := New(3)
		r.AddNode("server1")

		require.Equal(t, 3, r.Nodes.Len())
		require.Equal(t, "server1", r.Nodes[0].ID)
		require.Equal(t, "server1", r.Nodes[1].ID)
		require.Equal(t, "server1", r.Nodes[2].ID)

		require.Greater(t, r.Nodes[2].HashID, r.Nodes[1].HashID)
	})

	t.Run("with two nodes", func(t *testing.T) {
		r := New(1)
		r.AddNode("server2")
		r.AddNode("server1")

		require.Equal(t, 2, r.Nodes.Len())
		require.Equal(t, "server1", r.Nodes[0].ID)
		require.Equal(t, "server2", r.Nodes[1].ID)

		require.Greater(t, r.Nodes[1].HashID, r.Nodes[0].HashID)
	})
}

func TestHashRingRemoveNodes(t *testing.T) {
	r := New(2)
	r.AddNode("server1")
	r.AddNode("server2")

	t.Run("errors when node does not exist", func(t *testing.T) {
		err := r.RemoveNode("server3")
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrNodeNotFound))
	})

	t.Run("removes node and vnodes", func(t *testing.T) {
		err := r.RemoveNode("server2")
		require.NoError(t, err)

		require.Equal(t, 2, r.Nodes.Len())
		require.Equal(t, "server1", r.Nodes[0].ID)
		require.Equal(t, "server1", r.Nodes[1].ID)
	})
}

func TestHashRingGetN(t *testing.T) {
	r := New(3)
	r.AddNode("server1")
	r.AddNode("server2")
	r.AddNode("server3")

	hashkey := Hash("test-key")

	require.Equal(t, 9, r.Nodes.Len())

	t.Run("get one", func(t *testing.T) {
		nodeIDs, err := r.GetN(hashkey, 1)

		require.NoError(t, err)
		require.Len(t, nodeIDs, 1)
		require.Equal(t, "server3", nodeIDs[0].ID)
		require.Equal(t, 1, nodeIDs[0].Idx)
	})

	t.Run("get two", func(t *testing.T) {
		hashkey := uint32(3730899808)
		nodes, err := r.GetN(hashkey, 2)
		require.NoError(t, err)
		require.Len(t, nodes, 2)

		nodeIDs := []string{nodes[0].ID, nodes[1].ID}
		require.ElementsMatch(t, nodeIDs, []string{"server2", "server1"})
		require.Equal(t, 1, nodes[0].Idx)
		require.Equal(t, 1, nodes[1].Idx)
	})

	t.Run("errors if asking for too many servers", func(t *testing.T) {
		_, err := r.GetN(hashkey, 4)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrNotEnoughNodes))
	})
}

func TestRebalanceInstructions(t *testing.T) {
	r := New(3)
	r.AddNode("server1")
	r.AddNode("secondServer")
	r.AddNode("pod-3")
	r.AddNode("new-server")

	require.Equal(t, 12, r.Nodes.Len())

	instructions := r.RebalanceInstructions("new-server")

	require.Len(t, instructions, 6)
	require.Equal(t, instructions[0].FromNode, "pod-3")
	require.Equal(t, instructions[0].FromIdx, 2)
	require.Equal(t, instructions[0].ToIdx, 0)

	require.Equal(t, instructions[1].FromNode, "secondServer")
	require.Equal(t, instructions[1].FromIdx, 0)
	require.Equal(t, instructions[1].ToIdx, 0)
}
