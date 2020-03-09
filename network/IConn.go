package network

import "github.com/golang/protobuf/proto"

type IConn interface {
	Start()
	GetGuid() string
	Close()
	GetTimes() int64
	UpdateTime()
	Send(message proto.Message) error
	SendByte(message []byte) error
	SetData(data interface{})
}
