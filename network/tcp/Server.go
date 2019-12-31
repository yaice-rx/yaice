package tcp

import (
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/network"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	sync.Mutex
	network  string
	listener *net.TCPListener
}

var ServerMgr = _NewServer()

func _NewServer() network.IServer {
	return &Server{}
}

func (s *Server) Listen(startPort int, endPort int) int {
	port := make(chan int)
	defer close(port)
	for i := startPort; i < endPort; i++ {
		go func() {
			tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(i))
			if nil != err {
				port <- -1
				logrus.Debug("tcp resolve address :", i, " fail,error :", err)
				return
			}
			listener, err := net.ListenTCP("tcp", tcpAddr)
			if nil != err {
				port <- -1
				logrus.Debug("tcp listen port :", i, " fail,error :", err)
				return
			}
			logrus.Debug("listen port", i)
			port <- i
			s.listener = listener
			for {
				tcpConn, err := listener.AcceptTCP()
				if nil != err || nil == tcpConn {
					continue
				}
				if ConnManagerMgr.Len() > 5000 {
					tcpConn.Close()
				} else {
					tcpConn.SetReadDeadline(time.Now().Add(5 * time.Second))
					conn := NewConn(tcpConn)
					ConnManagerMgr.Add(conn)
					conn.Start()
				}
			}
		}()
		portData := <-port
		if portData > 0 {
			return portData
		}
	}
	return -1
}

func (s *Server) Close() {
	s.listener.Close()
}
