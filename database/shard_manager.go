package database

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ShardManager struct {
	executors map[int32]*ShardExecutor
}

func NewShardManager(
	db *mongo.Database,
	collection string,
	shardCount int,
	batchSize int,
	queueSize int,
) *ShardManager {
	mgr := &ShardManager{
		executors: make(map[int32]*ShardExecutor, shardCount),
	}

	for i := 0; i < shardCount; i++ {
		col := db.Collection(collection)
		mgr.executors[int32(i)] =
			NewShardExecutor(col, batchSize, queueSize)
	}
	return mgr
}

func (m *ShardManager) Submit(
	shardKey int32,
	evt *WriteEvent,
) {
	exec, ok := m.executors[shardKey]
	if !ok {
		panic(fmt.Sprintf("no shard executor %d", shardKey))
	}
	exec.Submit(evt)
}

func (m *ShardManager) Shutdown() {
	for _, e := range m.executors {
		e.Shutdown()
	}
}
