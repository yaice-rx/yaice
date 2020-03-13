package network

import "github.com/golang/protobuf/proto"

type IConn interface {
	Start()
	GetGuid() string
	Close()
	Send(message proto.Message) error
	SendByte(message []byte) error
	GetConn()interface{}
}
