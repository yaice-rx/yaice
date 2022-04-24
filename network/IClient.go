package network

import "github.com/golang/protobuf/proto"

type IClient interface {
	Connect() IClient
	Send(message proto.Message) error
	SendByte(message []byte) error
	Close()
}
