package tcp

import (
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/network"
	router_ "github.com/yaice-rx/yaice/router"
	"net"
	"sync"
)

type TCPClient struct {
	sync.Mutex
	connManager    network.IConnManager
	receiveMsgChan chan *network.Msg
}

var TCPClientMgr = newTCPClientMgr()

func newTCPClientMgr() network.IClient {
	return &TCPClient{
		connManager:    NewConnManager(),
		receiveMsgChan: make(chan *network.Msg),
	}
}

// 连接服务器
func (this *TCPClient) Connect(IP string, port int) network.IConn {
	this.Lock()
	defer this.Unlock()
	addr := &net.TCPAddr{
		IP:   net.ParseIP(IP),
		Port: port,
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil
	}
	dealConn := newConnect(conn)
	//接收数据
	go this.receivePackets(dealConn)
	go this.run()
	return dealConn
}

// 接收数据包
func (this *TCPClient) receivePackets(conn network.IConn) {
	tmpBuffer := make([]byte, 0)
	var buffer = make([]byte, 1024)
	for {
		//read
		n, err := conn.GetConn().(*net.TCPConn).Read(buffer)
		if err != nil {
			logrus.Debug(conn.GetConn().(*net.TCPConn).RemoteAddr().String(), " connection error: ", err)
			return
		}
		//写入接收消息队列中
		tmpBuffer = network.UnPacket(conn, append(tmpBuffer, buffer[:n]...), this.receiveMsgChan)
	}
}

func (this *TCPClient) run() {
	go func() {
		for {
			select {
			//调用服务器内部方法
			case data := <-this.receiveMsgChan:
				func_ := router_.RouterMgr.CallRouterFunc(int32(data.ID))
				if func_ != nil {
					func_(data.Conn, data.Data)
				}
				break
			default:
				break
			}
		}
	}()
}

func (this *TCPClient) GetConns() network.IConnManager {
	return this.connManager
}

//停止
func (this *TCPClient) Stop() {
	this.connManager.ClearConn()
}
