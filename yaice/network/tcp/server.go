package tcp

import (
	"io"
	"net"
	"strconv"
	"sync"
	"yaice/network"
	"yaice/utils"
)

type server struct {
	sync.Mutex
	listener        *net.TCPListener
	ReceiveMsgQueue chan *network.MsgQueue
	ConnectsMgr     network.IConnects
}

var TcpServerMgr  = newTcpServer()

func newTcpServer()network.IServer{
	mgr := &server{
		ReceiveMsgQueue: make(chan *network.MsgQueue),
		ConnectsMgr:     ConnectsMgr,
	}
	mgr.ConnectsMgr.RunThread()
	return	mgr
}

/**
 * 监听网络接口
 */
func (this *server)Listen(port int)error{
	tcpAddr,err := net.ResolveTCPAddr("tcp",":"+strconv.Itoa(port))
	if nil != err{
		return err
	}
	listener,err := net.ListenTCP("tcp",tcpAddr)
	if nil != err{
		return err
	}
	this.listener = listener
	return nil
}

/**
 * 关闭网络接口
 */
func (this *server)Close(){
	this.Lock()
	defer this.Unlock()
	this.listener.Close()
}

/**
 * 处理conn
 */
func (this *server)Accept(){
	go func() {
		for {
			conn, err := this.listener.AcceptTCP()
			if nil != err || nil == conn {
				continue
			}
			go this.handle(conn)
		}
	}()
}

/**
 * 处理消息
 */
func (this *server)handle(conn *net.TCPConn){
	for {
		//read
		var buffer = make([]byte, 1024)
		n, e := conn.Read(buffer)
		if e != nil {
			if e == io.EOF {
				break
			}
			break
		}
		msgId := utils.BytesToInt(buffer[:4])
		this.ReceiveMsgQueue <- network.NewMsgQueue(msgId,NewConnect(conn),buffer[4:n])
	}
}
