package database

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"sync"

	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// Manager 数据库管理器
type Manager struct {
	name    string
	config  *config.DatabaseConfig
	logger  *logger.Logger
	mongoDB *mongo.Client
	mu      sync.RWMutex
	isReady bool
}

// NewManager 创建数据库管理器
func NewManager(cfg *config.DatabaseConfig, logger *logger.Logger) *Manager {
	return &Manager{
		name:   "database_manager",
		config: cfg,
		logger: logger,
	}
}

// Start 启动数据库连接
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isReady {
		return nil
	}
	// 连接MongoDB
	if m.config.MongoDB != nil && m.config.MongoDB.Enabled {
		if err := m.connectMongoDB(ctx); err != nil {
			return fmt.Errorf("failed to connect to mongodb: %w", err)
		}
	}
	m.isReady = true
	m.logger.Info("Database manager started successfully")
	return nil
}

// Stop 停止数据库连接
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isReady {
		return nil
	}

	// 断开MongoDB连接
	if m.mongoDB != nil {
		if err := m.mongoDB.Disconnect(ctx); err != nil {
			m.logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		}
		m.mongoDB = nil
	}

	m.isReady = false
	m.logger.Info("Database manager stopped successfully")
	return nil
}

// Health 健康检查
func (m *Manager) Health() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isReady {
		return errors.New("database manager is not ready")
	}

	// 检查MongoDB连接
	if m.mongoDB != nil {
		if err := m.mongoDB.Ping(context.Background(), nil); err != nil {
			return fmt.Errorf("mongodb connection is unhealthy: %w", err)
		}
	}

	return nil
}

// Name 获取组件名称
func (m *Manager) Name() string {
	return m.name
}

// GetMongoDB 获取MongoDB客户端
func (m *Manager) GetMongoDB() *mongo.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mongoDB
}

// MongoDB连接
func (m *Manager) connectMongoDB(ctx context.Context) error {
	mongoCfg := m.config.MongoDB

	client, err := mongo.Connect(options.Client().
		ApplyURI(mongoCfg.URI).
		SetMaxPoolSize(mongoCfg.MaxPoolSize).
		SetMinPoolSize(mongoCfg.MinPoolSize))

	if err != nil {
		return err
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return err
	}

	m.mongoDB = client
	m.logger.Info("MongoDB connected successfully",
		zap.String("database", mongoCfg.Database))

	return nil
}
