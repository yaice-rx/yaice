package router

import (
	"github.com/golang/protobuf/proto"
	"sync"
	"yaice/network"
	"yaice/utils"
)

type IRouter interface {
	RegisterRouterFunc(msgObj proto.Message, handler func(conn network.IConnect, content []byte))
	CallRouterFunc(msgId int32) func(conn network.IConnect, content []byte)
}

type router struct {
	sync.RWMutex
	//外部路由
	ExternalRouterMap map[int32]func(conn network.IConnect, content []byte)
	//内部路由
	InternalRouterMap map[int32]func(conn network.IConnect, content []byte)
}

var RouterMgr = newRouter()

func newRouter() IRouter {
	return &router{}
}

//注册外部路由方法
func (this *router) RegisterRouterFunc(msgObj proto.Message, handler func(conn network.IConnect, content []byte)) {
	this.Lock()
	defer this.Unlock()
	msgName := utils.GetProtoName(msgObj)
	protocolNum := utils.ProtocalNumber(msgName)
	this.ExternalRouterMap[protocolNum] = handler
}

//调用内部方法
func (this *router) CallRouterFunc(msgId int32) func(conn network.IConnect, content []byte) {
	this.RLocker()
	defer this.RUnlock()
	return this.InternalRouterMap[msgId]
}
