package network

import "github.com/golang/protobuf/proto"

type IConn interface {
	GetGuid() string
	Close()
	Start()
	Send(message proto.Message) error
	SendByte(message []byte) error
	GetConn() interface{}
}
