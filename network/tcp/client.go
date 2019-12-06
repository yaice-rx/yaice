package tcp

import (
	"bufio"
	"fmt"
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
	go this.receivePackets()
	return NewConnect(conn)
}

// 接收数据包
func (this *client) receivePackets() {
	reader := bufio.NewReader(this.conn)
	for {
		//承接上面说的服务器端的偷懒，我这里读也只是以\n为界限来读区分包
		msg, err := reader.ReadString('\n')
		if err != nil {
			//在这里也请处理如果服务器关闭时的异常
			//close(client.stopChan)
			break
		}
		fmt.Println(msg)
	}
}

/**
 * 停止
 */
func (this *client) Stop() {
	this.conn.Close()
}
