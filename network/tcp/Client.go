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

func (c *TCPClient) Connect(address string) network.IConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.AppLogger.Error("网络地址序列化失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
		return nil
	}
	c.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.AppLogger.Error("网络连接失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
		return nil
	}
	conn := NewConn(c.conn)
	conn.Start()
	return conn
}

func (c *TCPClient) Close() {
	c.conn.Close()
}
