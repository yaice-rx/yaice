package router

import (
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
	"sync"
	"time"
)

type IRouter interface {
	AddRouter(msgObj proto.Message, handler func(conn network.IConn, content []byte))
	DoRouterHandler(data network.IMessage)
	GetRoutersLen() int
	SendMsgToReadQueue(data network.IMessage)
	Run()
}

type router struct {
	sync.RWMutex
	//路由列表
	Routers map[uint32]func(conn network.IConn, content []byte)
	//网络读取队列
	ReadQueue chan network.IMessage
}

var RouterMgr = newRouter()

func newRouter() IRouter {
	router := &router{
		Routers:   make(map[uint32]func(conn network.IConn, content []byte)),
		ReadQueue: make(chan network.IMessage),
	}
	router.Run()
	return router
}

//添加逻辑处理方法
func (this *router) AddRouter(msgObj proto.Message, handler func(conn network.IConn, content []byte)) {
	this.Lock()
	defer this.Unlock()
	msgName := utils.GetProtoName(msgObj)
	protocolNum := utils.ProtocalNumber(msgName)
	this.Routers[protocolNum] = handler
}

//调用逻辑
func (this *router) DoRouterHandler(data network.IMessage) {
	this.RLock()
	defer this.RUnlock()
	if this.Routers[data.GetMsgId()] != nil {
		this.Routers[data.GetMsgId()](data.GetConn(), data.GetData())
	}
}

//获取数量
func (this *router) GetRoutersLen() int {
	return len(this.Routers)
}

func (this *router) Run() {
	go this.StartWorker(this.ReadQueue)
}

func (this *router) SendMsgToReadQueue(data network.IMessage) {
	this.ReadQueue <- data
}

func (this *router) StartWorker(readQueue chan network.IMessage) {
	for {
		select {
		case data := <-readQueue:
			this.DoRouterHandler(data)
		case <-time.After(5 * time.Second):
			logrus.Debug("conn dead now")
		}
	}
}
