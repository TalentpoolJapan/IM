package http

import "github.com/gin-gonic/gin"

func RegisterHandler(engine *gin.Engine) {
	engine.GET("/v1/my_contacts", GetMyContacts)
	engine.GET("/v1/unread_message_state", GetUnreadMessageState)
	engine.GET("/v1/msg_before", ListBeforeImMessage)
	engine.GET("/v1/msg_after", ListAfterImMessage)

	engine.POST("/v1/sync_last_read", SyncLastReadClientMsgId)
}
