package distributed

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/hashicorp/raft"
	"github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/kv"
	"google.golang.org/protobuf/proto"
)

type fsm struct {
	db *kv.DB
}

var _ raft.FSM = (*fsm)(nil)

func (f *fsm) Apply(record *raft.Log) interface{} {
	buf := record.Data
	reqType := RequestType(buf[0])
	switch reqType {
	case SetRequestType:
		return f.append(buf[1:])
	}
	return nil
}

func (f *fsm) append(buf []byte) interface{} {
	var req api.SetRequest
	err := proto.Unmarshal(buf, &req)
	if err != nil {
		return err
	}
	return f.db.Set(req.Record)
}

var _ raft.FSMSnapshot = (*snapshot)(nil)

type snapshot struct {
	reader io.Reader
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	r := f.db.LogReader()
	return &snapshot{reader: r}, nil
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s.reader); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *snapshot) Release() {}

func (f *fsm) Restore(r io.ReadCloser) error {
	lenWidth := 8
	var enc = binary.BigEndian
	b := make([]byte, lenWidth)
	var buf bytes.Buffer
	for i := 0; ; i++ {
		_, err := io.ReadFull(r, b)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		size := int64(enc.Uint64(b))
		if _, err := io.CopyN(&buf, r, size); err != nil {
			return err
		}
		record := &api.Record{}
		if err := proto.Unmarshal(buf.Bytes(), record); err != nil {
			return err
		}
		if i == 0 {
			f.db.LogConfig.Segment.InitialOffset = record.Offset
			if err := f.db.Restore(); err != nil {
				return err
			}
		}
		if err := f.db.Set(record); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}
