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
	connCount uint32
	network   string
	opt       network.IOptions
	listener  *net.TCPListener
}

func NewServer() network.IServer {
	return &Server{
		connCount: 0,
	}
}

func (s *Server) Listen(packet network.IPacket, startPort int, endPort int, opt_ network.IOptions) int {
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
				if opt_.GetMaxConnCount() > atomic.LoadUint32(&s.connCount) {
					opt_.CallBackFunc()(tcpConn)
					return
				}
				atomic.AddUint32(&s.connCount, 1)
				conn := NewConn(s, tcpConn, packet)
				go conn.Start()
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
