package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetryWithExpBackOffSuccess(t *testing.T) {
	var global int = 0
	f := func() error {
		global = global + 1
		return nil
	}

	err := WithBackOff(f, 10, 1*time.Second, "")
	require.NoError(t, err)
	require.Equal(t, 1, global)
}

func TestRetryWithExpBackOffFailsThenSuccess(t *testing.T) {
	var global int = 0
	f := func() error {
		if global == 3 {
			return nil
		}
		global = global + 1
		return errors.New("test error")
	}

	err := WithBackOff(f, 5, 10*time.Millisecond, "")
	require.NoError(t, err)
	require.Equal(t, 3, global)
}

func TestRetryWithExpBackOffFails(t *testing.T) {
	testErr := errors.New("test error")
	f := func() error {
		return testErr
	}

	err := WithBackOff(f, 2, 10*time.Millisecond, "")
	require.Error(t, err)
	require.True(t, errors.Is(err, testErr))
}
