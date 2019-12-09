package network

import (
	"github.com/golang/protobuf/proto"
)

type IConn interface {
	Send(message proto.Message) error
	Stop()
	GetConn() interface{}
	GetGuid() string
}
