package tcp

import (
	"context"
	"github.com/yaice-rx/yaice/network"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
)

type Server struct {
	sync.Mutex
	type_     network.ServeType
	connCount int32
	listener  *net.TCPListener
	cancel  context.CancelFunc
	ctx    context.Context
}

func NewServer() network.IServer {
	s := &Server{
		type_:     network.Serve_Server,
		connCount: 0,
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.ctx = ctx
	return s
}

func (s *Server) Listen(packet network.IPacket, startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) int {
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
				if isAllowConnFunc != nil {
					if !isAllowConnFunc(tcpConn) {
						continue
					}
				}
				atomic.AddInt32(&s.connCount, 1)
				conn := NewConn(s, tcpConn, packet,nil, network.Serve_Server,s.ctx,s.cancel)
				go conn.Start()
			}
		}()
		portData := <- port
		if portData > 0 {
			return portData
		}
	}
	return -1
}

func (s *Server) Close() {
	s.listener.Close()
}
