package http

import (
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/resource"
	router_ "github.com/yaice-rx/yaice/router"
	"net/http"
	"strconv"
	"sync"
)

type HTTPServer struct {
	sync.Mutex
	server  *http.Server
	network string
}

var HttpServerMgr = newHttpServer()

func newHttpServer() network.IServer {
	srv := &HTTPServer{
		network: "http",
	}
	srv.server = &http.Server{}
	return srv
}

func (this *HTTPServer) GetNetworkName() string {
	return this.network
}

func (this *HTTPServer) Start() (int, error) {
	mux := http.NewServeMux()
	if router_.RouterMgr.GetHttpHandlerCount() > 0 {
		for r_key, handler := range router_.RouterMgr.GetHttpHandlerMap() {
			mux.HandleFunc(r_key, handler)
		}
	}
	this.server.Handler = mux
	port := resource.ServiceResMgr.HttpPort
	//开启监听
	this.server.Addr = ":" + strconv.Itoa(port)
	return port, nil
}

func (this *HTTPServer) GetConns() network.IConnManager {
	return nil
}

func (this *HTTPServer) Run() {
	go this.server.ListenAndServe()
}

func (this *HTTPServer) Close() {
	this.server.Close()
}
