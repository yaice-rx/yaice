package tcp

import (
	"github.com/yaice-rx/yaice/network"
	"net"
	"sync"
)

type client struct {
	sync.Mutex
	conn *net.TCPConn
}

var ClientMgr = newClient()

/**
 * 初始化客户端
 */
func newClient() network.IClient {
	return &client{}
}

/**
 * 连接服务器
 */
func (this *client) Connect(IP string, port int) network.IConnect {
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
	return NewConnect(conn)
}

/**
 * 停止
 */
func (this *client) Stop() {
	this.conn.Close()
}
