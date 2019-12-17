package tcp

import "C"
import (
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/resource"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
)

type TCPServer struct {
	sync.Mutex
	network  string
	listener *net.TCPListener
	//连接列表
	connectCount uint32
}

var TcpServerMgr = newTcpServer()

// 创建服务句柄
func newTcpServer() network.IServer {
	serve := &TCPServer{
		network: "tcp",
	}
	return serve
}

//获取网络连接名称
func (this *TCPServer) GetNetworkName() string {
	return this.network
}

// 开启网络服务
func (this *TCPServer) Start(port chan int) {
	for i := resource.ServiceResMgr.PortStart; i < resource.ServiceResMgr.PortEnd; i++ {
		tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(i))
		if nil != err {
			continue
		}
		listener, err := net.ListenTCP("tcp", tcpAddr)
		if nil != err {
			continue
		}
		logrus.Debug("tcp listen port :", i)
		port <- i
		this.listener = listener
		for {
			tcpConn, err := listener.AcceptTCP()
			if nil != err || nil == tcpConn {
				continue
			}
			//如果当前连接数大于最大的连接数，则退出
			if this.connectCount > uint32(resource.ServiceResMgr.MaxConnectNumber) {
				this.listener.Close()
				continue
			}
			//tcpConn.SetReadDeadline(time.Now().Add(time.Duration(20) * time.Second))
			atomic.AddUint32(&this.connectCount, 1)
			//添加用户句柄
			conn := newConnect(tcpConn)
			//处理用户数据
			go conn.Start()
		}
		return
	}
	port <- -1
	logrus.Debug("tcp port not found")
}

// 关闭网络接口
func (this *TCPServer) Close() {
	this.Lock()
	defer this.Unlock()
	this.listener.Close()
}
