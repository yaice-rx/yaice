package cluster

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"sync"
)

//集群服务治理
type iServerManager interface {
	Run() error
	Stop()
}

type serverManager struct {
	sync.Mutex
	_Network network.IServer
}

func _NewServerManager() iServerManager {
	server := &serverManager{
		_Network: tcp.TcpServerMgr,
	}
	return server
}

func (s *serverManager) Run() error {
	//启动监听Server端口
	portChan := make(chan int)
	go s._Network.Start(portChan)
	port := <-portChan
	if port <= 0 {
		logrus.Debug("内部服务启动失败，端口已全部被占用")
		return errors.New("内部服务启动失败，端口已全部被占用")
	}
	ServerConfMgr.InPort = port
	close(portChan)
	return nil
}

//关闭服务
func (s *serverManager) Stop() {
	s._Network.Close()
}
