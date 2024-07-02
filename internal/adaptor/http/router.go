package http

import "github.com/gin-gonic/gin"

func RegisterHandler(engine *gin.Engine) {
	engine.GET("/im/v1/my_contacts", GetMyContacts)

}
