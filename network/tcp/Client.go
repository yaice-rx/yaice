package tcp

import (
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"go.uber.org/zap"
	"net"
	"sync/atomic"
	"time"
)

type TCPClient struct {
	dialRetriesCount int32
	address          string
	tID              string
	conn             *net.TCPConn
	packet           network.IPacket
	opt              network.IOptions
	connStateFunc    func(conn network.IConn)
}

func NewClient(packet network.IPacket, address string, opt network.IOptions) network.IClient {
	c := &TCPClient{
		address:          address,
		opt:              opt,
		packet:           packet,
		dialRetriesCount: 0,
	}
	return c
}

func (c *TCPClient) Connect() network.IConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.address)
	if err != nil {
		log.AppLogger.Error("网络地址序列化失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
		return nil
	}
LOOP:
	c.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		time.Sleep(3 * time.Second)
		if c.opt.GetMaxRetires() > atomic.LoadInt32(&c.dialRetriesCount) {
			log.AppLogger.Error("网络重连失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
			return nil
		}
		log.AppLogger.Error("重连失败：" + err.Error())
		atomic.AddInt32(&c.dialRetriesCount, 1)
		goto LOOP
	}
	conn := NewConn(c, c.conn, c.packet)
	//读取网络通道数据
	go conn.Start()
	return conn
}

func (c *TCPClient) ReConnect() network.IConn {
	return c.Connect()
}

func (c *TCPClient) Close() {
	c.conn.Close()
}
