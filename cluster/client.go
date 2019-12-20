package cluster

import (
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	proto_ "github.com/yaice-rx/yaice/proto"
	"sync"
)

//集群-客户端
type iClientManager interface {
	RegisterData(data *ServerConf) error
	Run() error
	Stop()
}

//集群-客户端
type clientManager struct {
	sync.Mutex
	_Network     network.IClient //选择网络连接方式
	_MsgDataChan chan *ServerConf
	_EtcdManager IEtcdManager
	ConnManager  network.IConnManager
}

func _NewClientManger() iClientManager {
	this := &clientManager{
		_Network:     tcp.TCPClientMgr,
		_MsgDataChan: make(chan *ServerConf),
		ConnManager:  tcp.NewConnManager(),
		_EtcdManager: _NewEtcdManager(),
	}
	return this
}

func (this *clientManager) RegisterData(data *ServerConf) error {
	return this._EtcdManager.RegisterData(data)
}

func (this *clientManager) Run() error {
	go this._EtcdManager.Watch(this._MsgDataChan)
	go func() {
		for {
			select {
			case data := <-this._MsgDataChan:
				dealConn := this._Network.Connect(data.InHost, data.InPort)
				if dealConn == nil {
					continue
				}
				if data.Pid == ServerConfMgr.Pid {
					continue
				}
				this.ConnManager.Add(dealConn)
				dealConn.Start()
				//发送消息
				protoData := proto_.C2SServiceAssociate{TypeName: ServerConfMgr.TypeId, Pid: int64(ServerConfMgr.Pid)}
				err := dealConn.SendMsg(&protoData)
				if err != nil {
					logrus.Debug(dealConn, "发送消息失败，", err.Error())
				}
				break
			}
		}
	}()
	//开启连接
	for _, data := range this._EtcdManager.GetData() {
		//首先判读服务是否属于自己
		if data.TypeId == ServerConfMgr.TypeId {
			logrus.Debug("类型[" + data.TypeId + "]相同，不能连接")
			continue
		}
		if data.AllowConnect {
			//连接服务句柄
			dealConn := this._Network.Connect(data.InHost, data.InPort)
			if dealConn == nil {
				logrus.Debug(data.InHost, "：", data.InPort, "连接失败")
				continue
			}
			this.ConnManager.Add(dealConn)
			dealConn.Start()
			//发送消息
			protoData := proto_.C2SServiceAssociate{TypeName: ServerConfMgr.TypeId, Pid: int64(ServerConfMgr.Pid)}
			err := dealConn.SendMsg(&protoData)
			if err != nil {
				logrus.Debug(dealConn, "发送消息失败，", err.Error())
			}
		}
	}
	return nil
}

func (this *clientManager) Stop() {
	close(this._MsgDataChan)
	this._Network.Stop()
}
