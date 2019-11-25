package cluster

import (
	"encoding/json"
	"sync"
	"yaice/config"
	"yaice/job"
	"yaice/network"
	"yaice/network/tcp"
	proto_ "yaice/proto"
	"yaice/router"
)

//集群客户端接口
type IClusterClient interface {
}

//集群客户端
type clusterClient struct {
	sync.Mutex
	//连接服务的数组	 map[服务类型]map[进程id]连接句柄
	ModuleClientMap map[string]map[string]network.IConnect
}

var ClusterClientMgr = newClusterClient()

/**
 * 初始化客户端
 */
func newClusterClient()IClusterClient{
	mgr := &clusterClient{
		ModuleClientMap:make(map[string]map[string]network.IConnect),
	}
	router.RouterMgr.RegisterInternalRouterFunc(&proto_.S2CServiceAssociate{}, func(conn network.IConnect,content []byte) {
		//收到消息
		var data proto_.C2SServiceAssociate
		if json.Unmarshal(content,&data) != nil{
			return
		}
		//添加服务到列表中
		var connect map[string]network.IConnect
		connect[data.Pid] = conn
		mgr.ModuleClientMap[data.TypeName] = connect
		//心跳
		job.Crontab.AddCronTask(10,-1, func() {
			data,_ := json.Marshal(proto_.C2SServicePing{})
			if conn.Send(data) != nil {
				return
			}
		})
		return
	})
	mgr.connect()
	return mgr
}

/**
 * 连接服务
 */
func (this *clusterClient)connect(){
	for _,data := range ClusterEtcdMgr.GetData(){
		var moduleConfig config.ModuleConfig
		if json.Unmarshal(data,&moduleConfig) != nil{
			break
		}
		conn := tcp.ClientMgr.Connect(
			moduleConfig.InHost,
			moduleConfig.InPort,
			)
		protoData := proto_.C2SServiceAssociate{
			TypeName:config.ModuleConfigMgr.TypeName,
			Pid:config.ModuleConfigMgr.Pid,
		}
		data,err := json.Marshal(protoData)
		if err != nil{
			break
		}
		err = conn.Send(data)
		if err != nil{
			break
		}
	}
}






