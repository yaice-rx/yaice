package database

// BaseEntity 所有业务实体的基础结构
type BaseEntity struct {
	ID       int64 `bson:"_id"` // 你的 guid
	GUID     int64 `bson:"guid"`
	ShardKey int32 `bson:"shard_key"` // guid % 48
	Version  int64 `bson:"version"`   // 可选：乐观锁
}

func (b *BaseEntity) CalcShardKey() int32 {
	return int32(b.GUID % 48)
}

type Entity interface {
	CollectionName() string
}
