package network

import "google.golang.org/protobuf/proto"

type IClient interface {
	Connect() IConn
	Close()
	GetReceiveQueue() chan TransitData
	SendByte(message []byte) error
	SendProtobuf(message proto.Message) error
}
