package network

import (
	"github.com/golang/protobuf/proto"
)

type IConn interface {
	SendMsg(message proto.Message) error
	Start()
	Stop()
	GetNetworkConn() interface{}
	GetConnectType() string
	GetGuid() string
}
