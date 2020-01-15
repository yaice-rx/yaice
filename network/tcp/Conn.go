package tcp

import (
	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/satori/go.uuid"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/router"
	"github.com/yaice-rx/yaice/utils"
	"net"
	"time"
)

type Conn struct {
	guid         string
	conn         *net.TCPConn
	receiveQueue chan network.IMessage
	sendQueue    chan []byte
	times        int64
}

func NewConn(conn *net.TCPConn) network.IConn {
	return &Conn{
		guid:         uuid.NewV4().String(),
		conn:         conn,
		receiveQueue: make(chan network.IMessage),
		sendQueue:    make(chan []byte),
		times:        time.Now().Unix(),
	}
}

func (c *Conn) startReadThread() {
	defer func() {
		ConnManagerMgr.Remove(c.guid)
		c.conn.Close()
	}()
	var errs error
	tempBuff := make([]byte, 0)
	readBuff := make([]byte, 1024)
	data := make([]byte, 1024)
	msgId := int32(0)
	for {
		//read
		n, e := c.conn.Read(readBuff)
		if e != nil {
			continue
		}
		c.UpdateTime()
		//写入接收消息队列中
		dataPack := NewPacket()
		tempBuff = append(tempBuff, readBuff[:n]...)
		tempBuff, data, msgId, errs = dataPack.Unpack(tempBuff)
		if errs != nil {
			log.AppLogger.Error("接收消息时候，解压数据包错误 :" + errs.Error())
			continue
		}
		c.receiveQueue <- NewMessage(msgId, data, c)
	}
}

func (c *Conn) startWriteThread() {
	for {
		select {
		case data, state := <-c.sendQueue:
			if state {
				_, err := c.conn.Write(data)
				if err != nil {
					//发送错误,将数据重新写入通道重新发送
					log.AppLogger.Fatal("发送消息失败 ，发送人： " + c.guid + ",错误提示：" + err.Error())
					c.sendQueue <- data
					return
				}
			}
		}
	}
}

func (c *Conn) Send(message proto.Message) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	data, err := json.Marshal(message)
	if err != nil {
		log.AppLogger.Fatal("发送消息时，序列化失败 : " + err.Error())
		return err
	}
	protoNumber := utils.ProtocalNumber(utils.GetProtoName(message))
	dataPack := NewPacket()
	c.sendQueue <- dataPack.Pack(NewMessage(protoNumber, data, c))
	return nil
}

func (c *Conn) GetGuid() string {
	return c.guid
}

func (c *Conn) Start() {
	go c.startReadThread()
	go c.startWriteThread()
	go func() {
		for {
			select {
			case msg, state := <-c.receiveQueue:
				if state {
					router.RouterMgr.ExecRouterFunc(msg)
				}
				break
			case <-time.After(time.Second * 300):
				//连接超时
				ConnManagerMgr.Remove(c.guid)
				break
			}
		}
	}()
}

func (c *Conn) GetTimes() int64 {
	return c.times
}

func (c *Conn) UpdateTime() {
	c.times = time.Now().Unix()
}

func (c *Conn) Close() {
	close(c.sendQueue)
	close(c.receiveQueue)
}
