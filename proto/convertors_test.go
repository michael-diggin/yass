package proto

import (
	"testing"

	"github.com/michael-diggin/yass/models"
	"github.com/stretchr/testify/require"
)

func TestToModelPair(t *testing.T) {
	pbPair := Pair{Key: "test-key", Value: []byte(`"test-value"`)}
	modelPair, err := pbPair.ToModel()

	require.NoError(t, err)
	require.Equal(t, "test-key", modelPair.Key)
	require.Equal(t, "test-value", modelPair.Value)
}

func TestToModelPairWithNoValues(t *testing.T) {
	pbPair := Pair{}
	modelPair, err := pbPair.ToModel()

	require.NoError(t, err)
	require.Equal(t, "", modelPair.Key)
	require.Nil(t, modelPair.Value)
}

func TestToModelPairWithBadData(t *testing.T) {
	pbPair := Pair{Value: []byte(`bad-data`)}
	modelPair, err := pbPair.ToModel()

	require.Error(t, err)
	require.Nil(t, modelPair)
}

func TestToProtoPair(t *testing.T) {
	pbPair := Pair{Key: "test-key", Value: []byte(`"test-value"`)}
	modelPair := &models.Pair{Key: "test-key", Value: "test-value"}

	pair, err := ToPair(modelPair)

	require.NoError(t, err)
	require.Equal(t, pbPair.Key, pair.Key)
	require.Equal(t, pbPair.Value, pair.Value)
}

func TestToProtoPairWithNoValues(t *testing.T) {
	modelPair := &models.Pair{}
	pbPair, err := ToPair(modelPair)

	require.NoError(t, err)
	require.Equal(t, "", pbPair.Key)
	require.Nil(t, modelPair.Value)
}

func TestToProtoPairWithBadData(t *testing.T) {
	modelPair := &models.Pair{Value: make(chan int)}
	pbPair, err := ToPair(modelPair)

	require.Error(t, err)
	require.Nil(t, pbPair)
}

func TestToReplica(t *testing.T) {
	t.Run("main", func(t *testing.T) {
		rep := ToReplica(models.MainReplica)
		require.Equal(t, Replica_MAIN, rep)
	})
	t.Run("backup", func(t *testing.T) {
		rep := ToReplica(models.BackupReplica)
		require.Equal(t, Replica_BACKUP, rep)
	})
}
