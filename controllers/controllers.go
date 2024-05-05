package controllers

import (
	"imserver/models"
)

type Controller struct {
	M     *models.Model
	Debug bool
}

type WsMsg struct {
	Action string `json:"action"`
	//消息ID
	MsgId string `json:"msgid"`
	//谁发送的消息
	FromUser string `json:"fromuser"`
	//发送给/对 谁进行处理
	ToUser string `json:"touser"`
	//
	Msg string `json:"msg"`
}

type RecvP2PMsg struct {
	Action string `json:"action"`
	//消息ID
	MsgId string `json:"msgid"`
	//服务器上聊天记录ID
	ReadId uint64 `json:"readid"`
	//谁发送的消息
	FromUser string `json:"fromuser"`
	//发送用户的类型
	FromUserType int `json:"fromusertype"`
	//发送给/对 谁进行处理
	ToUser string `json:"touser"`
	//
	Msg string `json:"msg"`
}

type ErrorMsg struct {
	Action string `json:"action"`
	//消息ID
	MsgId string `json:"msgid"`
	//错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type RecvP2PMsgs struct {
	Action string `json:"action"`
	//消息ID
	MsgId string `json:"msgid"`
	//谁发送的消息
	FromUser string `json:"fromuser"`
	//发送给/对 谁进行处理
	ToUser string `json:"touser"`
	//
	Msg []models.ImMessage `json:"msg"`
}

type UserProfile struct {
	Action string `json:"action"`
	//消息ID
	MsgId string `json:"msgid"`
	//谁发送的消息
	FromUser string `json:"fromuser"`
	//发送给/对 谁进行处理
	ToUser string `json:"touser"`
	//
	Msg models.UserBasicInfo `json:"msg"`
}

type MyContacts struct {
	Action string `json:"action"`
	//消息ID
	MsgId string `json:"msgid"`
	//谁发送的消息
	FromUser string `json:"fromuser"`
	//发送给/对 谁进行处理
	ToUser string `json:"touser"`
	//
	Msg []models.MyContacts `json:"msg"`
}

//发送普通消息
//{"action":"SendP2PMsg","msgid":"","fromuser":"sender_uuid(我)","touser":"recv_uuid","msg":"some msgs"}
//回复
//{"action":"RecvP2PMsg","readid":"","msgid":"","fromuser":"sender_uuid","touser":"recv_uuid","msg":"some msgs"}

//发送已读最后消息ID 它发给我的消息我已经读过的消息ID是last_read_msg_id 系统处理
//{"action":"PutRecvMsgId","msgid":"","fromuser":"我","touser":"它","msg":"last_read_msg_id"}
//回复
//{"action":"RecvMsgId","msgid":"","fromuser":"我","touser":"它","msg":"last_read_msg_id"}

//通过msg_id 获取新的或者旧的消息
//{"action":"GetP2PMsgsNew","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":"msg_id"}
//{"action":"GetP2PMsgsOld","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":"msg_id"}
//{"action":"GetP2PMsgsRecent","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":"msg_id"}
//回复
//{"action":"RecvP2PMsgsNew","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":[{ImMessage}]}
//{"action":"RecvP2PMsgsOld","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":[{ImMessage}]}
//{"action":"RecvP2PMsgsRecent","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":[{ImMessage}]}
//ImMessage = {"id":123,"touser":"","fromuser":"","msg":"","totype":1,"fromtype":0,"created":123,"msgid":""}

//添加到黑名单 我要把它放到黑名单
//{"action":"PutBlack","msgid":"","fromuser":"sender_uuid(我)","touser":"recv_uuid(它)","msg":"空"}
//{"action":"RecvBlack","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":"msg_id"}
//移除黑名单 我要把它移出黑名单
//{"action":"DelBlack","msgid":"","fromuser":"sender_uuid(我)","touser":"recv_uuid(它)","msg":"空"}
//{"action":"RecvDelBlack","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":"msg_id"}

//获取个人基本信息 我要获取它的基本信息
//获取我的个人基本信息
//{"action":"GetMyProfile","msgid":"","fromuser":"sender_uuid(我)","touser":"我","msg":"空"}
//{"action":"RecvMyProfile","msgid":"","fromuser":"sender_uuid(我)","touser":"我","msg":"空"}
//获取其他人的基本信息
//{"action":"GetUserProfile","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":UserProfile}
//{"action":"RecvUserProfile","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":UserProfile}
//UserProfile = {"uuid":"","full_name_en":"","full_name_ja","avatar":""}

//{"action":"GetMyContacts","msgid":"","fromuser":"sender_uuid(我)","touser":"我","msg":"空"}
//{"action":"RecvyContacts","msgid":"","fromuser":"sender_uuid(我)","touser":"我","msg":[{MyContacts}]}
//MyContacts={"uuid":"","full_name_en":"","full_name_ja":"","avatar":""}
