package database

import "go.mongodb.org/mongo-driver/v2/mongo"

// WriteEvent 单条写事件
type WriteEvent struct {
	Collection string
	ShardKey   int32
	Model      mongo.WriteModel
	Callback   func(error)
}
