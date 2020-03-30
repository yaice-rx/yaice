package tcp

import (
	"github.com/yaice-rx/yaice/network"
	"net"
	"strconv"
	"sync"
)

type Server struct {
	sync.Mutex
	network  string
	listener *net.TCPListener
}

func NewServer() network.IServer {
	return &Server{}
}

func (s *Server) Listen(packet network.IPacket, startPort int, endPort int, noticeHandler func(conn network.IConn)) int {
	port := make(chan int)
	defer close(port)
	for i := startPort; i < endPort; i++ {
		go func() {
			tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(i))
			if nil != err {
				port <- -1
				return
			}
			listener, err := net.ListenTCP("tcp", tcpAddr)
			if nil != err {
				port <- -1
				return
			}
			port <- i
			s.listener = listener
			for {
				tcpConn, err := listener.AcceptTCP()
				if nil != err || nil == tcpConn {
					continue
				}
				if network.ConnManagerInstance().GetLen() > 5000 {
					tcpConn.Close()
				} else {
					conn := NewConn(s, tcpConn, packet)
					network.ConnManagerInstance().Modify(conn.GetGuid(), conn)
					go conn.Start(noticeHandler)
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
