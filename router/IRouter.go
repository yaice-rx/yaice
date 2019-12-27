package router

import (
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/network"
)

type IRouter interface {
	AddRouter(msgObj proto.Message, handler func(conn network.IConn, content []byte))
	ExecRouterFunc(message network.IMessage)
}
