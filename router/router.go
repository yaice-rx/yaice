package router

import (
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
	"net/http"
	"sync"
)

type IRouter interface {
	RegisterHttpHandlerFunc(router string, handler func(write http.ResponseWriter, request *http.Request))
	RegisterRouterFunc(msgObj proto.Message, handler func(conn network.IConnect, content []byte))
	GetHttpHandlerMap() map[string]func(write http.ResponseWriter, request *http.Request)
	CallRouterFunc(msgId int32) func(conn network.IConnect, content []byte)
	GetRouterMap() map[int32]func(conn network.IConnect, content []byte)
	GetHttpHandlerCount() int
	GetRouterCount() int
}

type router struct {
	sync.RWMutex
	//kcp tcp udp raknet 路由在此处理
	routerMap map[int32]func(conn network.IConnect, content []byte)
	//http路由
	httpHandlerMap map[string]func(write http.ResponseWriter, request *http.Request)
}

var RouterMgr = newRouter()

func newRouter() IRouter {
	return &router{
		routerMap:      make(map[int32]func(conn network.IConnect, content []byte)),
		httpHandlerMap: make(map[string]func(write http.ResponseWriter, request *http.Request)),
	}
}

func (this *router) RegisterHttpHandlerFunc(router string, handler func(write http.ResponseWriter, request *http.Request)) {
	this.Lock()
	defer this.Unlock()
	this.httpHandlerMap[router] = handler
}

func (this *router) GetHttpHandlerMap() map[string]func(write http.ResponseWriter, request *http.Request) {
	this.Lock()
	defer this.Unlock()
	return this.httpHandlerMap
}

//注册外部路由方法
func (this *router) RegisterRouterFunc(msgObj proto.Message, handler func(conn network.IConnect, content []byte)) {
	this.Lock()
	defer this.Unlock()
	msgName := utils.GetProtoName(msgObj)
	protocolNum := utils.ProtocalNumber(msgName)
	this.routerMap[protocolNum] = handler
}

//调用内部方法
func (this *router) CallRouterFunc(msgId int32) func(conn network.IConnect, content []byte) {
	this.Lock()
	defer this.Unlock()
	return this.routerMap[msgId]
}

func (this *router) GetRouterMap() map[int32]func(conn network.IConnect, content []byte) {
	this.Lock()
	defer this.Unlock()
	return this.routerMap
}

func (this *router) GetHttpHandlerCount() int {
	return len(this.httpHandlerMap)
}

func (this *router) GetRouterCount() int {
	return len(this.httpHandlerMap)
}
