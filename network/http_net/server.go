package http_net

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/logger"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// HTTPServer HTTP服务器实现
type HTTPServer struct {
	sync.RWMutex
	config        *config.HTTPConfig
	server        *http.Server
	isRunning     bool
	listenPort    int
	logger        *logger.Logger
	activeConns   int32
	healthHandler http.HandlerFunc
	// 新增：动态路由管理
	routes     map[string]http.HandlerFunc // path -> handler
	routeMutex sync.RWMutex
	mux        *http.ServeMux
}

// NewHTTPServer 创建HTTP服务器实例
func NewHTTPServer(cfg *config.HTTPConfig, logger *logger.Logger) *HTTPServer {
	return &HTTPServer{
		config:      cfg,
		logger:      logger,
		isRunning:   false,
		activeConns: 0,
	}
}

// Listen 实现IServer接口的Listen方法
func (h *HTTPServer) Listen(startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) int {
	h.Lock()
	defer h.Unlock()

	if h.isRunning {
		h.logger.Warn("HTTP server is already running")
		return -1
	}

	// 创建HTTP服务器
	mux := http.NewServeMux()
	h.setupRoutes(mux)

	h.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", h.config.Host, h.config.Port),
		Handler:        h.applyMiddleware(mux),
		ReadTimeout:    h.config.ReadTimeout,
		WriteTimeout:   h.config.WriteTimeout,
		IdleTimeout:    h.config.IdleTimeout,
		MaxHeaderBytes: h.config.MaxHeaderBytes,
	}

	// 尝试在指定端口范围内启动
	for port := startPort; port <= endPort; port++ {
		h.server.Addr = fmt.Sprintf("%s:%d", h.config.Host, port)
		listener, err := h.createListener()
		if err != nil {
			h.logger.Debug("Port not available, trying next",
				logger.Int("port", port),
				logger.Error(err))
			continue
		}
		// 成功绑定端口
		h.listenPort = port
		h.isRunning = true
		// 启动服务器
		go h.serve(listener)
		h.logger.Info("HTTP server started successfully",
			logger.String("host", h.config.Host),
			logger.Int("port", port),
			logger.Bool("enable_tls", h.config.EnableTLS))
		return port
	}

	h.logger.Error("No available ports in the specified range",
		logger.Int("start_port", startPort),
		logger.Int("end_port", endPort))
	return -1
}

// Close 实现IServer接口的Close方法
func (h *HTTPServer) Close(ctx context.Context) error {
	h.Lock()
	defer h.Unlock()
	if !h.isRunning {
		return nil
	}
	h.logger.Info("Shutting down HTTP server")
	// 创建关闭上下文，设置超时
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// 优雅关闭服务器
	if err := h.server.Shutdown(shutdownCtx); err != nil {
		h.logger.Error("HTTP server shutdown error", logger.Error(err))
		// 强制关闭
		if closeErr := h.server.Close(); closeErr != nil {
			h.logger.Error("HTTP server force close error", logger.Error(closeErr))
			return fmt.Errorf("shutdown error: %w, close error: %v", err, closeErr)
		}
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}
	h.isRunning = false
	h.logger.Info("HTTP server stopped successfully")
	return nil
}

// GetActiveConnections 获取活跃连接数
func (h *HTTPServer) GetActiveConnections() int {
	h.RLock()
	defer h.RUnlock()
	return int(h.activeConns)
}

// Health 健康检查
func (h *HTTPServer) Health() error {
	h.RLock()
	defer h.RUnlock()

	if !h.isRunning {
		return fmt.Errorf("HTTP server is not running")
	}
	// 简单的健康检查：尝试连接到服务器
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/health", h.server.Addr))
	if err != nil {
		return fmt.Errorf("HTTP server health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP server health check returned status: %d", resp.StatusCode)
	}

	return nil
}

// RegisterRoute 注册动态路由
func (h *HTTPServer) RegisterRoute(path string, method string, handler http.HandlerFunc) error {
	h.routeMutex.Lock()
	defer h.routeMutex.Unlock()

	if path == "" || handler == nil {
		return fmt.Errorf("invalid route parameters")
	}

	// 标准化路径
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 存储路由
	h.routes[path] = handler

	// 如果服务器正在运行，动态添加路由
	if h.isRunning && h.mux != nil {
		h.mux.HandleFunc(path, handler)
		h.logger.Info("Dynamic route registered",
			logger.String("path", path),
			logger.String("method", method))
	}

	return nil
}

// 创建监听器
func (h *HTTPServer) createListener() (net.Listener, error) {
	addr := h.server.Addr
	if h.config.EnableTLS {
		// TLS监听
		return tls.Listen("tcp", addr, h.getTLSConfig())
	}
	// 普通HTTP监听
	return net.Listen("tcp", addr)
}

// 获取TLS配置
func (h *HTTPServer) getTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
}

// 启动服务
func (h *HTTPServer) serve(listener net.Listener) {
	defer func() {
		h.Lock()
		h.isRunning = false
		h.Unlock()

		if r := recover(); r != nil {
			h.logger.Error("HTTP server panic recovered",
				logger.Any("recover", r))
		}
	}()

	var err error
	if h.config.EnableTLS {
		err = h.server.ServeTLS(listener, h.config.CertFile, h.config.KeyFile)
	} else {
		err = h.server.Serve(listener)
	}

	if err != nil && err != http.ErrServerClosed {
		h.logger.Error("HTTP server serve error", logger.Error(err))
	}
}

// 设置路由
func (h *HTTPServer) setupRoutes(mux *http.ServeMux) {
	// 健康检查端点
	mux.HandleFunc("/health", h.healthHandler)
	// pprof端点（如果启用）
	if h.config.EnablePprof {
		h.setupPprofRoutes(mux)
	}
	// 默认路由
	mux.HandleFunc("/", h.defaultHandler)
}

// 应用中间件
func (h *HTTPServer) applyMiddleware(handler http.Handler) http.Handler {
	// CORS中间件
	if h.config.EnableCORS {
		handler = h.corsMiddleware(handler)
	}
	// Gzip压缩中间件
	if h.config.EnableGzip {
		handler = h.gzipMiddleware(handler)
	}
	// 连接计数中间件
	handler = h.connectionCounterMiddleware(handler)
	// 日志中间件
	handler = h.loggingMiddleware(handler)
	return handler
}

// 默认处理器
func (h *HTTPServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Welcome to YAICE HTTP Server", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
}

// CORS中间件
func (h *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		origin := r.Header.Get("Origin")
		if len(h.config.CORSOrigins) == 0 || contains(h.config.CORSOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", h.config.CORSOrigins[0])
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Gzip压缩中间件
func (h *HTTPServer) gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查客户端是否支持gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// 创建gzip响应写入器
			gz := gzip.NewWriter(w)
			defer gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 连接计数中间件
func (h *HTTPServer) connectionCounterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&h.activeConns, 1)
		defer atomic.AddInt32(&h.activeConns, -1)

		next.ServeHTTP(w, r)
	})
}

// 日志中间件
func (h *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装ResponseWriter以捕获状态码
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		h.logger.Info("HTTP request",
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.String("remote_addr", r.RemoteAddr),
			logger.Int("status", wrapped.statusCode),
			logger.Duration("duration", duration),
			logger.String("user_agent", r.UserAgent()))
	})
}

// 设置pprof路由
func (h *HTTPServer) setupPprofRoutes(mux *http.ServeMux) {
	prefix := h.config.PprofPrefix
	if prefix == "" {
		prefix = "/debug/pprof"
	}

	mux.HandleFunc(prefix+"/", pprof.Index)
	mux.HandleFunc(prefix+"/cmdline", pprof.Cmdline)
	mux.HandleFunc(prefix+"/profile", pprof.Profile)
	mux.HandleFunc(prefix+"/symbol", pprof.Symbol)
	mux.HandleFunc(prefix+"/trace", pprof.Trace)
}

// 辅助函数：检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Gzip响应写入器包装器
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}

// 响应写入器包装器（用于捕获状态码）
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// GetConnCount 获取连接计数（实现IServer接口）
func (h *HTTPServer) GetConnCount() int32 {
	return atomic.LoadInt32(&h.activeConns)
}

// SetMaxConnCount 设置最大连接数（实现IServer接口）
func (h *HTTPServer) SetMaxConnCount(count int32) {
	// HTTP服务器通常不需要设置最大连接数限制
	// 由HTTP服务器本身管理连接池
	h.logger.Debug("SetMaxConnCount called for HTTP server",
		logger.Int32("max_conn", count))
}
