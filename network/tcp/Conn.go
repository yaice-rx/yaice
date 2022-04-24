package tcp

import (
	"errors"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/router"
	"github.com/yaice-rx/yaice/utils"
	"go.uber.org/zap"
)

type Conn struct {
	type_        network.ServeType
	isClosed     bool
	guid         uint64
	times        int64
	pkg          network.IPacket
	conn         *net.TCPConn
	serve        interface{}
	data         interface{}
	isPos        int64
	sendQueue    chan []byte
	receiveQueue chan network.TransitData
}

func NewConn(serve interface{}, conn *net.TCPConn, pkg network.IPacket, type_ network.ServeType) network.IConn {
	conn_ := &Conn{
		type_:        type_,
		serve:        serve,
		guid:         utils.GenSonyflakeToo(),
		conn:         conn,
		pkg:          pkg,
		isClosed:     false,
		receiveQueue: make(chan network.TransitData, 5000),
		sendQueue:    make(chan []byte, 5000),
		times:        time.Now().Unix(),
	}
	go func() {
		for data := range conn_.sendQueue {
			LOOP:
				_, err := conn_.conn.Write(data)
				//判断客户端，如果不是主动关闭，而是网络抖动的时候 多次连接
				if err != nil {
					log.AppLogger.Info("发送参数错误：" + err.Error())
					if conn_.serve.(*TCPClient).dialRetriesCount <= conn_.serve.(*TCPClient).opt.GetMaxRetires(){
						conn_.serve.(*TCPClient).dialRetriesCount += 1
						goto LOOP
					}else{
						conn_.serve.(*TCPClient).Close()
					}
				}else{
					log.AppLogger.Info("发送失败，原因：" + err.Error())
				}
		}
	}()
	go func() {
		for data := range conn_.receiveQueue {
			if data.MsgId != 0 {
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
	if c.type_ == network.Serve_Server {
		//断开连接减少对应的连接
		atomic.AddInt32(&c.serve.(*Server).connCount, int32(-1))
	}
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
		log.AppLogger.Error("发送消息时，序列化失败 : "+err.Error(), zap.Int32("MessageId", protoId))
		return err
	}
	if c.isClosed == true {
		log.AppLogger.Info("send msg(proto) channel It has been closed ... ")
		return errors.New("send msg(proto) channel It has been closed ... ")
	}
	c.sendQueue <- c.pkg.Pack(network.TransitData{protoId, data}, c.isPos)
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
		//读取限时
		if err := c.conn.SetReadDeadline(time.Now().Add(time.Minute * 70)); err != nil {
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
				if err != nil {
					log.AppLogger.Info("network io read data err - 1:" + err.Error())
				}
				continue
			}
			if func_ != nil {
				func_(c)
			}
			c.isPos = msgData.GetIsPos()
			//写入通道数据
			c.receiveQueue <- network.TransitData{
				MsgId: msgData.GetMsgId(),
				Data:  msgData.GetData(),
			}
		}
		if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
			return
		}
	}
}
