package tcp

import (
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
	"time"
)

type Conn struct {
	serve        interface{}
	guid         string
	pkg          network.IPacket
	conn         *net.TCPConn
	receiveQueue chan network.TransitData
	sendQueue    chan []byte
	isClosed     bool
	times        int64
	data         interface{}
}

func NewConn(serve interface{}, conn *net.TCPConn, pkg network.IPacket) network.IConn {
	return &Conn{
		serve:        serve,
		guid:         uuid.NewV4().String(),
		conn:         conn,
		pkg:          pkg,
		isClosed:     false,
		receiveQueue: make(chan network.TransitData, 10),
		sendQueue:    make(chan []byte, 10),
		times:        time.Now().Unix(),
	}
}

type headerMsg struct {
	DataLen uint32
}

func (c *Conn) Start() {
	go c.ReadThread()
	go c.WriteThread()
	go func() {
		for {
			select {
			//读取网络数据
			case data := <-c.receiveQueue:
				if data.MsgId != 0 {
					router.RouterMgr.ExecRouterFunc(c, data)
				}
				break
			}
		}
	}()
}

func (c *Conn) Close() {
	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true
	//c.conn.Close()
	close(c.receiveQueue)
	close(c.sendQueue)
}

func (c *Conn) GetGuid() string {
	return c.guid
}

func (c *Conn) GetConn() interface{} {
	return c.conn
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

func (c *Conn) ReadThread() {
	for {
		//1 先读出流中的head部分
		headData := make([]byte, c.pkg.GetHeadLen())
		_, err := io.ReadFull(c.conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			if err != io.EOF {
				log.AppLogger.Info("network io read data err:" + err.Error())
			}
			break
		}
		//强制规定网络数据包头4位必须是网络的长度
		//创建一个从输入二进制数据的ioReader
		msgLen := utils.BytesToInt(headData)
		if msgLen > 0 {
			//msg 是有data数据的，需要再次读取data数据
			contentData := make([]byte, msgLen)
			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(c.conn, contentData)
			if err != nil {
				log.AppLogger.Info("network io read data err - 0:" + err.Error())
				break
			}
			//解压网络数据包
			msgData, err := c.pkg.Unpack(contentData)
			if msgData == nil {
				log.AppLogger.Info("network io read data err - 1:" + err.Error())
				break
			}
			//写入通道数据
			c.receiveQueue <- network.TransitData{
				MsgId: msgData.GetMsgId(),
				Data:  msgData.GetData(),
			}
		}
	}
}

func (c *Conn) WriteThread() {
	for {
		select {
		case data, state := <-c.sendQueue:
			if state {
				_, err := c.conn.Write(data)
				if err != nil {
					if c.serve.(*TCPClient) != nil {
						c.conn = c.serve.(*TCPClient).ReConnect().GetConn().(*net.TCPConn)
					}
				}
			}
			break
		}
	}
}
