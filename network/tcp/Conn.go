package tcp

import (
	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/satori/go.uuid"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/router"
	"github.com/yaice-rx/yaice/utils"
	"go.uber.org/zap"
	"net"
	"time"
)

type Conn struct {
	guid         string
	pkg          network.IPacket
	conn         *net.TCPConn
	receiveQueue chan network.IMessage
	sendQueue    chan network.IMessage
	stopChan     chan bool
	times        int64
	data         interface{}
}

func NewConn(conn *net.TCPConn, pkg network.IPacket) network.IConn {
	return &Conn{
		guid:         uuid.NewV4().String(),
		conn:         conn,
		pkg:          pkg,
		receiveQueue: make(chan network.IMessage, 10),
		sendQueue:    make(chan network.IMessage, 10),
		stopChan:     make(chan bool),
		times:        time.Now().Unix(),
	}
}

func (c *Conn) readThread() {
	var errs error
	tempBuff := make([]byte, 0)
	readBuff := make([]byte, 1024)
	data := make([]byte, 1024)
	msgId := int32(0)
	for {
		//判定连接句柄是否否关闭
		if <-c.stopChan {
			return
		}
		//开启从网络中读取数据
		n, e := c.conn.Read(readBuff)
		if e != nil {
			//网络数据读取失败，关闭该连接句柄
			c.stopChan <- true
			continue
		}
		c.UpdateTime()
		//写入接收消息队列中
		tempBuff = append(tempBuff, readBuff[:n]...)
		tempBuff, data, msgId, errs = c.pkg.Unpack(tempBuff)
		if errs != nil {
			//数据验证不过关，关闭该连接句柄
			log.AppLogger.Error("接收消息时候，解压数据包错误 :" + errs.Error())
			c.stopChan <- true
			continue
		}
		//写入通道数据
		c.receiveQueue <- NewMessage(msgId, data, c)
	}
}

func (c *Conn) writeThread() {
	for {
		select {
		case data, state := <-c.sendQueue:
			if state {
				_, err := c.conn.Write(c.pkg.Pack(data))
				if err != nil {
					//首先判断 发送多次，依然不能连接服务器，就此直接断开
					if data.GetCount() > 3 {
						c.Close()
						return
					} else {
						//发送错误,将数据重新写入通道重新发送
						log.AppLogger.Error("发送消息失败 ，发送人： "+c.guid+",错误提示："+err.Error(), zap.Int32("MessageId", data.GetMsgId()))
						data.AddCount()
						c.sendQueue <- data
						break
					}
				}
			} else {
				log.AppLogger.Error("发送消息失败 ，发送人： "+c.guid+",通道已关闭", zap.Int32("MessageId", data.GetMsgId()))
				return
			}
		}
	}
}

func (c *Conn) Send(message proto.Message) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	data, err := json.Marshal(message)
	protoId := utils.ProtocalNumber(utils.GetProtoName(message))
	if err != nil {
		log.AppLogger.Error("发送消息时，序列化失败 : "+err.Error(), zap.Int32("MessageId", protoId))
		return err
	}
	c.sendQueue <- NewMessage(protoId, data, c)
	return nil
}

func (c *Conn) GetGuid() string {
	return c.guid
}

func (c *Conn) Start() {
	go c.readThread()
	go c.writeThread()
	go func() {
		for {
			select {
			//读取网络数据
			case data := <-c.receiveQueue:
				if data != nil {
					router.RouterMgr.ExecRouterFunc(data)
				}
				break
			//关闭Conn连接
			case data := <-c.stopChan:
				if data {
					//关闭通道
					return
				}
				break
			}
		}
	}()
}

func (c *Conn) Close() {
	//关闭读取写入通道
	c.stopChan <- true
}

func (c *Conn) GetTimes() int64 {
	return c.times
}

func (c *Conn) UpdateTime() {
	c.times = time.Now().Unix()
}

func (c *Conn) SetData(data interface{}) {
	c.data = data
}
