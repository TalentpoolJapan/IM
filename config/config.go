package config

import "time"

var (
	TOKEN_KEY = "df198138b5518471"
	//MYSQL_HOST    = "tcp(127.0.0.1:3306)"
	MYSQL_HOST    = "tcp(13.231.174.2:3306)"
	MYSQL_DB      = "talentpool"
	MYSQL_SECRECT = "yYVim5WbqzkWziNY"

	//可以在没有回复的情况下发送几条消息
	CAN_SEND_COUNT = 2
	//过滤多少小时以后还能再发消息的间隔
	CAN_SEND_NEXT_TIME = 3 * time.Hour
)
