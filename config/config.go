package config

type Config struct {
	Pid         int    //服务进程编号
	TypeId      string //服务类型
	ServerGroup string //服务分组
	OutHost     string //外部连接ip
	OutPort     int    //外部连接端口
	InHost      string //内部连接ip
	InPort      int    //内部连接端口
}
