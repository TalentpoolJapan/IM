package main

import (
	"fmt"
	"imserver/config"
	"imserver/controllers"
	"imserver/models"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gogf/gf/v2/container/gmap"
	"xorm.io/xorm"
)

var (
	memDB          = gmap.New(true)
	mysqlDB, _     = xorm.NewEngine("mysql", fmt.Sprintf("root:%s@%s/%s?charset=utf8", config.MYSQL_SECRECT, config.MYSQL_HOST, config.MYSQL_DB))
	manticoreDB, _ = xorm.NewEngine("mysql", fmt.Sprintf("``:``@tcp(%s)/Manticore", "127.0.0.1:9306"))
	noSqlDB, _     = models.NewNoSqlDB("IMKVDB")

	ct = &controllers.Controller{
		M: &models.Model{
			ManticoreDB: manticoreDB,
			NoSqlDB:     noSqlDB,
			MemDB:       memDB,
			MySQLDB:     mysqlDB,
		},
		Debug: true,
	}
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization,TalentPool-Language,token")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func main() {
	//初始化全局内存数据结构
	initializeData()

	//WS服务开始
	r := gin.New()
	r.Use(Cors())
	r.Use(gin.Recovery())
	r.GET("/ws", ct.WS)
	//r.GET("/wst", ct.MockWs)
	r.Run(":8888")
}

func initializeData() {
	ct.M.MemDB.Set("Profile", gmap.New(true))
	ct.M.MemDB.Set("IsInitUser", gmap.New(true))
}
