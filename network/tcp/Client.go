package tcp

import (
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"go.uber.org/zap"
	"net"
	"sync/atomic"
)

type TCPClient struct {
	conn *net.TCPConn
}

var TCPClientMgr = _NewTCPClient()

func _NewTCPClient() network.IClient {
	return &TCPClient{}
}

type Options struct {
	max uint32
}

func WithMax(maxRetries uint32) network.IOptions {
	return &Options{
		max: maxRetries,
	}
}

func (o *Options) GetMax() uint32 {
	return o.max
}

func (o *Options) SetMax() {
	atomic.AddUint32(&o.max, 1)
}

func (c *TCPClient) Connect(packet network.IPacket, address string, opt network.IOptions) network.IConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.AppLogger.Error("网络地址序列化失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
		return nil
	}
LOOP:
	c.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		if opt.GetMax() > 3 {
			log.AppLogger.Error("网络重连失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
			return nil
		}
		opt.SetMax()
		goto LOOP
	}
	conn := NewConn(c.conn, packet)
	//加入到连接列表中
	ConnManagerMgr.Add(conn)
	//读取网络通道数据
	conn.Start()
	return conn
}

func (c *TCPClient) Close() {
	c.conn.Close()
}
