package network

import "github.com/golang/protobuf/proto"

type IConn interface {
	GetGuid() uint64
	Close()
	Start()
	Send(message proto.Message) error
	SendByte(message []byte) error
	GetConn() interface{}
}
