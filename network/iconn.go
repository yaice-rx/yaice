package network

import "github.com/golang/protobuf/proto"

type IConn interface {
	SendMsg(message proto.Message) error
	Start()
	Stop()
	SetTypeId(typeId string)
	GetTypeId() string
	ResetConnectTime()
	GetNetworkConn() interface{}
	GetGuid() string
}
