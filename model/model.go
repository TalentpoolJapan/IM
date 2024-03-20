package model

import (
	"imserver/db/fulltext"
	"imserver/db/nosql"
)

type Model struct {
	FulltextDB *fulltext.FulltextDB
	NoSqlDB    *nosql.NoSqlDB
}
