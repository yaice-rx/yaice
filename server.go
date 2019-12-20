package yaice

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/yaice-rx/yaice/cluster"
	"github.com/yaice-rx/yaice/cron"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/resource"
	"github.com/yaice-rx/yaice/router"
	"os"
)

type IServer interface {
	//适配网络
	AdaptationNetwork(network string)
	//添加路由
	AddRouter(message proto.Message, handler func(conn network.IConn, content []byte))
	//开启业务服务方法
	Serve() error
	//停止服务器方法
	Stop()
}

//服务运行状态
var running = make(chan bool, 1)

type yaice struct {
	_RouterMgr     router.IRouter  //路由配置
	_Network       network.IServer //适配网络
	_Cron          cron.ICron
	_ServiceResMgr *resource.ServiceResource //资源配置
	_Cluster       cluster.ICluster
}

func NewServer(typeId string, groundId string, allowConn bool) IServer {
	server := &yaice{
		_Cron:          cron.CronMgr,
		_RouterMgr:     router.RouterMgr,       //路由配置
		_ServiceResMgr: resource.ServiceResMgr, //系统资源配置
		_Cluster:       cluster.ClusterMgr,     //集群服务
	}
	//更新服务配置文件
	cluster.ServerConfMgr.Pid = os.Getpid()
	cluster.ServerConfMgr.TypeId = typeId
	cluster.ServerConfMgr.GroupId = groundId
	cluster.ServerConfMgr.AllowConnect = allowConn
	cluster.ServerConfMgr.OutHost = server._ServiceResMgr.ExtranetHost
	cluster.ServerConfMgr.InHost = server._ServiceResMgr.IntranetHost
	return server
}

//适配网络
func (this *yaice) AdaptationNetwork(network string) {
	switch network {
	case "tcp":
		this._Network = tcp.TcpServerMgr
		cluster.ServerConfMgr.Network = this._Network.GetNetworkName()
		break
	case "kcp":
		break
	case "raknet":
		break
	default:
		break
	}
}

func (this *yaice) AddRouter(message proto.Message, handler func(conn network.IConn, content []byte)) {
	this._RouterMgr.AddRouter(message, handler)
}

//启动服务
func (this *yaice) Serve() error {
	//开启外网运行
	if this._Network == nil {
		portChan := make(chan int)
		go this._Network.Start(portChan)
		port := <-portChan
		if port <= 0 {
			return errors.New("port listen fail")
		}
		close(portChan)
		cluster.ServerConfMgr.OutPort = port
	}
	//开启内网运行
	this._Cluster.Start()
	//注册配置中心数据
	this._Cluster.RegisterServerConfData()
	//退出运行
	<-running
	return nil
}

//停止服务器方法
func (this *yaice) Stop() {
	this._Cron.Stop()
	this._RouterMgr.Stop()
}
