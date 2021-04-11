package xid

import "sync"

type IDStore struct {
	xid  uint64
	lock sync.RWMutex
}

func New(seed uint64) *IDStore {
	return &IDStore{xid: seed, lock: sync.RWMutex{}}
}

func (s *IDStore) IncrementXID() uint64 {
	s.lock.Lock()
	s.xid++
	curr := s.xid
	s.lock.Unlock()
	return curr
}

func (s *IDStore) GetXID() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.xid
}

func (s *IDStore) SetNewSeed(seed uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if seed > s.xid {
		s.xid = seed
	}
	return
}
