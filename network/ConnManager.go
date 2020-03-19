package network

import "sync"

type IConnManager interface {
	Modify(tId string, sId int64, c IConn)
	Get(tId string, sId int64) IConn
	GetServerTypeAll(tId string) map[int64]IConn
	Remove(tId string, sId int64)
}

var ConnManagerMgr IConnManager

var mu sync.Mutex

type connManagerData struct {
	sync.Mutex
	conns map[string]map[int64]IConn
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
		conns: map[string]map[int64]IConn{},
	}
}

func (m *connManagerData) Modify(tId string, sId int64, c IConn) {
	m.Lock()
	defer m.Unlock()
	if _, exits := m.conns[tId]; exits {
		if _, exits := m.conns[tId]; exits {
			m.conns[tId][sId] = c
			return
		}
	}
	var data map[int64]IConn
	data[sId] = c
	m.conns[tId] = data
	return
}

func (m *connManagerData) Get(tId string, sId int64) IConn {
	m.Lock()
	defer m.Unlock()
	if _, exits := m.conns[tId]; exits {
		if _, exits := m.conns[tId][sId]; exits {
			return m.conns[tId][sId]
		}
	}
	return nil
}

func (m *connManagerData) GetServerTypeAll(tId string) map[int64]IConn {
	m.Lock()
	defer m.Unlock()
	if _, exits := m.conns[tId]; exits {
		return m.conns[tId]
	}
	return nil
}

func (m *connManagerData) Remove(tId string, sId int64) {
	m.Lock()
	defer m.Unlock()
	if _, exits := m.conns[tId]; exits {
		if _, exits := m.conns[tId][sId]; exits {
			delete(m.conns[tId], sId)
		}
	}
}
