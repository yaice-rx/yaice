package tcp_net

import (
	"context"
	"github.com/yaice-rx/yaice/core"
	"github.com/yaice-rx/yaice/packates"
	"net"
	"time"
)

type TCPClient struct {
	type_            core.ServeType //网络类型
	dialRetriesCount int32          //拨号重试次数
	address          string         //地址
	conn             core.Connection
	packet           packates.IPacket
	ctx              context.Context
	cancel           context.CancelFunc
	callFunc         func(conn core.Connection, err error)
}

func NewClient(packet packates.IPacket, address string, callFunc func(conn core.Connection, err error)) core.Client {
	c := &TCPClient{
		type_:            core.Serve_Client,
		address:          address,
		packet:           packet,
		dialRetriesCount: 0,
		callFunc:         callFunc,
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel
	return c
}

func (c *TCPClient) Connect(ctx context.Context) core.Connection {
	tcpAddr, err := net.ResolveTCPAddr("tcp_net", c.address)
	if err != nil {
		c.callFunc(c.conn, err)
		return nil
	}
LOOP:
	tcpConn, err := net.DialTCP("tcp_net", nil, tcpAddr)
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
	c.conn = NewConn(c.ctx, c, tcpConn, c.packet, c.opt, conns.Serve_Client, c.ctx, c.cancel)
	//读取网络通道数据
	go c.conn.Start()
	return c.conn
}

func (c *TCPClient) ReConnect(ctx context.Context) core.Connection {
	return c.Connect(ctx)
}

func (c *TCPClient) Close(ctx context.Context, err error) error {
	if c.cancel != nil {
		c.cancel()
	}
	// 关闭连接
	if c.conn != nil {
		c.conn.Close()
	}
	// 调用回调函数通知连接关闭
	if c.callFunc != nil {
		c.callFunc(c.conn, err)
	}
	return nil
}
