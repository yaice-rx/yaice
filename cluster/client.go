package cluster

import (
	"encoding/json"
	"github.com/yaice-rx/yaice/job"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	proto_ "github.com/yaice-rx/yaice/proto"
	"github.com/yaice-rx/yaice/router"
	"sync"
)

//集群-客户端
type IClusterClient interface {
	//注册路由
	registerRouter()
	//连接服务
	connectServices()
}

//集群-客户端
type clusterClient struct {
	sync.Mutex
	//客户端
	client network.IClient
	//连接服务的数组	 map[服务类型]map[进程id]连接句柄
	ClientList map[string]map[string]network.IConnect
}

var ClusterClientMgr = newClusterClient()

/**
 * 初始化客户端
 */
func newClusterClient() IClusterClient {
	this := &clusterClient{
		client:     tcp.ClientMgr,
		ClientList: make(map[string]map[string]network.IConnect),
	}
	this.registerRouter()
	this.connectServices()
	return this
}

/**
 * 连接服务
 */
func (this *clusterClient) connectServices() {
	for _, data := range ClusterEtcdMgr.GetData() {
		//集群配置文件
		var config clusterConf
		if json.Unmarshal(data, &config) != nil {
			break
		}
		//连接服务句柄
		conn := this.client.Connect(config.InHost, config.InPort)
		if conn != nil {
			break
		}
		//发送服务关联协议数据
		protoData := proto_.C2SServiceAssociate{TypeName: clusterConfMgr.TypeName, Pid: clusterConfMgr.Pid}
		data, err := json.Marshal(protoData)
		if err != nil {
			break
		}
		err = conn.Send(data)
		if err != nil {
			break
		}
	}
}

/**
 * 注册路由方法
 */
func (this *clusterClient) registerRouter() {
	router.RouterMgr.RegisterInternalRouterFunc(&proto_.S2CServiceAssociate{}, this.serviceAssociateFunc)
}

func (this *clusterClient) serviceAssociateFunc(conn network.IConnect, content []byte) {
	var data proto_.C2SServiceAssociate
	if json.Unmarshal(content, &data) != nil {
		return
	}
	//添加服务到列表中
	this.Lock()
	defer this.Unlock()
	var connect map[string]network.IConnect
	connect[data.Pid] = conn
	this.ClientList[data.TypeName] = connect
	//心跳
	job.Crontab.AddCronTask(10, -1, func() {
		data, _ := json.Marshal(proto_.C2SServicePing{})
		if conn.Send(data) != nil {
			return
		}
	})
}
