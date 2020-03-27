package network

import (
	"errors"
	"sync"
)

type IConnManager interface {
	Modify(sId uint64, c IConn)
	Get(sId uint64) IConn
	GetLen() int
	Remove(sId uint64) error
	Clear()
}

var ConnManagerMgr IConnManager

var mu sync.Mutex

type connManagerData struct {
	sync.Mutex
	conns map[uint64]IConn
}

func ConnManagerInstance() IConnManager {
	mu.Lock()
	defer mu.Unlock()
	if ConnManagerMgr == nil {
		ConnManagerMgr = newConnManager()
	}
	return ConnManagerMgr
}

func newConnManager() IConnManager {
	return &connManagerData{
		conns: map[uint64]IConn{},
	}
}

func (m *connManagerData) Modify(sId uint64, c IConn) {
	m.Lock()
	defer m.Unlock()
	m.conns[sId] = c
	return
}

func (m *connManagerData) Get(sId uint64) IConn {
	m.Lock()
	defer m.Unlock()
	if _, exits := m.conns[sId]; exits {
		return m.conns[sId]
	}
	return nil
}

func (m *connManagerData) GetLen() int {
	return len(m.conns)
}

func (m *connManagerData) Remove(sId uint64) error {
	m.Lock()
	defer m.Unlock()
	if _, exits := m.conns[sId]; exits {
		delete(m.conns, sId)
		return nil
	}
	return errors.New("not found ConnectManagerMap key")
}

func (m *connManagerData) Clear() {
	m.Lock()
	defer m.Unlock()
	for key, conn := range m.conns {
		conn.Close()
		delete(m.conns, key)
	}
}
