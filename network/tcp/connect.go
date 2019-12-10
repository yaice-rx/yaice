package tcp

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
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
	sendQueue chan []byte
}

// 创建Connect
func newConnect(conn *net.TCPConn) network.IConn {
	return &Connect{
		guid:      uuid.NewV4().String(),
		session:   conn,
		timer:     time.Now().Unix(),
		sendQueue: make(chan []byte),
	}
}

//获取guid
func (this *Connect) GetGuid() string {
	return this.guid
}

func (this *Connect) Start() {
	go this.StartReadThread()
	go this.StartWriteThread()
}

func (this *Connect) StartReadThread() {
	tmpBuffer := make([]byte, 0)
	var buffer = make([]byte, 1024)
	for {
		//read
		n, e := this.session.Read(buffer)
		if e != nil {
			logrus.Debug(this.session.RemoteAddr().String(), " connection error: ", e)
			return
		}
		//写入接收消息队列中
		dataPack := NewPacket()
		tmpBuffer = dataPack.Unpack(this, append(tmpBuffer, buffer[:n]...))
	}
}

func (this *Connect) StartWriteThread() {
	for {
		select {
		case data := <-this.sendQueue:
			if _, err := this.session.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		}
	}
}

// 发送消息
func (this *Connect) SendMsg(message proto.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	protoNumber := utils.ProtocalNumber(utils.GetProtoName(message))
	dataPack := NewPacket()
	this.sendQueue <- dataPack.Pack(NewMessage(protoNumber, data, this))
	return nil
}

func (this *Connect) GetConn() interface{} {
	return this.session
}

// 停止连接
func (this *Connect) Stop() {
	this.session.Close()
}
