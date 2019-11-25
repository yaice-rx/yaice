package config

type ModuleConfig struct {
	Pid                string   //服务进程编号
	TypeName           string   //服务类型
	GroupName          string   //服务组编号
	ConnectServiceList []string //连接服务列表
	OutHost            string   //外部连接ip
	OutPort            int   //外部连接端口
	InHost             string   //内部连接ip
	InPort             int   //内部连接端口
	HttpPort           string
}

var ModuleConfigMgr  = newModuleConfig()

func newModuleConfig()*ModuleConfig{
	return &ModuleConfig{}
}