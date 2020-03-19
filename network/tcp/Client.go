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
	address string
	conn    *net.TCPConn
	packet  network.IPacket
	opt     network.IOptions
}

type Options struct {
	max int32
}

func WithMax(maxRetries int32) network.IOptions {
	return &Options{
		max: maxRetries,
	}
}

func (o *Options) GetMax() int32 {
	return o.max
}

func (o *Options) SetMax() {
	atomic.AddInt32(&o.max, -1)
}

func NewClient(packet network.IPacket, address string, opt network.IOptions) network.IClient {
	c := &TCPClient{
		address: address,
		opt:     opt,
		packet:  packet,
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
		if c.opt.GetMax() <= 0 {
			log.AppLogger.Error("网络重连失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
			return nil
		}
		log.AppLogger.Error("重连失败：" + err.Error())
		c.opt.SetMax()
		goto LOOP
	}
	log.AppLogger.Info("网络连接中")
	conn := NewConn(c, c.conn, c.packet)
	//读取网络通道数据
	go conn.ReadThread()
	return conn
}

func (c *TCPClient) ReConnect() network.IConn {
	return c.Connect()
}

func (c *TCPClient) Close() {
	c.conn.Close()
}
