package core

import (
	"context"
	"fmt"
	"github.com/yaice-rx/yaice/logger"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

// GlobalMQ 全局消息队列管理器
type GlobalMQ struct {
	ctx         context.Context
	globalQueue chan Message
	handlers    map[int32]MessageHandler
	once        sync.Once
	mu          sync.RWMutex
	isRunning   bool
	startTime   time.Time
	workerCount int32
	maxWorkers  int32

	// 固定帧率控制
	frameRate     int32
	frameInterval time.Duration
	frameTicker   *time.Ticker

	// 性能指标
	msgProcessed  int64
	queueLength   int64
	droppedMsgs   int64
	frameCount    int64
	lastFrameTime time.Time
}

// 全局消息队列实例
var gMQ *GlobalMQ

// GetGlobalMQ 获取全局消息队列管理器实例
func NewGlobalMQ(ctx context.Context, frameRate int32, maxWorkers int32, queueSize int) *GlobalMQ {
	gMQ.once.Do(func() {
		gMQ = &GlobalMQ{
			ctx:         ctx,
			globalQueue: make(chan Message, queueSize),
			handlers:    make(map[int32]MessageHandler),
			workerCount: 1,
			maxWorkers:  maxWorkers,
			isRunning:   true,
			startTime:   time.Now(),
		}
		if frameRate > 0 {
			if err := gMQ.SetFrameRate(frameRate); err != nil {
				logger.Warn("Failed to set frame rate", zap.Int32("frameRate", frameRate), zap.Error(err))
			}
		}
	})
	go gMQ.monitorPerformance()
	return gMQ
}

// GetGlobalMQWithConfig 通过配置获取全局消息队列管理器实例
func GetGlobalMQWithConfig(ctx context.Context, frameRate int32, maxWorkers int32, queueSize int) *GlobalMQ {
	return NewGlobalMQ(ctx, frameRate, maxWorkers, queueSize)
}

// GetGlobalMQChannel 获取全局消息队列通道
func (gmq *GlobalMQ) GetGlobalMQChannel() chan Message {
	return gmq.globalQueue
}

// SendToGlobalQueue 发送消息到全局队列
func (gmq *GlobalMQ) SendToGlobalQueue(conn Connection, msgData []byte, msgId int32) bool {
	connInfo := NewConnectionInfo(conn)
	globalMsg := &GlobalMessage{
		MsgId:      msgId,
		Data:       msgData,
		Connection: connInfo,
	}

	select {
	case gmq.globalQueue <- globalMsg:
		atomic.AddInt64(&gmq.queueLength, 1)
		return true
	default:
		atomic.AddInt64(&gmq.droppedMsgs, 1)
		logger.Warn("GlobalMQ queue full, message dropped",
			zap.Int32("msg_id", msgId),
			zap.Uint64("conn_id", conn.GetGuid()))
		return false
	}
}

// RegisterHandler 注册消息处理器
func (gmq *GlobalMQ) RegisterHandler(msgID int32, handler MessageHandler) {
	gmq.mu.Lock()
	defer gmq.mu.Unlock()
	gmq.handlers[msgID] = handler
	logger.Info("Message handler registered to GlobalMQ", zap.Int32("msg_id", msgID))
}

// RegisterHandlerFunc 注册消息处理器函数
func (gmq *GlobalMQ) RegisterHandlerFunc(msgID int32, handlerFunc HandlerFunc) {
	gmq.RegisterHandler(msgID, handlerFunc)
}

// StartWorker 启动全局消息处理协程
func (gmq *GlobalMQ) StartWorker(workerCount int) {
	if workerCount <= 0 {
		workerCount = int(atomic.LoadInt32(&gmq.workerCount))
	}
	targetCount := int32(workerCount)
	if targetCount > gmq.maxWorkers {
		targetCount = gmq.maxWorkers
	}
	atomic.StoreInt32(&gmq.workerCount, targetCount)

	for i := 0; i < workerCount; i++ {
		gmq.startMonitoredFrameRateProcessor()
	}
	logger.Info("Global message queue workers started with frame rate control",
		zap.Int("count", workerCount),
		zap.Int32("fps", gmq.GetFrameRate()))
}

// processMessagesWithFrameRate 带帧率控制的消息处理
func (gmq *GlobalMQ) processMessagesWithFrameRate() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Errorf("message processing panic recovered: %v", r))
			panic(r)
		}
	}()

	for {
		select {
		case <-gmq.ctx.Done():
			logger.Debug("GlobalMQ worker received context cancellation")
			return
		case <-gmq.frameTicker.C:
			for {
				select {
				case msg, ok := <-gmq.globalQueue:
					if !ok {
						return
					}
					if err := gmq.processMessage(msg); err != nil {
						return
					}
				default:
					break
				}
			}
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

// SetFrameRate 设置帧率控制
func (gmq *GlobalMQ) SetFrameRate(fps int32) error {
	if fps <= 0 || fps > 1000 {
		return fmt.Errorf("invalid frame rate: %d (must be between 1 and 1000)", fps)
	}

	gmq.mu.Lock()
	defer gmq.mu.Unlock()

	gmq.frameRate = fps
	gmq.frameInterval = time.Second / time.Duration(fps)

	gmq.frameTicker = time.NewTicker(gmq.frameInterval)

	logger.Info("GlobalMQ frame rate configured",
		zap.Int32("fps", fps),
		zap.Duration("interval", gmq.frameInterval))

	return nil
}

// GetFrameRate 获取当前帧率
func (gmq *GlobalMQ) GetFrameRate() int32 {
	gmq.mu.RLock()
	defer gmq.mu.RUnlock()
	return gmq.frameRate
}

// GetQueueStats 获取队列统计信息
func (gmq *GlobalMQ) GetQueueStats() map[string]int64 {
	return map[string]int64{
		"processed": atomic.LoadInt64(&gmq.msgProcessed),
		"queued":    atomic.LoadInt64(&gmq.queueLength),
		"dropped":   atomic.LoadInt64(&gmq.droppedMsgs),
		"frames":    atomic.LoadInt64(&gmq.frameCount),
	}
}

// 启动带监控的消息处理协程
func (gmq *GlobalMQ) startMonitoredFrameRateProcessor() {
	go func() {
		for {
			select {
			case <-gmq.ctx.Done():
				return
			default:
				done := make(chan struct{})
				go func() {
					defer close(done)
					gmq.processMessagesWithFrameRate()
				}()

				select {
				case <-done:
					logger.Error(fmt.Errorf("message processing goroutine exited"))
					time.Sleep(time.Second)
				case <-gmq.ctx.Done():
					return
				}
			}
		}
	}()
}

func (gmq *GlobalMQ) processMessage(msg Message) error {
	atomic.AddInt64(&gmq.msgProcessed, 1)
	atomic.AddInt64(&gmq.queueLength, -1)

	gmq.mu.RLock()
	handler, ok := gmq.handlers[msg.GetMsgId()]
	gmq.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no handler registered for message")
	}

	if err := handler.Handle(gmq.ctx, msg); err != nil {
		logger.Error(fmt.Errorf("message processing failed: %w", err))
		return err
	}
	return nil
}

// monitorPerformance 监控性能指标
func (gmq *GlobalMQ) monitorPerformance() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			processed := atomic.LoadInt64(&gmq.msgProcessed)
			frameCount := atomic.LoadInt64(&gmq.frameCount)
			frameRate := gmq.GetFrameRate()

			processingRate := float64(processed) / 5.0
			actualFrameRate := float64(frameCount) / 5.0

			logger.Info("Global message queue performance metrics",
				zap.Float64("processing_rate", processingRate),
				zap.Float64("actual_frame_rate", actualFrameRate),
				zap.Int32("target_frame_rate", frameRate),
				zap.Int64("queue_length", atomic.LoadInt64(&gmq.queueLength)),
				zap.Int64("dropped_messages", atomic.LoadInt64(&gmq.droppedMsgs)))

		case <-gmq.ctx.Done():
			return
		}
	}
}
