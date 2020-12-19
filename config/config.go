package config


type Config struct {
	Port     	int    	//监听的端口
	ServerGroup string 	//服务分组
	ServerName 	string	//服务名字
	LogPath 	string	//日志文件路径
	LogName 	string	//日志文件名
	ServiceDiscoveryURLs []string //服务管理中心的连接地址
	MsgQueueConnectURL	 string	//消息中心连接地址
}

