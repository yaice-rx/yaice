package tcp_net

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/yaice-rx/yaice/core"
	"github.com/yaice-rx/yaice/logger"
	"github.com/yaice-rx/yaice/packates"
	"github.com/yaice-rx/yaice/utils"
	"go.uber.org/zap"
	"io"
	"net"
	"sync/atomic"
	"time"
)

// Conn TCP连接实现
type Conn struct {
	type_        core.ServeType
	guid         uint64
	isClosed     int32
	times        int64
	pkg          packates.IPacket
	conn         *net.TCPConn
	serve        interface{}
	data         interface{}
	isPos        int64
	sendQueue    chan []byte
	opt          interface{}
	ctx          context.Context
	lastActivity time.Time
	cancel       context.CancelFunc
	logger       *logger.Logger
	// GlobalMQ集成
	globalMQ chan<- core.Message
	// 消息处理器映射（保留用于向后兼容）
	handlers map[int32]func(conn core.Connection, data packates.ProtocolContentData)
}

// NewConn 创建带GlobalMQ的TCP连接
func NewConn(ctx context.Context, conn *net.TCPConn, pkg packates.IPacket, opt interface{}, type_ core.ServeType,
	globalMQ chan<- core.Message) core.Connection {
	// 内部创建子上下文
	connCtx, cancel := context.WithCancel(ctx)
	// 优化TCP连接参数
	conn.SetNoDelay(true)
	conn.SetReadBuffer(65535)
	conn.SetWriteBuffer(65535)
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(30 * time.Second)
	conn_ := &Conn{
		type_:        type_,
		guid:         utils.GenSnowflakeToo(),
		conn:         conn,
		pkg:          pkg,
		sendQueue:    make(chan []byte, 5000),
		times:        time.Now().Unix(),
		isClosed:     0,
		opt:          opt,
		ctx:          connCtx, // 内部创建的子上下文
		cancel:       cancel,  // 内部创建的取消函数
		lastActivity: time.Now(),
		globalMQ:     globalMQ,
		handlers:     make(map[int32]func(conn core.Connection, data packates.ProtocolContentData)),
	}

	// 批量发送数据
	go conn_.batchWriteLoop()
	return conn_
}

// 实现core.Connection接口的方法
func (c *Conn) GetGuid() uint64 {
	return c.guid
}

func (c *Conn) GetRemoteAddr() string {
	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return ""
}

func (c *Conn) GetLocalAddr() string {
	if c.conn != nil {
		return c.conn.LocalAddr().String()
	}
	return ""
}

func (c *Conn) IsClosed() bool {
	return atomic.LoadInt32(&c.isClosed) == 1
}

func (c *Conn) Send(data []byte) error {
	// 实现发送逻辑
	select {
	case c.sendQueue <- data:
		return nil
	default:
		return errors.New("send queue is full")
	}
}

func (c *Conn) Close() error {
	// 先取消上下文，通知所有协程退出
	if c.cancel != nil {
		c.cancel()
	}

	atomic.StoreInt32(&c.isClosed, 1)
	if c.conn != nil {
		err := c.conn.Close()
		logger.Debug("TCP connection closed",
			zap.Uint64("guid", c.guid),
			zap.String("remote_addr", c.GetRemoteAddr()),
			zap.Error(err))
		return err
	}
	return nil
}

func (c *Conn) Start() error {
	// 启动连接处理逻辑
	go c.readLoop()
	return nil
}

func (c *Conn) GetPacket() packates.IPacket {
	return c.pkg
}

func (c *Conn) GetLastActivity() int64 {
	return c.lastActivity.Unix()
}

func (c *Conn) UpdateLastActivity() {
	c.lastActivity = time.Now()
}

func (c *Conn) GetServeType() core.ServeType {
	return c.type_
}

func (c *Conn) batchWriteLoop() {
	const batchSize = 8192                     // 增大批量写入大小
	const flushInterval = 5 * time.Millisecond // 减小最大等待时间
	// 优化：使用bytes.Buffer减少内存分配
	batch := bytes.NewBuffer(make([]byte, 0, batchSize))
	var timer *time.Timer
	var flushChan <-chan time.Time
	resetTimer := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.NewTimer(flushInterval)
		flushChan = timer.C
	}
	resetTimer()
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	for {
		select {
		case <-c.ctx.Done():
			// 发送剩余数据
			if batch.Len() > 0 {
				c.conn.Write(batch.Bytes())
			}
			return
		case data, ok := <-c.sendQueue:
			if !ok {
				// 发送队列关闭
				if batch.Len() > 0 {
					c.conn.Write(batch.Bytes())
				}
				return
			}
			// 更新最后活动时间
			c.lastActivity = time.Now()
			// 检查是否需要立即发送
			if batch.Len()+len(data) >= batchSize {
				// 先发送已有数据
				if batch.Len() > 0 {
					if _, err := c.conn.Write(batch.Bytes()); err != nil {
						logger.Error(fmt.Errorf("failed to write data,guid: %d, err: %w", c.guid, err))
						c.Close()
						return
					}
					batch.Reset()
				}
				// 发送新数据
				if _, err := c.conn.Write(data); err != nil {
					logger.Error(fmt.Errorf("failed to write data,guid: %d, err: %w", c.guid, err))
					c.Close()
					return
				}
				resetTimer()
			} else {
				// 添加到批量
				batch.Write(data)
			}
		case <-flushChan:
			// 超时，发送批量数据
			if batch.Len() > 0 {
				if _, err := c.conn.Write(batch.Bytes()); err != nil {
					logger.Error(fmt.Errorf("failed to write batch data,guid: %d, err: %w", c.guid, err))
					c.Close()
					return
				}
				batch.Reset()
			}
			resetTimer()
		}
	}
}

func (c *Conn) readLoop() {
	// 预分配读取缓冲区减少内存分配
	headSize := c.pkg.GetHeadLen()
	headData := make([]byte, headSize)
	// 预分配消息体缓冲区
	msgBuffer := make([]byte, 0, 65536)
	for {
		select {
		case <-c.ctx.Done():
			c.Close()
			return
		default:
			// 检查连接是否已关闭
			if c.IsClosed() {
				return
			}
			// 更新最后活动时间
			c.lastActivity = time.Now()
			// 设置合理的读取超时
			if err := c.conn.SetReadDeadline(time.Now().Add(5 * time.Minute)); err != nil {
				c.logger.Error("Failed to set read deadline", zap.Uint64("guid", c.guid), zap.Error(err))
				c.Close()
				return
			}
			// 读取消息头
			n, err := io.ReadFull(c.conn, headData)
			if err != nil {
				c.handleReadError(err)
				return
			}
			if n != int(headSize) {
				c.logger.Error("Incomplete header read", zap.Uint64("guid", c.guid), zap.Int("expected", int(headSize)), zap.Int("actual", n))
				c.Close()
				return
			}

			// 解析消息长度
			msgLen := utils.BytesToInt(headData)
			if msgLen <= 0 || msgLen > 1048576 { // 限制最大消息大小为1MB
				c.logger.Error("Invalid message length", zap.Uint64("guid", c.guid), zap.Int("length", int(msgLen)))
				c.Close()
				return
			}
			// 调整消息体缓冲区大小
			if len(msgBuffer) < int(msgLen) {
				msgBuffer = make([]byte, msgLen)
			} else {
				msgBuffer = msgBuffer[:msgLen]
			}
			// 读取消息体
			n, err = io.ReadFull(c.conn, msgBuffer)
			if err != nil {
				c.handleReadError(err)
				return
			}
			if n != int(msgLen) {
				c.logger.Error("Incomplete message read", zap.Uint64("guid", c.guid), zap.Int("expected", int(msgLen)), zap.Int("actual", n))
				c.Close()
				return
			}
			// 解压网络数据包
			msgData, err, func_ := c.pkg.Unpack(msgBuffer)
			if err != nil {
				c.logger.Error("Failed to unpack message", zap.Uint64("guid", c.guid), zap.Error(err))
				continue
			}
			if msgData == nil {
				c.logger.Error("Unpack returned nil message", zap.Uint64("guid", c.guid))
				continue
			}
			if func_ != nil {
				func_(c)
			}
			c.isPos = msgData.GetIsPos()

			// 写入接收队列
			select {
			case c.globalMQ <- &core.GlobalMessage{
				MsgId: msgData.GetMsgId(),
				Data:  msgData.GetData(),
			}:
			default:
				// 接收队列已满，关闭连接以防止内存泄漏
				c.logger.Warn("Receive queue full, closing connection", zap.Uint64("guid", c.guid))
				c.Close()
				return
			}
		}
	}
}

func (c *Conn) handleReadError(err error) {
	if err == io.EOF {
		c.logger.Info("Connection closed by peer", zap.Uint64("guid", c.guid))
	} else {
		c.logger.Error("Error reading from connection", zap.Uint64("guid", c.guid), zap.Error(err))
	}
	c.Close()
}
