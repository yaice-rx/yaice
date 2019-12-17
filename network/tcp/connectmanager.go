package tcp

import (
	"github.com/yaice-rx/yaice/network"
	"sync"
)

type ConnManager struct {
	sync.RWMutex
	//map[guid]连接句柄
	Connects map[string]map[string]network.IConn
}

// 初始化
func NewConnManager() network.IConnManager {
	return &ConnManager{
		Connects: make(map[string]map[string]network.IConn),
	}
}

// 添加连接句柄
func (this *ConnManager) Add(typeId string, conn network.IConn) {
	this.Lock()
	defer this.Unlock()
	if this.Connects[typeId] != nil {
		this.Connects[typeId][conn.GetGuid()] = conn
	} else {
		conns := make(map[string]network.IConn)
		conns[conn.GetGuid()] = conn
		this.Connects[typeId] = conns
	}
}

// 移除连接句柄
func (this *ConnManager) Remove(typeId string, guid string) {
	this.Lock()
	defer this.Unlock()
	if this.Connects[typeId] != nil {
		delete(this.Connects[typeId], guid)
	}
}

// 根据guid获取连接句柄
func (this *ConnManager) GetConn(typeId string, guid string) network.IConn {
	if this.Connects[typeId] != nil {
		return this.Connects[typeId][guid]
	}
	return nil
}

// 根据type获取连接列表
func (this *ConnManager) GetTypeConnMap(typeId string) map[string]network.IConn {
	return this.Connects[typeId]
}

// 获取连接数量
func (this *ConnManager) Len() int {
	return len(this.Connects)
}

// 清除并停止所有连接
func (this *ConnManager) ClearConn() {
	//保护共享资源Map 加写锁
	this.Lock()
	defer this.Unlock()
	//停止并删除全部的连接信息
	for _, conns := range this.Connects {
		//停止
		for connID, conn := range conns {
			conn.Stop()
			//删除
			delete(this.Connects, connID)
		}
	}
}
