package main

import (
	"fmt"
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

	//ct.M.FulltextDB.Exec("alter table im_message drop column sessionid")
	//ct.M.FulltextDB.Exec("alter table im_message add column sessionid text")
	// _, err := ct.M.FulltextDB.Exec(`CREATE TABLE im_friend_list (
	// 	id bigint,
	// 	touser text,
	// 	fromuser text,
	// 	isblack integer,
	// 	status integer,
	// 	created bigint)`)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// res, err := ct.GetDescribe("im_friend_list")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(res)

	_, err := ct.M.CheckOrSetFriends(model.ImFreindList{
		Touser:   "ddddccbaaaodfgvbhnjjkkl",
		Fromuser: "ddddccbaaajfghjklgthjkl",
	})
	if err != nil {
		fmt.Println(err)
	}

	msgs, err := ct.GetAllFreinds()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(msgs)

	ct.M.NoSqlDB.TRUNCATE()
	// err = ct.SetLastReadId(model.SetLastReadId{
	// 	Touser:   "odfgvbhnjjkkl",
	// 	Fromuser: "jfghjklgthjkl",
	// 	Id:       2596660176354279464,
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// v, err := ct.M.GetLastReadId(model.GetLastReadId{
	// 	Touser:   "odfgvbhnjjkkl",
	// 	Fromuser: "jfghjklgthjkl",
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(v)
	// for i := 0; i < 20; i++ {
	// 	err = ct.InsertMessages(model.ImMessage{
	// 		Touser:   "odfgvbhnjjkkl",
	// 		Fromuser: "jfghjklgthjkl",
	// 		Msg:      "msg",
	// 		Msgtype:  1,
	// 		Fromtype: 0,
	// 		Totype:   1,
	// 		Created:  time.Now().UnixNano(),
	// 		Msgid:    "2345t6y7rfedgthj",
	// 	})
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	// msgs, err := ct.GetAllMessages()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(msgs)
	// msgs, _ := ct.GetMessagesByStartId(model.MessagesByStartId{
	// 	Touser:   "odfgvbhnjjkkl",
	// 	Fromuser: "jfghjklgthjkl",
	// 	Method:   "new",
	// 	Id:       2596660176354279464,
	// })
	// fmt.Println(msgs)
	//ct.ClearAllMessages()

}
