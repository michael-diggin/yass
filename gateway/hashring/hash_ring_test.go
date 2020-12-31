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
		require.Equal(t, uint32(0xdd6101cf), r.Nodes[0].HashID)
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

	key := "test-key"

	require.Equal(t, 9, r.Nodes.Len())

	t.Run("get one", func(t *testing.T) {
		nodeIDs, err := r.GetN(key, 1)

		require.NoError(t, err)
		require.Len(t, nodeIDs, 1)
		require.Equal(t, "server2", nodeIDs[0])
	})

	t.Run("get two", func(t *testing.T) {
		nodeIDs, err := r.GetN(key, 2)

		require.NoError(t, err)
		require.Len(t, nodeIDs, 2)
		require.ElementsMatch(t, nodeIDs, []string{"server2", "server3"})
	})

	t.Run("errors if asking for too many servers", func(t *testing.T) {
		_, err := r.GetN(key, 4)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrNotEnoughNodes))
	})
}
