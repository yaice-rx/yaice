package tcp

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
	"net"
	"sync"
	"time"
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
func (this *Connect) Send(message proto.Message) error {
	protoNumber := utils.ProtocalNumber(utils.GetProtoName(message))
	data, err := json.Marshal(message)
	msg := utils.IntToBytes(protoNumber)
	msg = append(msg, data...)
	this.sendGuard.Lock()
	defer this.sendGuard.Unlock()
	_, err = this.session.Write(msg)
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
