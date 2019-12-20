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
	sendGuard   sync.RWMutex
	guid        string
	session     *net.TCPConn
	timer       int64
	connectType string
	sendQueue   chan []byte
}

// 创建Connect
func newConnect(conn *net.TCPConn) network.IConn {
	return &Connect{
		guid:        uuid.NewV4().String(),
		session:     conn,
		timer:       time.Now().Unix(),
		sendQueue:   make(chan []byte),
		connectType: "tcp",
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
				//发送错误,将数据重新写入通道重新发送
				this.sendQueue <- data
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
	go func() {
		protoNumber := utils.ProtocalNumber(utils.GetProtoName(message))
		dataPack := NewPacket()
		this.sendQueue <- dataPack.Pack(NewMessage(protoNumber, data, this))
	}()
	return nil
}

func (this *Connect) GetNetworkConn() interface{} {
	return this.session
}

func (this *Connect) GetConnectType() string {
	return this.connectType
}

// 停止连接
func (this *Connect) Stop() {
	this.session.Close()
}
