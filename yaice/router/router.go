package router

import (
	"github.com/golang/protobuf/proto"
	"sync"
	"yaice/network"
	"yaice/utils"
)

type IRouter interface {
	RegisterExternalRouterFunc(msgObj proto.Message,handler func(conn network.IConnect,content []byte))
	RegisterInternalRouterFunc(msgObj proto.Message,handler func(conn network.IConnect,content []byte))
	CallInternalRouterFunc(msgId int32)func(conn network.IConnect,content []byte)
	CallExternalRouterFunc(msgId int32)func(conn network.IConnect,content []byte)
}

type router struct {
	sync.RWMutex
	//外部路由
	ExternalRouterMap map[int32]func(conn network.IConnect,content []byte)
	//内部路由
	InternalRouterMap map[int32]func(conn network.IConnect,content []byte)
}

var RouterMgr = newRouter()

func newRouter()IRouter{
	return &router{}
}

//注册外部路由方法
func (this *router)RegisterExternalRouterFunc(msgObj proto.Message,handler func(conn network.IConnect,content []byte)){
	this.Lock()
	defer this.Unlock()
	msgName := utils.GetProtoName(msgObj)
	protocolNum := utils.ProtocalNumber(msgName)
	this.ExternalRouterMap[protocolNum] = handler
}


//注册内部路由方法
func (this *router)RegisterInternalRouterFunc(msgObj proto.Message,handler func(conn network.IConnect,content []byte)){
	this.Lock()
	defer this.Unlock()
	msgName := utils.GetProtoName(msgObj)
	protocolNum := utils.ProtocalNumber(msgName)
	this.ExternalRouterMap[protocolNum] = handler
}

//调用内部方法
func (this *router)CallInternalRouterFunc(msgId int32)func(conn network.IConnect,content []byte){
	return this.InternalRouterMap[msgId]
}

//调用外部方法
func (this *router)CallExternalRouterFunc(msgId int32)func(conn network.IConnect,content []byte){
	return this.ExternalRouterMap[msgId]
}

