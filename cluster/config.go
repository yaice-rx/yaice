package cluster

type clusterConf struct {
	Pid                int      //服务进程编号
	TypeId             string   //服务类型
	GroupId            string   //服务组编号
	ConnectServiceList []string //连接服务列表
	OutHost            string   //外部连接ip
	OutPort            int      //外部连接端口
	InHost             string   //内部连接ip
	InPort             int      //内部连接端口
	Network            string
}

var ClusterConfMgr = newClusterConf()

func newClusterConf() *clusterConf {
	return &clusterConf{}
}