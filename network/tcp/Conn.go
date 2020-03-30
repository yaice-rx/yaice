package tcp

import (
	"errors"
	"github.com/golang/protobuf/proto"
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
	isClosed     bool
	guid         uint64
	times        int64
	sendQueue    chan []byte
	pkg          network.IPacket
	conn         *net.TCPConn
	receiveQueue chan network.TransitData
	serve        interface{}
	data         interface{}
}

func NewConn(serve interface{}, conn *net.TCPConn, pkg network.IPacket) network.IConn {
	conn_ := &Conn{
		serve:        serve,
		guid:         utils.GenSonyflake(),
		conn:         conn,
		pkg:          pkg,
		isClosed:     false,
		receiveQueue: make(chan network.TransitData, 10),
		sendQueue:    make(chan []byte, 10),
		times:        time.Now().Unix(),
	}
	go func() {
		for data := range conn_.sendQueue {
			_, err := conn_.conn.Write(data)
			if err != nil {
				log.AppLogger.Info("发送参数错误：" + err.Error())
			}
		}
	}()
	go func() {
		for data := range conn_.receiveQueue {
			if data.MsgId != 0 {
				conn_.conn.SetReadDeadline(time.Now().Add(time.Duration(60) * time.Second))
				router.RouterMgr.ExecRouterFunc(conn_, data)
			}
		}
	}()
	return conn_
}

func (c *Conn) Close() {
	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	//断开连接减少对应的连接
	atomic.AddUint32(&c.serve.(*Server).connCount, uint32(int32(-1)))
	//设置当前的句柄为关闭状态
	c.isClosed = true
	//关闭接收通道
	close(c.receiveQueue)
	//关闭发送通道
	close(c.sendQueue)
}

func (c *Conn) GetGuid() uint64 {
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
		log.AppLogger.Error("发送消息时，序列化失败 : "+err.Error(), zap.Int64("MessageId", protoId))
		return err
	}
	if c.isClosed == true {
		log.AppLogger.Info("send msg(proto) channel It has been closed ... ")
		return errors.New("send msg(proto) channel It has been closed ... ")
	}
	c.sendQueue <- c.pkg.Pack(network.TransitData{protoId, data})
	return nil
}

//发送组装好的协议，但是加密始终是在组装包的时候完成加密功能
func (c *Conn) SendByte(message []byte) error {
	if c.isClosed == true {
		log.AppLogger.Info("send msg(byte[]) channel It has been closed ... ")
	}
	c.sendQueue <- message
	return nil
}

func (c *Conn) Start() {
	for {
		if c.isClosed == true {
			log.AppLogger.Info("conn It has been closed ... ")
			return
		}
		//1 先读出流中的head部分
		headData := make([]byte, c.pkg.GetHeadLen())
		_, err := io.ReadFull(c.conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			if err != io.EOF {
				log.AppLogger.Info("network io read data err:" + err.Error())
			}
			break
		}
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
			msgData, err, func_ := c.pkg.Unpack(contentData)
			if msgData == nil {
				log.AppLogger.Info("network io read data err - 1:" + err.Error())
				break
			}
			if func_ != nil {
				func_(c)
			}
			//写入通道数据
			c.receiveQueue <- network.TransitData{
				MsgId: msgData.GetMsgId(),
				Data:  msgData.GetData(),
			}
		}
	}
}
