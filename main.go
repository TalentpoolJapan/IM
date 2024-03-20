package main

import (
	"imserver/controller"
	"imserver/db/fulltext"
	"imserver/db/nosql"
	"imserver/model"

	_ "github.com/go-sql-driver/mysql"
)

var (
	fulltextDb, _ = fulltext.NewFulltextDB("127.0.0.1:9306")
	noSqlDB, _    = nosql.NewNoSqlDB("IMKVDB")
)

func main() {
	ct := &controller.Controller{
		M: &model.Model{
			FulltextDB: fulltextDb,
			NoSqlDB:    noSqlDB,
		},
	}

}
