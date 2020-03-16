package tcp

import (
	"github.com/yaice-rx/yaice/network"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
)

type Server struct {
	sync.Mutex
	network  string
	listener *net.TCPListener
}

var ServerMgr = _NewServer()

var ServerConnectNumber uint32 = 0

func _NewServer() network.IServer {
	return &Server{}
}

func (s *Server) Listen(packet network.IPacket, startPort int, endPort int) int {
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
				if atomic.LoadUint32(&ServerConnectNumber) > 5000 {
					tcpConn.Close()
				} else {
					conn := NewConn(s, tcpConn, packet)
					atomic.AddUint32(&ServerConnectNumber, 1)
					go conn.ReadThread()
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
