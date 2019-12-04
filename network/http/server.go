package http

import (
	"github.com/yaice-rx/yaice/network"
	"net/http"
	"strconv"
	"sync"
)

type httpServer struct {
	sync.Mutex
	server *http.Server
}

var HttpServerMgr = newHttpServer()

func newHttpServer() network.IServer {
	srv := &httpServer{}
	srv.server = &http.Server{}
	return nil
}

func (this *httpServer) AddRouter() {
	this.server.Close()
}

func (this *httpServer) Start(port int) error {
	//开启监听
	this.server.Addr = ":" + strconv.Itoa(port)
	err := this.server.ListenAndServe()
	return err
}

func (this *httpServer) Close() {
	this.server.Close()
}
