package xid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXIDStoreIncrement(t *testing.T) {
	seed := uint64(0)
	idstore := New(seed)
	next := idstore.IncrementXID()
	require.Equal(t, uint64(1), next)
	require.Equal(t, uint64(2), idstore.IncrementXID())
}

func TestXIDStoreGet(t *testing.T) {
	seed := uint64(100)
	idstore := New(seed)
	curr := idstore.GetXID()
	require.Equal(t, seed, curr)
}

func TestXIDStoreSetNewNeed(t *testing.T) {
	seed := uint64(100)
	idstore := New(seed)

	newSeed := uint64(101)
	idstore.SetNewSeed(newSeed)
	require.Equal(t, newSeed, idstore.GetXID())

	// setting lower value does not decrease it
	idstore.SetNewSeed(seed)
	require.Equal(t, newSeed, idstore.GetXID())
}
