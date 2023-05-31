package tcp

import (
	"context"
	"github.com/yaice-rx/yaice/network"
	"net"
	"time"
)

type TCPClient struct {
	type_            network.ServeType //网络类型
	dialRetriesCount int32             //拨号重试次数
	address          string            //地址
	conn             network.IConn
	packet           network.IPacket
	opt              network.IOptions
	ctx              context.Context
	cancel           context.CancelFunc
	callFunc         func(conn network.IConn, err error)
}

func NewClient(packet network.IPacket, address string, opt network.IOptions, callFunc func(conn network.IConn, err error)) network.IClient {
	c := &TCPClient{
		type_:            network.Serve_Client,
		address:          address,
		packet:           packet,
		opt:              opt,
		dialRetriesCount: 0,
		callFunc:         callFunc,
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel
	return c
}

func (c *TCPClient) Connect() network.IConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.address)
	if err != nil {
		c.callFunc(c.conn, err)
		return nil
	}
LOOP:
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		time.Sleep(3 * time.Second)
		if c.opt.GetMaxRetires() < c.dialRetriesCount {
			c.callFunc(c.conn, err)
			return nil
		}
		c.dialRetriesCount += 1
		goto LOOP
	}
	//连接上的时候，重置连接次数
	c.dialRetriesCount = 0
	c.conn = NewConn(c, tcpConn, c.packet, c.opt, network.Serve_Client, c.ctx, c.cancel)
	//读取网络通道数据
	go c.conn.Start()
	return c.conn
}

func (c *TCPClient) ReConnect() network.IConn {
	return c.Connect()
}

func (c *TCPClient) Close(err error) {
	c.cancel()
	//设置当前客户端的状态
	c.callFunc(c.conn, err)
	c.conn.Close()
}
