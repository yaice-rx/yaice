package core

import (
	"github.com/yaice-rx/yaice/packates"
)

// Connection 连接接口（核心定义）
type Connection interface {
	// 基本信息
	GetGuid() uint64
	GetRemoteAddr() string
	GetLocalAddr() string
	IsClosed() bool

	// 连接操作
	Send(data []byte) error
	Close() error
	Start() error

	// 数据包处理
	GetPacket() packates.IPacket

	// 活动时间
	GetLastActivity() int64
	UpdateLastActivity()

	// 服务类型
	GetServeType() ServeType
}

// ServeType 服务类型枚举
type ServeType int

const (
	Serve_Unknown ServeType = iota
	Serve_Server
	Serve_Client
)

// ConnectionInfo 连接信息（用于消息传递）
type ConnectionInfo struct {
	Guid       uint64
	RemoteAddr string
	LocalAddr  string
	IsClosed   bool
	ServeType  ServeType
}

// NewConnectionInfo 从Connection创建连接信息
func NewConnectionInfo(conn Connection) ConnectionInfo {
	return ConnectionInfo{
		Guid:       conn.GetGuid(),
		RemoteAddr: conn.GetRemoteAddr(),
		LocalAddr:  conn.GetLocalAddr(),
		IsClosed:   conn.IsClosed(),
		ServeType:  conn.GetServeType(),
	}
}
