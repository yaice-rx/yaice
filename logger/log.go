package logger

import (
	"fmt"
	"github.com/yaice-rx/yaice/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Logger 自定义日志器结构
type Logger struct {
	*zap.Logger
	config *config.LogConfig
	mu     sync.RWMutex
}

var (
	globalLogger *Logger
	once         sync.Once
)

// NewLogger 创建新的日志器
func NewLogger(cfg *config.LogConfig) *Logger {
	if cfg == nil {
		cfg = &config.LogConfig{
			Level:      "info",
			Path:       "./logs",
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     7,
			Compress:   true,
		}
	}
	// 确保日志目录存在
	if err := os.MkdirAll(cfg.Path, 0755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %v", err))
	}
	// 创建日志核心
	core := createLogCore(cfg)
	// 创建日志器
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return &Logger{
		Logger: zapLogger,
		config: cfg,
	}
}

// WithField 添加字段到日志
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.Any(key, value)),
		config: l.config,
	}
}

// WithFields 添加多个字段到日志
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &Logger{
		Logger: l.Logger.With(zapFields...),
		config: l.config,
	}
}

// Sync 同步日志缓冲区
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// createLogCore 创建日志核心
func createLogCore(cfg *config.LogConfig) zapcore.Core {
	// 设置日志级别
	level := getLogLevel(cfg.Level)
	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "function",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	// 创建编码器
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	// 文件写入器
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.Path, "app.log"),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})
	// 控制台写入器
	consoleWriter := zapcore.Lock(os.Stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	// 创建核心
	cores := []zapcore.Core{
		zapcore.NewCore(encoder, fileWriter, level),
	}
	// 开发环境添加控制台输出
	if cfg.Level == "debug" {
		cores = append(cores, zapcore.NewCore(consoleEncoder, consoleWriter, level))
	}
	return zapcore.NewTee(cores...)
}
func getLogger() *Logger {
	once.Do(func() {
		globalLogger = NewLogger(nil)
	})
	return globalLogger
}

// getLogLevel 获取日志级别
func getLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// 便捷方法
func Debug(msg string, fields ...zap.Field) {
	getLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	getLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	getLogger().Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	getLogger().Fatal(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	getLogger().Panic(msg, fields...)
}

// 便捷字段创建方法
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Int32(key string, val int32) zap.Field {
	return zap.Int32(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func Uint64(key string, val uint64) zap.Field {
	return zap.Uint64(key, val)
}

func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}
