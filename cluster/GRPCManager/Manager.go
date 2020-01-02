package GRPCManager

import "sync"

type IGRPCManager interface {
	AddConn(conn IGRPCConn)
	DelConn(typeId string)
}

type GRPCManager struct {
	sync.Mutex
	Conns map[string]IGRPCConn
}

var ManagerMgr = _NewManager()

func _NewManager() IGRPCManager {
	return &GRPCManager{
		Conns: make(map[string]IGRPCConn),
	}
}

func (m *GRPCManager) AddConn(conn IGRPCConn) {
	m.Lock()
	defer m.Unlock()
	m.Conns[conn.GetTypeId()] = conn
}

func (m *GRPCManager) DelConn(typeId string) {
	m.Lock()
	defer m.Unlock()
	delete(m.Conns, typeId)
}
