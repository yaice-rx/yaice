package tcp

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/yaice-rx/yaice/network"
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

// 创建Connect
func newConnect(conn *net.TCPConn) network.IConn {
	return &Connect{
		guid:    uuid.NewV4().String(),
		session: conn,
		timer:   time.Now().Unix(),
	}
}

//获取guid
func (this *Connect) GetGuid() string {
	return this.guid
}

// 发送消息
func (this *Connect) Send(message proto.Message) error {
	data := network.Packet(message)
	if data != nil {
		this.session.Write(data)
		return nil
	}
	return errors.New("data assembly error")
}

func (this *Connect) GetConn() interface{} {
	return this.session
}

// 停止连接
func (this *Connect) Stop() {
	this.session.Close()
}
