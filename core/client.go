package core

// Client 客户端接口（核心定义）
type Client interface {
	// 客户端操作
	Connect(addr string) error
	Close() error
	Reconnect() error

	// 消息发送
	Send(data []byte) error
	SendAsync(data []byte) error

	// 状态查询
	IsConnected() bool
	GetServerAddr() string
}

// ClientConfig 客户端配置
type ClientConfig struct {
	Enabled    bool
	ServerAddr string
	Reconnect  bool
	MaxRetries int
	RetryDelay int
}
