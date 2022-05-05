package network

import "google.golang.org/protobuf/proto"

type IConn interface {
	GetGuid() uint64
	Close()
	Start()
	Send(message proto.Message) error
	SendByte(message []byte) error
	GetConn() interface{}
	GetCreateTime() int64
	GetOptions() IOptions
}
