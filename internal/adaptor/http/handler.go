package http

import (
	"github.com/gin-gonic/gin"
	"imserver/config"
	"imserver/internal/application/user"
	"imserver/models"
	"imserver/util"
	"net/http"
)

func GetMyContacts(c *gin.Context) {
	userToken, err := checkAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewApiRestResult(RestResult{Code: 0, Message: "auth failed"}))
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

func GetUnreadMessageState(c *gin.Context) {
	userToken, err := checkAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewApiRestResult(RestResult{Code: 0, Message: "auth failed"}))
		return
	}

	qry := &user.UnreadMessageStateQry{
		Uuid: userToken.Uuid,
	}
	unreadMessageState := config.UserAppServ.GetUnreadMessageState(qry)
	c.JSON(http.StatusOK, NewApiRestResult(RestResult{Code: 0, Message: unreadMessageState.Msg, Data: unreadMessageState.Data}))
}

func SyncLastReadClientMsgId(c *gin.Context) {
	userToken, err := checkAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewApiRestResult(RestResult{Code: 0, Message: "auth failed"}))
		return
	}

	// 1. get query
	qry := &user.SyncLastReadClientMsgIdCmd{
		Uuid: userToken.Uuid,
	}
	err = c.ShouldBindJSON(qry)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewApiRestResult(RestResult{Code: 0, Message: "param error"}))
		return
	}

	// 2. sync last read client msg id
	syncLastReadClientMsgId := config.UserAppServ.SyncLastReadClientMsgId(qry)
	c.JSON(http.StatusOK, NewApiRestResult(RestResult{Code: 0, Message: syncLastReadClientMsgId.Msg, Data: syncLastReadClientMsgId.Data}))
}

func ListBeforeImMessage(c *gin.Context) {
	userToken, err := checkAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewApiRestResult(RestResult{Code: 0, Message: "auth failed"}))
		return
	}

	friendUuid := c.Query("friend_uuid")
	clientMsgId := c.Query("client_msg_id")
	qry := &user.ListImMessageBeforeClientMsgQry{
		Uuid:        userToken.Uuid,
		FriendUuid:  friendUuid,
		ClientMsgId: clientMsgId,
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, NewApiRestResult(RestResult{Code: 0, Message: "param error"}))
		return
	}

	msgs := config.UserAppServ.ListImMessageBeforeClientMsgId(qry)
	c.JSON(http.StatusOK, NewApiRestResult(RestResult{Code: 0, Message: msgs.Msg, Data: msgs.Data}))
}

func ListAfterImMessage(c *gin.Context) {
	userToken, err := checkAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, NewApiRestResult(RestResult{Code: 0, Message: "auth failed"}))
		return
	}

	friendUuid := c.Query("friend_uuid")
	clientMsgId := c.Query("client_msg_id")
	qry := &user.ListImMessageAfterClientMsgQry{
		Uuid:        userToken.Uuid,
		FriendUuid:  friendUuid,
		ClientMsgId: clientMsgId,
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, NewApiRestResult(RestResult{Code: 0, Message: "param error"}))
		return
	}

	msgs := config.UserAppServ.ListImMessageAfterClientMsgId(qry)
	c.JSON(http.StatusOK, NewApiRestResult(RestResult{Code: 0, Message: msgs.Msg, Data: msgs.Data}))
}

func checkAuth(c *gin.Context) (models.UserToken, error) {
	auth := c.Query("token")

	if auth == "" {
		auth = c.Copy().GetHeader("Authorization")
	}

	return util.CheckAuthHeader(auth)
}
