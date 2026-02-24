package network

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/core"
	"github.com/yaice-rx/yaice/logger"
	"github.com/yaice-rx/yaice/network/http_net"
	"github.com/yaice-rx/yaice/network/tcp_net"
	"github.com/yaice-rx/yaice/packates"
	"sync"
)

// Manager 网络管理器
type Manager struct {
	ctx      context.Context
	name     string
	config   *config.NetworkConfig
	logger   *logger.Logger
	servers  map[string]core.Server
	clients  map[string]core.Client
	mu       sync.RWMutex
	isReady  bool
	packet   packates.IPacket
	globalMQ chan<- core.Message
}

// NewManager 创建网络管理器
func NewManager(ctx context.Context, cfg *config.NetworkConfig, logger *logger.Logger, packet packates.IPacket, globalMQ chan<- core.Message) *Manager {
	return &Manager{
		ctx:      ctx,
		name:     "network_manager",
		config:   cfg,
		logger:   logger,
		servers:  make(map[string]core.Server),
		clients:  make(map[string]core.Client),
		packet:   packet,
		globalMQ: globalMQ,
	}
}

// Start 启动网络服务
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isReady {
		return nil
	}
	// 启动TCP服务器
	if m.config.TCP != nil && m.config.TCP.Enabled {
		if err := m.startTCPServer(m.ctx); err != nil {
			return fmt.Errorf("failed to start TCP server: %w", err)
		}
	}
	// 启动HTTP服务器
	if m.config.HTTP != nil && m.config.HTTP.Enabled {
		if err := m.startHTTPServer(m.ctx); err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}
	// 检查是否至少有一个服务器启动
	if len(m.servers) == 0 {
		m.logger.Warn("No network servers are enabled, network manager will run without serving capabilities")
	}

	m.isReady = true
	m.logger.Info("Network manager started successfully",
		logger.Int("servers_count", len(m.servers)))

	return nil
}

// 启动TCP服务器
func (m *Manager) startTCPServer(ctx context.Context) error {
	tcpCfg := m.config.TCP
	if m.packet == nil {
		m.packet = tcp_net.NewPacket()
	}
	server := tcp_net.NewServer(ctx, m.packet, m.globalMQ, m.logger)
	if port := server.Listen(m.config.TCP.StartPort, m.config.TCP.EndPort, nil); port <= 0 {
		return errors.New(fmt.Sprintf("failed to start TCP server: start port %d end port %d", m.config.TCP.StartPort, m.config.TCP.EndPort))
	}
	m.servers["tcp_net"] = server
	m.logger.Info("TCP server started successfully",
		logger.String("host", tcpCfg.Host),
		logger.Int("port", tcpCfg.Port),
		logger.Int("max_conn", tcpCfg.MaxConn))
	return nil
}

// 启动HTTP服务器
func (m *Manager) startHTTPServer(ctx context.Context) error {
	httpCfg := m.config.HTTP
	server := http_net.NewHTTPServer(httpCfg, m.logger)

	port := server.Listen(httpCfg.Port, httpCfg.Port+10, nil)
	if port <= 0 {
		return fmt.Errorf("failed to start HTTP server on port %d", httpCfg.Port)
	}

	m.servers["http"] = server
	m.logger.Info("HTTP server started successfully",
		logger.String("host", httpCfg.Host),
		logger.Int("port", port),
		logger.Bool("enable_tls", httpCfg.EnableTLS))

	return nil
}

// Stop 停止网络服务
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.isReady {
		return nil
	}
	// 停止所有服务器
	for name, server := range m.servers {
		m.logger.Info("Stopping server", logger.String("name", name))
		if err := server.Close(ctx); err != nil {
			m.logger.Error("Failed to stop server",
				logger.String("name", name), logger.Error(err))
		}
	}
	// 断开所有客户端连接
	for name, client := range m.clients {
		m.logger.Info("Disconnecting client", logger.String("name", name))
		if err := client.Close(); err != nil {
			m.logger.Error("Failed to disconnect client",
				logger.String("name", name), logger.Error(err))
		}
	}
	m.servers = make(map[string]core.Server)
	m.clients = make(map[string]core.Client)
	m.isReady = false

	m.logger.Info("Network manager stopped successfully",
		logger.Int("servers_stopped", len(m.servers)),
		logger.Int("clients_disconnected", len(m.clients)))

	return nil
}

// GetServer 获取指定类型的服务器
func (m *Manager) GetServer(serverType string) core.Server {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.servers[serverType]
}

// GetActiveConnections 获取所有服务器的活跃连接总数
func (m *Manager) GetActiveConnections() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0
	for _, server := range m.servers {
		// 安全地检查服务器是否实现了GetActiveConnections方法
		if connServer, ok := server.(interface{ GetActiveConnections() int }); ok {
			total += connServer.GetActiveConnections()
		}
	}
	return total
}

// IsServerEnabled 检查指定类型的服务器是否启用
func (m *Manager) IsServerEnabled(serverType string) bool {
	switch serverType {
	case "tcp_net":
		return m.config.TCP != nil && m.config.TCP.Enabled
	case "http":
		return m.config.HTTP != nil && m.config.HTTP.Enabled
	default:
		return false
	}
}

// GetServerStatus 获取所有服务器的状态信息
func (m *Manager) GetServerStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]interface{})
	for name, server := range m.servers {
		status[name] = map[string]interface{}{
			"active_connections": server.GetActiveConnections(),
			"health":             server.Health() == nil,
		}
	}
	return status
}

func (m *Manager) Health() error {
	for name, server := range m.servers {
		if err := server.Health(); err != nil {
			return fmt.Errorf("failed to health check server %s: %w", name, err)
		}
	}
	return nil
}

func (m *Manager) Name() string {
	return m.name
}
