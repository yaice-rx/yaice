// config/config.go
package config

import (
	"errors"
	"sync"
	"time"
)

// Config 全局配置结构
type Config struct {
	App            *AppConfig      `yaml:"app"`
	Log            *LogConfig      `yaml:"logger"`
	Network        *NetworkConfig  `yaml:"network"`
	Database       *DatabaseConfig `yaml:"database"`
	GlobalMQConfig *GlobalMQConfig `yaml:"global_mq"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Env     string `yaml:"env"`
	Debug   bool   `yaml:"debug"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	TCP  *TCPConfig  `yaml:"tcp_net"`
	HTTP *HTTPConfig `yaml:"http_net"`
}

// HTTPConfig HTTP配置
type HTTPConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`     // 读取超时
	WriteTimeout   time.Duration `yaml:"write_timeout"`    // 写入超时
	IdleTimeout    time.Duration `yaml:"idle_timeout"`     // 空闲超时
	MaxHeaderBytes int           `yaml:"max_header_bytes"` // 最大头部字节数
	EnableTLS      bool          `yaml:"enable_tls"`       // 是否启用TLS
	CertFile       string        `yaml:"cert_file"`        // 证书文件路径
	KeyFile        string        `yaml:"key_file"`         // 私钥文件路径
	EnableCORS     bool          `yaml:"enable_cors"`      // 是否启用CORS
	CORSOrigins    []string      `yaml:"cors_origins"`     // CORS允许的源
	EnableGzip     bool          `yaml:"enable_gzip"`      // 是否启用Gzip压缩
	GzipLevel      int           `yaml:"gzip_level"`       // Gzip压缩级别
	EnableMetrics  bool          `yaml:"enable_metrics"`   // 是否启用指标收集
	MetricsPath    string        `yaml:"metrics_path"`     // 指标路径
	EnablePprof    bool          `yaml:"enable_pprof"`     // 是否启用pprof
	PprofPrefix    string        `yaml:"pprof_prefix"`     // pprof路径前缀
}

// TCPConfig TCP配置
type TCPConfig struct {
	Enabled      bool          `yaml:"enabled"`
	Host         string        `yaml:"host"`
	StartPort    int           `yaml:"start_port"`
	EndPort      int           `yaml:"end_port"`
	Port         int           `yaml:"port"`
	MaxConn      int           `yaml:"max_conn"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MongoDB *MongoDBConfig `yaml:"mongodb"`
}

// GlobalMQConfig 全局MQ配置
type GlobalMQConfig struct {
	Enabled    bool `yaml:"enabled"`
	MaxWorkers int  `yaml:"max_workers"`
	FrameRate  int  `yaml:"frame_rate"`
	QueueSize  int  `yaml:"queue_size"`
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	Enabled     bool          `yaml:"enabled"`
	URI         string        `yaml:"uri"`
	Database    string        `yaml:"database"`
	MaxPoolSize uint64        `yaml:"max_pool_size"`
	MinPoolSize uint64        `yaml:"min_pool_size"`
	Timeout     time.Duration `yaml:"timeout"`
}

// 配置管理器
type ConfigManager struct {
	config *Config
	mu     sync.RWMutex
}

var (
	globalConfig *ConfigManager
	once         sync.Once
)

// Init 初始化配置
func Init(config *Config) error {
	once.Do(func() {
		globalConfig = &ConfigManager{config: config}
	})
	return globalConfig.validate()
}

// Get 获取配置
func Get() *Config {
	if globalConfig == nil {
		return &Config{}
	}
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.config
}

// 验证配置
func (cm *ConfigManager) validate() error {
	if cm.config == nil {
		return errors.New("config is nil")
	}
	if cm.config.App == nil {
		cm.config.App = &AppConfig{
			Name:    "yaice",
			Version: "1.0.0",
			Env:     "development",
		}
	}
	if cm.config.Log == nil {
		cm.config.Log = &LogConfig{
			Level:    "info",
			Path:     "./logs",
			MaxSize:  100,
			MaxAge:   7,
			Compress: true,
		}
	}
	if err := cm.validateNetworkConfig(); err != nil {
		return err
	}
	if err := cm.validateDatabaseConfig(); err != nil {
		return err
	}
	return nil
}

// validateNetworkConfig 验证网络配置
func (cm *ConfigManager) validateNetworkConfig() error {
	cfg := cm.config.Network
	// 验证TCP配置
	if cfg.TCP != nil && cfg.TCP.Enabled {
		if cfg.TCP.Port <= 0 || cfg.TCP.Port > 65535 {
			return errors.New("tcp_net port must be between 1 and 65535")
		}
		if cfg.TCP.MaxConn <= 0 {
			cfg.TCP.MaxConn = 10000
		}
	}

	// 验证HTTP配置
	if cfg.HTTP != nil && cfg.HTTP.Enabled {
		if cfg.HTTP.Port <= 0 || cfg.HTTP.Port > 65535 {
			return errors.New("http port must be between 1 and 65535")
		}
		if cfg.HTTP.EnableTLS {
			if cfg.HTTP.CertFile == "" || cfg.HTTP.KeyFile == "" {
				return errors.New("cert_file and key_file are required when enable_tls is true")
			}
		}
		if cfg.HTTP.GzipLevel < 0 || cfg.HTTP.GzipLevel > 9 {
			cfg.HTTP.GzipLevel = 6
		}
	}
	return nil
}

// validateDatabaseConfig 验证数据库配置
func (cm *ConfigManager) validateDatabaseConfig() error {
	cfg := cm.config.Database

	// 验证MongoDB配置
	if cfg.MongoDB != nil && cfg.MongoDB.Enabled {
		if cfg.MongoDB.URI == "" {
			return errors.New("mongodb uri is required")
		}
		if cfg.MongoDB.Database == "" {
			return errors.New("mongodb database name is required")
		}
		if cfg.MongoDB.MaxPoolSize == 0 {
			cfg.MongoDB.MaxPoolSize = 100
		}
		if cfg.MongoDB.MinPoolSize == 0 {
			cfg.MongoDB.MinPoolSize = 10
		}
		if cfg.MongoDB.Timeout == 0 {
			cfg.MongoDB.Timeout = 10 * time.Second
		}
	}
	return nil
}
