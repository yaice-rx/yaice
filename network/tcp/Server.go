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
	for i := startPort; i < endPort; i++ {
		go func() {
			tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(i))
			if nil != err {
				logrus.Debug("tcp resolve address :", i, " fail,error :", err)
				return
			}
			listener, err := net.ListenTCP("tcp", tcpAddr)
			if nil != err {
				logrus.Debug("tcp listen port :", i, " fail,error :", err)
				return
			}
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
	}
	data := <-port
	close(port)
	return data
}

func (s *Server) Close() {
	s.listener.Close()
}
