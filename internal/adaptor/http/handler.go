package http

import (
	"github.com/gin-gonic/gin"
	"imserver/config"
	"imserver/internal/application"
	"imserver/internal/application/user"
	"imserver/models"
	"imserver/util"
	"net/http"
)

func GetMyContacts(c *gin.Context) {
	userToken, err := checkAuth(c)
	if err != nil {
		c.JSON(401, application.SingleRespFail[any]("auth failed"))
		return
	}
	// 1. get query
	qry := &user.GetMyContactsQry{
		Uuid: userToken.Uuid,
	}

	// 2. get my contacts
	myContacts := config.UserAppServ.GetMyContacts(qry)
	c.JSON(http.StatusOK, NewApiRestResult(RestResult{Code: 0, Message: myContacts.Msg, Data: myContacts.Data}))
}

func checkAuth(c *gin.Context) (models.UserToken, error) {
	auth := c.Query("token")

	if auth == "" {
		auth = c.Copy().GetHeader("Authorization")
	}

	return util.CheckAuthHeader(auth)
}
