package tcp

import (
	"sync"
	"yaice/network"
	"yaice/router"
	"yaice/utils"
)

type Connects struct {
	sync.Mutex
	//map[guid]连接句柄
	Connects 		map[string]network.IConnect
	//消息接收句柄
	ReceiveMsgQueue	chan *network.MsgQueue
	//消息发送句柄
	SendMsgQueue	chan *network.MsgQueue
}

var ConnectsMgr = newConnects()

/**
 * 初始化
 */
func newConnects() network.IConnects {
	return &Connects{
		Connects:		make(map[string]network.IConnect),
		ReceiveMsgQueue:make(chan *network.MsgQueue),
		SendMsgQueue:	make(chan *network.MsgQueue),
	}
}

/**
 * 添加连接句柄
 */
func (this *Connects)Add(conn network.IConnect){
	this.Lock()
	defer this.Unlock()
	this.Connects[conn.GetGuid()] = conn
}

/**
 * 移除连接句柄
 */
func (this *Connects)Remove(conn network.IConnect){
	this.Lock()
	defer this.Unlock()
	delete(this.Connects,conn.GetGuid())
}

/**
 * 根据guid获取连接句柄
 */
func (this *Connects)Get(guid string)network.IConnect {
	return this.Connects[guid]
}

/**
 * 获取连接数量
 */
func (this *Connects)Len()int{
	return len(this.Connects)
}

/**
 * 启动线程
 */
func (this *Connects)RunThread(){
	go func() {
		for  {
			select {
			case data := <- this.ReceiveMsgQueue:
				func_ := router.RouterMgr.CallInternalRouterFunc(data.MsgId)
				func_(data.Conn,data.Data)
				break
			case data := <- this.SendMsgQueue:
				go func() {
					content := utils.IntToBytes(data.MsgId)
					content = append(content,data.Data...)
					err := data.Conn.Send(content)
					if err != nil {
						return
					}
				}()
				break
			default:
				break
			}
		}
	}()
}