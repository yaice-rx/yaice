package tcp

import (
	"fmt"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"go.uber.org/zap"
	"net"
	"time"
)


type TCPClient struct {
	type_            network.ServeType
	dialRetriesCount int32
	address          string
	conn             network.IConn
	packet           network.IPacket
	opt              network.IOptions
	connStateFunc    func(conn network.IConn)
	state 			 int32
}



func NewClient(packet network.IPacket, address string, opt network.IOptions) network.IClient {
	c := &TCPClient{
		type_:            network.Serve_Client,
		address:          address,
		packet:           packet,
		opt:              opt,
		dialRetriesCount: 0,
		state:1,
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
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		time.Sleep(3 * time.Second)
		if c.opt.GetMaxRetires() < c.dialRetriesCount {
			log.AppLogger.Error("网络重连失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
			return nil
		}
		log.AppLogger.Warn(fmt.Sprintf("第{%d}网络重连中", c.dialRetriesCount))
		c.dialRetriesCount += 1
		goto LOOP
	}
	//连接上的时候，重置连接次数
	c.dialRetriesCount = 0
	c.conn = NewConn(c, tcpConn, c.packet, network.Serve_Client)
	//读取网络通道数据
	go c.conn.Start()
	return c.conn
}

func (c *TCPClient) ReConnect() network.IConn {
	return c.Connect()
}

func (c *TCPClient) Close() {
	//设置当前客户端的状态
	c.state = 0
	c.conn.Close()
}
