package tcp

import (
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/network"
	"net"
	"sync"
)

type TCPClient struct {
	sync.Mutex
	ConnManager network.IConnManager
}

var TCPClientMgr = newTCPClientMgr()

func newTCPClientMgr() network.IClient {
	return &TCPClient{
		ConnManager: NewConnManager(),
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
	return dealConn
}

// 接收数据包
func (this *TCPClient) receivePackets(conn network.IConn) {
	tmpBuffer := make([]byte, 0)
	var buffer = make([]byte, 1024)
	for {
		//read
		n, err := conn.GetNetworkConn().(*net.TCPConn).Read(buffer)
		if err != nil {
			logrus.Debug(conn.GetNetworkConn().(*net.TCPConn).RemoteAddr().String(), " connection error: ", err)
			return
		}
		//写入接收消息队列中
		dataPack := NewPacket()
		tmpBuffer = dataPack.Unpack(conn, append(tmpBuffer, buffer[:n]...))
	}
}

func (this *TCPClient) GetConns() network.IConnManager {
	return this.ConnManager
}

//停止
func (this *TCPClient) Stop() {
	this.ConnManager.ClearConn()
}
