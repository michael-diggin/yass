package api

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/michael-diggin/yass/common/mocks"
	"github.com/stretchr/testify/require"
)

func setUpTestFile(t *testing.T, data string) *os.File {
	content := []byte(data)
	tmpfile, err := ioutil.TempFile("", "wt-test-file")
	require.NoError(t, err)

	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)
	return tmpfile
}

func TestWTLoadData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmpfile := setUpTestFile(t, "node-1\nnode-2\nnode-3")

	defer os.Remove(tmpfile.Name()) // clean up

	factory := mocks.NewMockClientFactory(ctrl)
	ring := mocks.NewMockHashRing(ctrl)
	for _, node := range []string{"node-1", "node-2", "node-3"} {
		factory.EXPECT().New(gomock.Any(), node).Return(mocks.NewMockClientInterface(ctrl), nil)
		ring.EXPECT().AddNode(node)
	}
	wt := NewWatchTower(3, 10, factory, tmpfile.Name())
	wt.hashRing = ring

	err := wt.LoadData()

	require.NoError(t, err)
}
