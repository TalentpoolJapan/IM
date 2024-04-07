package model

import (
	"imserver/db/fulltext"
	mysqldb "imserver/db/mysql"
	"imserver/db/nosql"

	"github.com/gogf/gf/v2/container/gmap"
)

type Model struct {
	FulltextDB *fulltext.FulltextDB
	NoSqlDB    *nosql.NoSqlDB
	MemDB      *gmap.AnyAnyMap
	MysqlDB    *mysqldb.MysqlDB
}
