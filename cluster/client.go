package cluster

import (
	"github.com/sirupsen/logrus"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	proto_ "github.com/yaice-rx/yaice/proto"
	"sync"
)

//集群-客户端
type IClusterClient interface {
	Run()
}

//集群-客户端
type ClusterClient struct {
	sync.Mutex
	_NetworkType network.IClient //选择网络连接方式
	_MsgDataChan chan *ClusterConf
	ConnManager  network.IConnManager
}

//集群客户端
var ClusterClientMgr = _NewClusterClient()

func _NewClusterClient() IClusterClient {
	this := &ClusterClient{
		_NetworkType: tcp.TCPClientMgr,
		_MsgDataChan: make(chan *ClusterConf),
		ConnManager:  tcp.NewConnManager(),
	}
	return this
}

func (this *ClusterClient) Run() {
	go ClusterServiceManagerMgr.Watch(this._MsgDataChan)
	go func() {
		for {
			select {
			case data := <-this._MsgDataChan:
				dealConn := this._NetworkType.Connect(data.InHost, data.InPort)
				if dealConn == nil {
					continue
				}
				if data.Pid == ClusterConfMgr.Pid {
					continue
				}
				this.ConnManager.Add(dealConn)
				dealConn.Start()
				//发送消息
				protoData := proto_.C2SServiceAssociate{TypeName: ClusterConfMgr.TypeId, Pid: int64(ClusterConfMgr.Pid)}
				err := dealConn.SendMsg(&protoData)
				if err != nil {
					logrus.Debug(dealConn, "发送消息失败，", err.Error())
				}
				break
			}
		}
	}()
	//开启连接
	for _, data := range ClusterServiceManagerMgr.GetClusterServiceData() {
		//首先判读服务是否属于自己
		if data.TypeId == ClusterConfMgr.TypeId {
			logrus.Debug("类型[" + data.TypeId + "]相同，不能连接")
			continue
		}
		if data.AllowConnect {
			//连接服务句柄
			dealConn := this._NetworkType.Connect(data.InHost, data.InPort)
			if dealConn == nil {
				logrus.Debug(data.InHost, "：", data.InPort, "连接失败")
				continue
			}
			this.ConnManager.Add(dealConn)
			dealConn.Start()
			//发送消息
			protoData := proto_.C2SServiceAssociate{TypeName: ClusterConfMgr.TypeId, Pid: int64(ClusterConfMgr.Pid)}
			err := dealConn.SendMsg(&protoData)
			if err != nil {
				logrus.Debug(dealConn, "发送消息失败，", err.Error())
			}
		}
	}
}
