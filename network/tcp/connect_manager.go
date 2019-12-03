package tcp

import (
	"sync"
	"yaice/network"
)

type ConnManager struct {
	sync.RWMutex
	//map[guid]连接句柄
	Connects map[string]network.IConnect
}

/**
 * 初始化
 */
func NewConnManager() network.IConnectList {
	return &ConnManager{
		Connects: make(map[string]network.IConnect),
	}
}

/**
 * 添加连接句柄
 */
func (this *ConnManager) Add(conn network.IConnect) {
	this.Lock()
	defer this.Unlock()
	this.Connects[conn.GetGuid()] = conn
}

/**
 * 移除连接句柄
 */
func (this *ConnManager) Remove(conn network.IConnect) {
	this.Lock()
	defer this.Unlock()
	delete(this.Connects, conn.GetGuid())
}

/**
 * 根据guid获取连接句柄
 */
func (this *ConnManager) Get(guid string) network.IConnect {
	return this.Connects[guid]
}

/**
 * 获取连接数量
 */
func (this *ConnManager) Len() int {
	return len(this.Connects)
}

/**
 * 清除并停止所有连接
 */
func (this *ConnManager) ClearConn() {
	//保护共享资源Map 加写锁
	this.Lock()
	defer this.Unlock()
	//停止并删除全部的连接信息
	for connID, conn := range this.Connects {
		//停止
		conn.Stop()
		//删除
		delete(this.Connects, connID)
	}
}
