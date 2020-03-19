package config

import "sync"

type Config struct {
	OutPort     int    //外部连接端口
	InPort      int    //内部连接端口
	Pid         string //服务进程编号
	InHost      string //内部连接ip
	TypeId      string //服务类型
	ServerGroup string //服务分组
	OutHost     string //外部连接ip
}

type IConfig interface {
	SetPid(pid string)
	GetPid() string
	SetTypeId(tId string)
	GetTypeId() string
	SetServerGroup(group string)
	GetServerGroup() string
	SetInHost(host string)
	GetInHost() string
	SetInPort(port int)
	GetInPort() int
	SetOutHost(host string)
	GetOutHost() string
	SetOutPort(port int)
	GetOutPort() int
}

var configMgr IConfig

var mu sync.Mutex

func ConfInstance() IConfig {
	mu.Lock()
	defer mu.Unlock()
	if configMgr == nil {
		configMgr = newConfig()
	}
	return configMgr
}

func newConfig() IConfig {
	return &Config{}
}

func (c *Config) SetPid(pid string) {
	c.Pid = pid
}

func (c *Config) GetPid() string {
	return c.Pid
}

func (c *Config) SetTypeId(tId string) {
	c.TypeId = tId
}

func (c *Config) GetTypeId() string {
	return c.TypeId
}

func (c *Config) SetServerGroup(group string) {
	c.ServerGroup = group
}

func (c *Config) GetServerGroup() string {
	return c.ServerGroup
}

func (c *Config) SetInHost(host string) {
	c.InHost = host
}

func (c *Config) GetInHost() string {
	return c.InHost
}

func (c *Config) SetInPort(port int) {
	c.InPort = port
}

func (c *Config) GetInPort() int {
	return c.InPort
}

func (c *Config) SetOutHost(host string) {
	c.OutHost = host
}
func (c *Config) GetOutHost() string {
	return c.InHost
}

func (c *Config) SetOutPort(port int) {
	c.OutPort = port
}

func (c *Config) GetOutPort() int {
	return c.InPort
}
