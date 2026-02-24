package tcp_net

import (
	"context"
	"errors"
	"fmt"
	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/core"
	"github.com/yaice-rx/yaice/logger"
	"github.com/yaice-rx/yaice/packates"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	sync.RWMutex
	config          *config.TCPConfig
	type_           core.ServeType
	connCount       int32
	maxConnCount    int32
	listener        *net.TCPListener
	cancel          context.CancelFunc
	ctx             context.Context
	activeConns     map[uint64]core.Connection
	packet          packates.IPacket
	isRunning       bool
	isAllowConnFunc func(conn interface{}) bool
	logger          *logger.Logger
	globalMQ        chan<- core.Message
}

func NewServer(ctx context.Context, packet packates.IPacket, globalMQ chan<- core.Message, logger *logger.Logger) core.Server {
	s := &Server{
		type_:        core.Serve_Server,
		connCount:    0,
		maxConnCount: 10000,
		activeConns:  make(map[uint64]core.Connection),
		ctx:          ctx,
		isRunning:    false,
		packet:       packet,
		logger:       logger,
		globalMQ:     globalMQ,
	}
	return s
}

func (s *Server) Listen(startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) int {
	if s.isRunning {
		return -1 // 服务器已在运行
	}
	s.isAllowConnFunc = isAllowConnFunc
	s.isRunning = true
	// 彻底优化：不再为每个端口创建goroutine，改为顺序尝试
	for port := startPort; port < endPort; port++ {
		tcpAddr, err := net.ResolveTCPAddr("tcp_net", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}
		listener, err := net.ListenTCP("tcp_net", tcpAddr)
		if err != nil {
			continue
		}
		// 成功监听，保存监听器
		s.listener = listener
		// 启动监听goroutine处理连接
		go s.acceptLoop()
		return port
	}
	// 没有找到可用端口
	s.isRunning = false
	return -1
}

// acceptLoop 接受连接的循环
func (s *Server) acceptLoop() {
	defer func() {
		s.isRunning = false
		if r := recover(); r != nil {
			logger.Error(fmt.Errorf("TCP Server acceptLoop panic: %v", r))
		}
	}()
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// 设置接受超时，定期检查ctx是否已取消
			s.listener.SetDeadline(time.Now().Add(100 * time.Millisecond))
			tcpConn, err := s.listener.AcceptTCP()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// 超时错误，继续下一次循环
					continue
				}
				// 其他错误，关闭服务器
				logger.Error(fmt.Errorf("TCP Server accept error: %v", err))
				return
			}
			// 检查连接数限制
			if atomic.LoadInt32(&s.connCount) >= s.maxConnCount {
				tcpConn.Close()
				continue
			}
			// 检查是否允许连接
			if s.isAllowConnFunc != nil {
				if !s.isAllowConnFunc(tcpConn) {
					tcpConn.Close()
					continue
				}
			}
			// 创建连接
			conn := NewConn(s.ctx, tcpConn, s.packet, nil, core.Serve_Server, s.globalMQ)
			// 保存连接到活跃连接列表
			s.addConn(conn)
			// 启动连接处理
			go func() {
				defer func() {
					// 连接处理结束后从活跃连接列表移除
					s.removeConn(conn.GetGuid())
					if r := recover(); r != nil {
						logger.Error(fmt.Errorf("TCP Conn panic: %v", r))
					}
				}()
				conn.Start()
			}()
		}
	}
}

func (s *Server) Close(ctx context.Context) error {
	if !s.isRunning {
		return errors.New("TCP Server is not running")
	}
	// 关闭监听器
	if s.listener != nil {
		s.listener.Close()
	}
	// 关闭所有活跃连接
	s.RLock()
	for _, conn := range s.activeConns {
		conn.Close()
	}
	s.RUnlock()
	s.isRunning = false
	return nil
}

// addConn 添加连接到活跃连接列表
func (s *Server) addConn(conn core.Connection) {
	atomic.AddInt32(&s.connCount, 1)
	s.Lock()
	s.activeConns[conn.GetGuid()] = conn
	s.Unlock()
}

// removeConn 从活跃连接列表移除连接
func (s *Server) removeConn(guid uint64) {
	atomic.AddInt32(&s.connCount, -1)
	s.Lock()
	delete(s.activeConns, guid)
	s.Unlock()
}

// GetConnCount 获取当前连接数
func (s *Server) GetConnCount() int32 {
	return atomic.LoadInt32(&s.connCount)
}

// SetMaxConnCount 设置最大连接数
func (s *Server) SetMaxConnCount(count int32) {
	if count > 0 {
		s.maxConnCount = count
	}
}

func (s *Server) GetActiveConnections() int {
	return int(atomic.LoadInt32(&s.connCount))
}

func (s *Server) Health() error {
	return nil
}
