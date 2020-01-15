package tcp

import (
	"github.com/yaice-rx/yaice/network"
	"sync"
)

type ConnManager struct {
	Connects sync.Map
}

var ConnManagerMgr = _NewConnManager()

// 初始化
func _NewConnManager() network.IConnManager {
	manager := &ConnManager{
		Connects: sync.Map{},
	}
	manager.Connects.SetCounter()
	return manager
}

func (c *ConnManager) Add(conn network.IConn) {
	c.Connects.Store(conn.GetGuid(), conn)
}

func (c *ConnManager) Remove(guid string) {
	c.Connects.Delete(guid)
}

// 获取连接数量
func (c *ConnManager) Len() int64 {
	return *c.Connects.Length()
}
