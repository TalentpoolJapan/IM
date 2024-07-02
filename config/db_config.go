package config

import (
	"fmt"
	"github.com/gogf/gf/v2/container/gmap"
	"imserver/models"
	"xorm.io/xorm"
)

var (
	MemDB          = gmap.New(true)
	MysqlDB, _     = xorm.NewEngine("mysql", fmt.Sprintf("root:%s@%s/%s?charset=utf8", MYSQL_SECRECT, MYSQL_HOST, MYSQL_DB))
	ManticoreDB, _ = xorm.NewEngine("mysql", fmt.Sprintf("``:``@tcp(%s)/Manticore", "127.0.0.1:9306"))
	NoSqlDB, _     = models.NewNoSqlDB("IMKVDB")
)
