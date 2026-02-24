package core

import (
	"context"
)

// Server 服务器接口（核心定义）
type Server interface {
	// 服务器操作
	Listen(startPort, endPort int, allowFunc func(interface{}) bool) int
	Close(ctx context.Context) error
	Health() error
	// 连接管理
	GetActiveConnections() int
	GetConnCount() int32
	SetMaxConnCount(count int32)
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Enabled   bool
	Host      string
	StartPort int
	EndPort   int
	MaxConn   int
}
