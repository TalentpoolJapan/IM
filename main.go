package main

import (
	"imserver/controller"
	"imserver/db/fulltext"
	"imserver/db/nosql"
	"imserver/model"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gogf/gf/v2/container/gmap"
	"golang.org/x/sync/errgroup"
)

var (
	fulltextDb, _ = fulltext.NewFulltextDB("127.0.0.1:9306")
	noSqlDB, _    = nosql.NewNoSqlDB("IMKVDB")
	memDB         = gmap.New(true)

	ct = &controller.Controller{
		M: &model.Model{
			FulltextDB: fulltextDb,
			NoSqlDB:    noSqlDB,
			MemDB:      memDB,
		},
	}

	g errgroup.Group
)

func WsService() http.Handler {
	e := gin.New()
	e.Use(gin.Recovery())
	e.GET("/ws", ct.WsHandler)
	return e
}

func SseService() http.Handler {
	e := gin.New()
	e.Use(gin.Recovery())
	e.GET("/TODSSE", ct.WsHandler)
	return e
}

func main() {
	server01 := &http.Server{
		Addr:         ":39260",
		Handler:      WsService(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server02 := &http.Server{
		Addr:         ":8081",
		Handler:      SseService(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.Go(func() error {
		return server01.ListenAndServe()
	})

	g.Go(func() error {
		return server02.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	// ct := &controller.Controller{
	// 	M: &model.Model{
	// 		FulltextDB: fulltextDb,
	// 		NoSqlDB:    noSqlDB,
	// 		MemDB:      memDB,
	// 	},
	// }

	// wsService := &http.Server{
	// 	Addr: ":8080",
	// 	Handler:
	// }
	// ct := &controller.Controller{
	// 	M: &model.Model{
	// 		FulltextDB: fulltextDb,
	// 		NoSqlDB:    noSqlDB,
	// 		MemDB:      memDB,
	// 	},
	// }
	// ct.M.FulltextDB.Exec("alter table im_friend_list add column count integer")
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

	// _, err := ct.M.CheckOrSetFriends(model.ImFreindList{
	// 	Touser:   "ddddccbaaaodfgvbhnjjkkl",
	// 	Fromuser: "ddddccbaaajfghjklgthjkl",
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// msgs, err := ct.GetAllFreinds()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(msgs)

	// ct.M.NoSqlDB.TRUNCATE()
	// ct.M.FulltextDB.TRUNCATE(`im_friend_list`)
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
