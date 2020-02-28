package db

import (
	"github.com/yaice-rx/yaice/db/mongo"
	"github.com/yaice-rx/yaice/db/mysql"
)

type IDBDrive interface {
	Connect(host string, port int)
}

//选择驱动
func DBDrive(driveName string) IDBDrive {
	switch driveName {
	case "Mongo":
		return &mongo.Mongo{}
		break
	case "Mysql":
		return &mysql.Mysql{}
		break
	}
	return nil
}
