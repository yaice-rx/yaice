package database

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ShardExecutor struct {
	collection *mongo.Collection
	queue      chan *WriteEvent
	wg         sync.WaitGroup
	stopCh     chan struct{}
	batchSize  int
}

func NewShardExecutor(
	col *mongo.Collection,
	batchSize int,
	queueSize int,
) *ShardExecutor {
	exec := &ShardExecutor{
		collection: col,
		queue:      make(chan *WriteEvent, queueSize),
		stopCh:     make(chan struct{}),
		batchSize:  batchSize,
	}
	exec.wg.Add(1)
	go exec.loop()
	return exec
}

func (s *ShardExecutor) Submit(evt *WriteEvent) {
	s.queue <- evt
}

func (s *ShardExecutor) loop() {
	defer s.wg.Done()

	buffer := make([]mongo.WriteModel, 0, s.batchSize)
	events := make([]*WriteEvent, 0, s.batchSize)

	for {
		select {
		case evt, ok := <-s.queue:
			if !ok {
				s.flush(buffer, events)
				return
			}
			buffer = append(buffer, evt.Model)
			events = append(events, evt)

			if len(buffer) >= s.batchSize {
				s.flush(buffer, events)
				buffer = buffer[:0]
				events = events[:0]
			}
		case <-time.After(5 * time.Millisecond):
			if len(buffer) > 0 {
				s.flush(buffer, events)
				buffer = buffer[:0]
				events = events[:0]
			}
		case <-s.stopCh:
			// 收到停止信号，处理剩余事件
			if len(buffer) > 0 {
				s.flush(buffer, events)
			}
			// 关闭队列
			close(s.queue)
			log.Printf("ShardExecutor loop exiting due to stop signal")
			return
		}
	}
}

func (s *ShardExecutor) flush(
	models []mongo.WriteModel,
	events []*WriteEvent,
) {
	_, err := s.collection.BulkWrite(context.Background(), models)
	for _, e := range events {
		if e.Callback != nil {
			e.Callback(err)
		}
	}
}

func (s *ShardExecutor) Shutdown() {
	close(s.queue)
	s.wg.Wait()
}
