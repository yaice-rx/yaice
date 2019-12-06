package http

import (
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/resource"
	router_ "github.com/yaice-rx/yaice/router"
	"net/http"
	"strconv"
	"sync"
)

type httpServer struct {
	sync.Mutex
	server  *http.Server
	network string
}

var HttpServerMgr = newHttpServer()

func newHttpServer() network.IServer {
	srv := &httpServer{
		network: "http",
	}
	srv.server = &http.Server{}
	return srv
}

func (this *httpServer) GetNetwork() string {
	return this.network
}

func (this *httpServer) Start() (int, error) {
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

func (this *httpServer)Run(){
	go this.server.ListenAndServe()
}

func (this *httpServer) Close() {
	this.server.Close()
}
