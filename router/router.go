package router

import (
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
	"sync"
)

type IRouter interface {
	RegisterRouterFunc(msgObj proto.Message, handler func(conn network.IConnect, content []byte))
	CallRouterFunc(msgObj proto.Message) func(conn network.IConnect, content []byte)
	GetRouterList() map[int32]func(conn network.IConnect, content []byte)
}

type router struct {
	sync.RWMutex
	//外部路由
	RouterMap map[int32]func(conn network.IConnect, content []byte)
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
	this.RouterMap[protocolNum] = handler
}

//调用内部方法
func (this *router) CallRouterFunc(msgObj proto.Message) func(conn network.IConnect, content []byte) {
	this.RLocker()
	defer this.RUnlock()
	msgName := utils.GetProtoName(msgObj)
	msgId := utils.ProtocalNumber(msgName)
	return this.RouterMap[msgId]
}

func (this *router) GetRouterList() map[int32]func(conn network.IConnect, content []byte) {
	this.RLocker()
	defer this.RUnlock()
	return this.RouterMap
}
