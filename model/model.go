package model

import (
	"imserver/db/fulltext"
	"imserver/db/nosql"

	"github.com/gogf/gf/v2/container/gmap"
)

type Model struct {
	FulltextDB *fulltext.FulltextDB
	NoSqlDB    *nosql.NoSqlDB
	MemDB      *gmap.AnyAnyMap
}
