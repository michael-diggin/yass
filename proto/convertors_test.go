package proto

import (
	"testing"

	"github.com/michael-diggin/yass/common/models"
	"github.com/stretchr/testify/require"
)

func TestToModelPair(t *testing.T) {
	pbPair := Pair{Key: "test-key", Hash: uint32(100), Value: []byte(`"test-value"`)}
	modelPair, err := pbPair.ToModel()

	require.NoError(t, err)
	require.Equal(t, "test-key", modelPair.Key)
	require.Equal(t, uint32(100), modelPair.Hash)
	require.Equal(t, "test-value", modelPair.Value)
}

func TestToModelPairWithNoValues(t *testing.T) {
	pbPair := Pair{}
	modelPair, err := pbPair.ToModel()

	require.NoError(t, err)
	require.Equal(t, "", modelPair.Key)
	require.Equal(t, uint32(0), modelPair.Hash)
	require.Nil(t, modelPair.Value)
}

func TestToModelPairWithBadData(t *testing.T) {
	pbPair := Pair{Value: []byte(`bad-data`)}
	modelPair, err := pbPair.ToModel()

	require.Error(t, err)
	require.Nil(t, modelPair)
}

func TestToProtoPair(t *testing.T) {
	pbPair := Pair{Key: "test-key", Hash: uint32(100), Value: []byte(`"test-value"`)}
	modelPair := &models.Pair{Key: "test-key", Hash: uint32(100), Value: "test-value"}

	pair, err := ToPair(modelPair)

	require.NoError(t, err)
	require.Equal(t, pbPair.Key, pair.Key)
	require.Equal(t, pbPair.Hash, pair.Hash)
	require.Equal(t, pbPair.Value, pair.Value)
}

func TestToProtoPairWithNoValues(t *testing.T) {
	modelPair := &models.Pair{}
	pbPair, err := ToPair(modelPair)

	require.NoError(t, err)
	require.Equal(t, "", pbPair.Key)
	require.Equal(t, uint32(0), pbPair.Hash)
	require.Nil(t, modelPair.Value)
}

func TestToProtoPairWithBadData(t *testing.T) {
	modelPair := &models.Pair{Value: make(chan int)}
	pbPair, err := ToPair(modelPair)

	require.Error(t, err)
	require.Nil(t, pbPair)
}
