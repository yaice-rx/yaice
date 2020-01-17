package tcp

import (
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"go.uber.org/zap"
	"net"
)

type TCPClient struct {
	conn *net.TCPConn
}

var TCPClientMgr = _NewTCPClient()

func _NewTCPClient() network.IClient {
	return &TCPClient{}
}

type Options struct {
	max uint
}

func WithMax(maxRetries uint) network.IOptions {
	return &Options{
		max: maxRetries,
	}
}

func (o *Options) GetMax() uint {
	return o.max
}

func (c *TCPClient) Connect(address string, opt network.IOptions) network.IConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.AppLogger.Error("网络地址序列化失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
		return nil
	}
LOOP:
	c.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		goto LOOP
		log.AppLogger.Error("网络连接失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
		return nil
	}
	conn := NewConn(c.conn)
	//加入到连接列表中
	ConnManagerMgr.Add(conn)
	//读取网络通道数据
	conn.Start()
	return conn
}

func (c *TCPClient) Close() {
	c.conn.Close()
}
