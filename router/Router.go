package router

import (
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
	"sync"
)

type router struct {
	sync.RWMutex
	routers map[int32]func(conn network.IConn, content []byte)
}

var RouterMgr = _NewRouterMgr()

func _NewRouterMgr() IRouter {
	return &router{
		routers: make(map[int32]func(conn network.IConn, content []byte)),
	}
}

func (r *router) AddRouter(msgObj proto.Message, handler func(conn network.IConn, content []byte)) {
	r.Lock()
	defer r.Unlock()
	msgName := utils.GetProtoName(msgObj)
	protocolNum := utils.ProtocalNumber(msgName)
	r.routers[protocolNum] = handler
}

func (r *router) ExecRouterFunc(message network.IMessage) {
	r.Lock()
	defer r.Unlock()
	handler := r.routers[message.GetMsgId()]
	if handler != nil {
		handler(message.GetConn(), message.GetData())
	}
}
