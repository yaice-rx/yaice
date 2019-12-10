package network

import (
	"github.com/golang/protobuf/proto"
)

type IConn interface {
	SendMsg(message proto.Message) error
	Start()
	Stop()
	GetConn() interface{}
	GetGuid() string
}
