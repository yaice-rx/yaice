package tcp

import (
	"errors"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/resource"
	router_ "github.com/yaice-rx/yaice/router"
	"github.com/yaice-rx/yaice/utils"
	"io"
	"net"
	"strconv"
	"sync"
)

type server struct {
	sync.Mutex
	network  string
	listener *net.TCPListener
	//连接列表
	ConnectsMgr network.IConnectList
	//发送消息chan
	sendMsgChan chan *network.Msg
	//接收消息chan
	receiveMsgChan chan *network.Msg
}

var TcpServerMgr = newTcpServer()

/**
 * 创建服务句柄
 */
func newTcpServer() network.IServer {
	serve := &server{
		network:     "tcp",
		ConnectsMgr: NewConnManager(),
	}
	return serve
}

//获取网络连接名称
func (this *server) GetNetwork() string {
	return this.network
}

// 开启网络服务
func (this *server) Start() error {
	for i := resource.ServiceResMgr.ExtranetPortStart; i < resource.ServiceResMgr.ExtranetPortEnd; i++ {
		tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(i))
		if nil != err {
			break
		}
		listener, err := net.ListenTCP("tcp", tcpAddr)
		if nil != err {
			break
		}
		this.listener = listener
		go func() {
			for {
				conn, err := this.listener.AcceptTCP()
				if nil != err || nil == conn {
					continue
				}
				//添加用户句柄
				dealConn := NewConnect(conn)
				this.ConnectsMgr.Add(dealConn)
				//如果当前连接数大于最大的连接数，则退出
				if this.ConnectsMgr.Len() > resource.ServiceResMgr.MaxConnectNumber {
					this.listener.Close()
					continue
				}
				//处理用户数据
				go func(conn *net.TCPConn) {
					for {
						//read
						var buffer = make([]byte, 1024)
						n, e := conn.Read(buffer)
						if e != nil {
							if e == io.EOF {
								break
							}
							break
						}
						//协议号
						msgId := utils.BytesToInt(buffer[:4])
						//写入接收消息队列中
						this.receiveMsgChan <- network.NewMsg(msgId, dealConn, buffer[4:n])
					}
				}(conn)
			}
		}()
		return nil
	}
	return errors.New("tcp port not found")
}

//读取网络数据
func (this *server) Run() {
	go func() {
		for {
			select {
			//调用服务器内部方法
			case data := <-this.receiveMsgChan:
				func_ := router_.RouterMgr.CallRouterFunc(data.ID)
				if func_ != nil {
					func_(data.Conn, data.Data)
				}
				break
			//调用网络流
			case data := <-this.sendMsgChan:
				go data.Conn.Send(data.Data)
				break
			}
		}
	}()
}

// 关闭网络接口
func (this *server) Close() {
	this.Lock()
	defer this.Unlock()
	this.listener.Close()
}
