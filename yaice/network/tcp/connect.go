package tcp

import (
	"github.com/satori/go.uuid"
	"net"
	"sync"
	"time"
	"yaice/network"
)

type Connect struct {
	sendGuard sync.RWMutex
	guid      string
	session   *net.TCPConn
	timer     int64
}

/**
 * 创建Connect
 */
func NewConnect(conn *net.TCPConn) network.IConnect {
	return &Connect{
		guid:    uuid.NewV4().String(),
		session: conn,
		timer:   time.Now().Unix(),
	}
}

/**
 * 获取guid
 */
func (this *Connect) GetGuid() string {
	return this.guid
}

/**
 * 发送消息
 */
func (this *Connect) Send(data []byte) error {
	this.sendGuard.Lock()
	defer this.sendGuard.Unlock()
	_, err := this.session.Write(data)
	return err
}

/**
 * 接收消息
 */
func (this *Connect) Receive() {

}

/**
 * 停止连接
 */
func (this *Connect) Stop() {
	this.session.Close()
}
