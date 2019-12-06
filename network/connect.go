package network

import "github.com/golang/protobuf/proto"

type IConnect interface {
	Send(message proto.Message) error
	Stop()
	GetGuid() string
	Receive()
}
