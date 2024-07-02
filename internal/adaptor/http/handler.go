package http

import (
	"github.com/gin-gonic/gin"
	"imserver/config"
	"imserver/internal/application/user"
)

func GetMyContacts(engine *gin.Context) {
	// 1. get query
	uuid := engine.Query("uuid")
	qry := &user.GetMyContactsQry{
		Uuid: uuid,
	}

	// 2. get my contacts
	myContacts := config.UserAppServ.GetMyContacts(qry)
	engine.JSON(200, myContacts)
}
