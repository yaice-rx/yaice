package service

import (
	"context"
	"fmt"
	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/core"
	"github.com/yaice-rx/yaice/logger"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/packates"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

// Service 服务结构体
type Service struct {
	name       string
	version    string
	state      State
	stateMutex sync.RWMutex
	startTime  time.Time
	uptime     atomic.Int64
	config     *config.Config
	logger     *logger.Logger
	networkMgr *network.Manager
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	packet     packates.IPacket

	// GlobalMQ配置
	globalMQConfig *config.GlobalMQConfig
	globalMQ       *core.GlobalMQ
}

// NewService 创建服务实例
func NewService(name, version string, cfg *config.Config, packet packates.IPacket) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		name:           name,
		version:        version,
		state:          StateStopped,
		startTime:      time.Now(),
		config:         cfg,
		logger:         logger.NewLogger(cfg.Log),
		ctx:            ctx,
		cancel:         cancel,
		packet:         packet,
		globalMQConfig: cfg.GlobalMQConfig,
	}
}

// Start 启动服务
func (s *Service) Start() error {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	if s.state != StateStopped {
		return fmt.Errorf("service is already running")
	}

	s.state = StateStarting
	s.logger.Info("Starting service...",
		zap.String("name", s.name),
		zap.String("version", s.version))

	// 初始化GlobalMQ
	if err := s.initGlobalMQ(); err != nil {
		return fmt.Errorf("failed to initialize GlobalMQ: %w", err)
	}

	// 初始化网络管理器
	if err := s.initNetworkManager(); err != nil {
		return fmt.Errorf("failed to initialize network manager: %w", err)
	}

	s.state = StateRunning
	s.startTime = time.Now()

	s.logger.Info("Service started successfully",
		zap.String("name", s.name),
		zap.String("version", s.version),
		zap.Time("start_time", s.startTime))

	return nil
}

// initGlobalMQ 初始化GlobalMQ
func (s *Service) initGlobalMQ() error {
	if s.globalMQConfig == nil || !s.globalMQConfig.Enabled {
		s.logger.Info("GlobalMQ is not enabled")
		return nil
	}

	s.globalMQ = core.GetGlobalMQWithConfig(
		s.ctx,
		int32(s.globalMQConfig.FrameRate),
		int32(s.globalMQConfig.MaxWorkers),
		s.globalMQConfig.QueueSize,
	)

	// 启动GlobalMQ worker
	s.globalMQ.StartWorker(s.globalMQConfig.MaxWorkers)

	s.logger.Info("GlobalMQ initialized",
		zap.Int32("frame_rate", int32(s.globalMQConfig.FrameRate)),
		zap.Int32("max_workers", int32(s.globalMQConfig.MaxWorkers)),
		zap.Int("queue_size", s.globalMQConfig.QueueSize))

	return nil
}

// initNetworkManager 初始化网络管理器
func (s *Service) initNetworkManager() error {
	s.networkMgr = network.NewManager(s.ctx, s.config.Network, s.logger, s.packet, s.globalMQ.GetGlobalMQChannel())
	// 启动网络管理器
	if err := s.networkMgr.Start(); err != nil {
		return fmt.Errorf("failed to start network manager: %w", err)
	}
	s.logger.Info("Network manager started successfully")
	return nil
}

// RegisterGlobalMQHandler 注册GlobalMQ消息处理器
func (s *Service) RegisterGlobalMQHandler(msgID int32, handler core.MessageHandler) error {
	if s.globalMQ == nil {
		return fmt.Errorf("GlobalMQ is not initialized")
	}

	s.globalMQ.RegisterHandler(msgID, handler)
	s.logger.Debug("GlobalMQ handler registered",
		zap.Int32("msg_id", msgID))

	return nil
}

// RegisterGlobalMQHandlerFunc 注册GlobalMQ消息处理器函数
func (s *Service) RegisterGlobalMQHandlerFunc(msgID int32, handlerFunc core.HandlerFunc) error {
	return s.RegisterGlobalMQHandler(msgID, handlerFunc)
}

// GetGlobalMQ 获取GlobalMQ实例
func (s *Service) GetGlobalMQ() *core.GlobalMQ {
	return s.globalMQ
}

// Stop 停止服务
func (s *Service) Stop() error {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	if s.state != StateRunning {
		return fmt.Errorf("service is not running")
	}

	s.state = StateStopping
	s.logger.Info("Stopping service...")

	// 停止网络管理器
	if s.networkMgr != nil {
		if err := s.networkMgr.Stop(s.ctx); err != nil {
			s.logger.Error("Failed to stop network manager", zap.Error(err))
		}
	}
	// 取消上下文
	s.cancel()
	// 等待所有goroutine结束
	s.wg.Wait()
	s.state = StateStopped
	s.uptime.Store(int64(time.Since(s.startTime).Seconds()))

	s.logger.Info("Service stopped successfully",
		zap.Duration("uptime", time.Since(s.startTime)))

	return nil
}

// State 服务状态枚举
type State int

const (
	StateStopped State = iota
	StateStarting
	StateRunning
	StateStopping
	StateError
)
