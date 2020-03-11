package tcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/router"
	"github.com/yaice-rx/yaice/utils"
	"go.uber.org/zap"
	"io"
	"net"
	"sync/atomic"
	"time"
)

type Conn struct {
	guid         string
	pkg          network.IPacket
	conn         *net.TCPConn
	receiveQueue chan network.TransitData
	sendQueue    chan []byte
	stopChan     chan bool
	isClosed     bool
	times        int64
	reconnCount  uint32
	reconnTime   int64
	data         interface{}
}

const TOKEN_CHECK_MAX_TRY_NUM = 60
const TOKEN_CHECK_MAX_WAIT_TIME = 5

func NewConn(conn *net.TCPConn, pkg network.IPacket) network.IConn {
	return &Conn{
		guid:         uuid.NewV4().String(),
		conn:         conn,
		pkg:          pkg,
		isClosed:     false,
		receiveQueue: make(chan network.TransitData, 10),
		sendQueue:    make(chan []byte, 10),
		stopChan:     make(chan bool),
		reconnCount:  0,
		reconnTime:   0,
		times:        time.Now().Unix(),
	}
}

type headerMsg struct {
	DataLen uint32
}

func (c *Conn) readThread() {
	for {
		//1 先读出流中的head部分
		headData := make([]byte, c.pkg.GetHeadLen())
		_, err := io.ReadFull(c.conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			log.AppLogger.Debug("network io read data err:" + err.Error())
			break
		}
		//强制规定网络数据包头4位必须是网络的长度
		//创建一个从输入二进制数据的ioReader
		headerBuff := bytes.NewReader(headData)
		msg := &headerMsg{}
		if err := binary.Read(headerBuff, binary.BigEndian, &msg.DataLen); err != nil {
			log.AppLogger.Debug("server unpack err:" + err.Error())
			break
		}
		if msg.DataLen > 0 {
			//msg 是有data数据的，需要再次读取data数据
			contentData := make([]byte, msg.DataLen)
			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(c.conn, contentData)
			if err != nil {
				log.AppLogger.Debug("network io read data err:" + err.Error())
				break
			}
			//解压网络数据包
			msgData, err := c.pkg.Unpack(contentData)
			//写入通道数据
			c.receiveQueue <- network.TransitData{
				MsgId: msgData.GetMsgId(),
				Data:  msgData.GetData(),
			}
		}
	}
}

func (c *Conn) writeThread() {
	for {
		select {
		case data, state := <-c.sendQueue:
			if state {
			LOOP:
				_, err := c.conn.Write(data)
				if err != nil {
					time.Sleep(800 * time.Microsecond)
					if atomic.LoadUint32(&c.reconnCount) >= TOKEN_CHECK_MAX_TRY_NUM {
						log.AppLogger.Info("reconnect faild")
						atomic.SwapUint32(&c.reconnCount, 0)
						c.Close()
					} else {
						log.AppLogger.Info("reconnecting ...")
						atomic.AddUint32(&c.reconnCount, 1)
						goto LOOP
					}
				} else {
					log.AppLogger.Info("reconnect  success ...")
					//连通后，重置为0
					atomic.SwapUint32(&c.reconnCount, 0)
				}
				break
			}
		}
	}
}

//发送协议体
func (c *Conn) Send(message proto.Message) error {
	data, err := proto.Marshal(message)
	protoId := utils.ProtocalNumber(utils.GetProtoName(message))
	if err != nil {
		log.AppLogger.Error("发送消息时，序列化失败 : "+err.Error(), zap.Int32("MessageId", protoId))
		return err
	}
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	c.sendQueue <- c.pkg.Pack(network.TransitData{protoId, data})
	return nil
}

//发送组装好的协议，但是加密始终是在组装包的时候完成加密功能
func (c *Conn) SendByte(message []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	c.sendQueue <- message
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
				if data.MsgId != 0 {
					router.RouterMgr.ExecRouterFunc(c, data)
				}
				break
			//关闭Conn连接
			case <-c.stopChan:
				return
			}
		}
	}()
}

func (c *Conn) Close() {
	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.stopChan <- true
	c.isClosed = true
	c.conn.Close()
	close(c.receiveQueue)
	close(c.sendQueue)
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

func (c *Conn) GetConn() interface{} {
	return c.conn
}
