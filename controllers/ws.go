package controllers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"imserver/config"
	"imserver/models"
	"imserver/util"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	newline = []byte{'\n'}
	space   = []byte{' '}
	lock    sync.Mutex
)

// 主WS handle
func (cl *Controller) WS(c *gin.Context) {

	//这是一个新的链接，以下操作可能有N个线程同时进行操作
	//升级websocket前验证http头部
	var (
		isChecked = true
		//isLang    = true
		auth = c.Query("token")
	)

	if auth == "" {
		auth = c.Copy().GetHeader("Authorization")
	}

	user, err := util.CheckAuthHeader(auth)
	if err != nil {
		isChecked = false
	}

	// lang := c.Copy().GetHeader("TalentPool-Language")
	// if lang == "" {
	// 	isLang = false
	// }

	//这里的conn是内存指针类型，可以各处引用
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	//确认头部解析完成
	if !isChecked {
		conn.WriteJSON(gin.H{"action": "SystemMsg", "code": 10000, "msg": "AUTH_CHECK_ERROR"})
		conn.Close()
		return
	}
	//改为前端自己判断
	//确认多语言不是为空
	// if !isLang {
	// 	conn.WriteJSON(gin.H{"action": "ErrorMsg", "code": 10001, "msg": "LANGUAGE_CHECK_ERROR"})
	// 	conn.Close()
	// 	return
	// }
	//判断该用户的验证token是否已经超时了
	//if util.CheckAuthHeaderIsExpired(user.Expired) {
	//	conn.WriteJSON(gin.H{"action": "ErrorMsg", "code": 10002, "msg": "USER_TOKEN_EXPIRED"})
	//	conn.Close()
	//	return
	//}
	//获取唯一链接标识
	connTag := util.GetAuthToken()

	//初始化用户如果不存在内存中
	//这里需要快速判断不能有多的读写任务否则会影响后续的线程

	//内存用户结构
	memUser := &models.InitUser{
		ConnTag:  connTag,
		Conn:     conn,
		UUID:     user.Uuid,
		Usertype: user.UserType,
		Send:     make(chan []byte, 2048),
		Model:    cl.M,
	}

	lock.Lock()
	isInit := cl.M.CheckIsInitUser(user.Uuid)
	if isInit {
		//初始化用户1正在初始化2初始化完毕nil还没有初始化
		cl.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Set(user.Uuid, 1)
	}
	lock.Unlock()
	//是否是初始化用户 true 程序初始化 false有已经初始化的值了
	if isInit {
		//提示前端正在初始化用户
		err = conn.WriteJSON(gin.H{"action": "SystemMsg", "code": 20000, "msg": "INITIALIZING_USER"})
		if err != nil {
			cl.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Remove(user.Uuid)
			conn.Close()
		}

		//获取当前用户的Profile
		userinfo, err := cl.M.GetUserProfileByUUID(user)
		if err != nil {
			cl.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Remove(user.Uuid)
			err = conn.WriteJSON(gin.H{"action": "SystemMsg", "code": 10003, "msg": err.Error()})
			if err != nil {
				conn.Close()
			}
		}
		//初始化内存用户
		cl.M.InitUser(memUser, &userinfo)

		//用户的profile到全局
		cl.M.MemDB.Get("Profile").(*gmap.AnyAnyMap).Set(user.Uuid, userinfo)

		//读取用户的联系人
		friendList, err := cl.M.GetUserContactsByUUID(user.Uuid)
		if err != nil {
			cl.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Remove(user.Uuid)
			err = conn.WriteJSON(gin.H{"action": "SystemMsg", "code": 10004, "msg": "GET_FRIEND_LIST_ERROR"})
			if err != nil {
				conn.Close()
			}
		}
		//设置联系人和黑名单
		if len(friendList) > 0 {
			for _, v := range friendList {
				if v.Isblack == 0 {
					cl.M.MemDB.Get(user.Uuid).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(v.Fromuser, models.RecvStatus{
						RecvCount:    v.Count,
						NextRecvTime: v.Nexttime,
					})
				}
				if v.Isblack == 1 {
					cl.M.MemDB.Get(user.Uuid).(*gmap.AnyAnyMap).Get("Blacklist").(*gmap.AnyAnyMap).Set(v.Fromuser, models.RecvStatus{
						RecvCount:    v.Count,
						NextRecvTime: v.Nexttime,
					})
				}
			}
		}

		//用户初始化完毕
		cl.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Set(user.Uuid, 2)

	}

	//检查其他线程中是否唤起这个初始化用户
	for {
		isInitedUser := cl.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Get(user.Uuid)
		if isInitedUser == nil {
			//其他线程初始化这个用户发生一个致命错误
			err = conn.WriteJSON(gin.H{"action": "SystemMsg", "code": 10006, "msg": "INIT_USER_ERROR"})
			if err != nil {
				conn.Close()
			}
			break
		}
		if isInitedUser == 2 {
			//初始化完成
			break
		}
	}

	//设置连接 如果存在这个conntoken 说明唯一的链接ID出现了一个重复的链接ID，遇到亿万分之一的可能性
	if false == cl.M.MemDB.Get(user.Uuid).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).SetIfNotExist(connTag, memUser.Send) {
		err = conn.WriteJSON(gin.H{"action": "SystemMsg", "code": 10005, "msg": "REPEATED_CONN_TAG"})
		if err != nil {
			conn.Close()
		}
	}

	err = conn.WriteJSON(gin.H{"action": "SystemMsg", "code": 20001, "msg": "IS_READY", "uuid": user.Uuid})
	if err != nil {
		conn.Close()
	}

	go cl.WsReadMsgs(memUser)
	go cl.WsWriteMsg(memUser)

}

func (cl *Controller) MockWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	var user models.UserToken
	user.Uuid = "xx123456789"
	connTag := util.GetAuthToken()
	memUser := &models.InitUser{
		ConnTag:  connTag,
		Conn:     conn,
		UUID:     user.Uuid,
		Usertype: user.UserType,
		Send:     make(chan []byte, 2048),
		Model:    cl.M,
	}

	lock.Lock()
	isInit := cl.M.CheckIsInitUser(user.Uuid)
	if isInit {
		//初始化用户1正在初始化2初始化完毕nil还没有初始化
		cl.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Set(user.Uuid, 1)
	}
	lock.Unlock()
	if isInit {
		cl.M.InitUser(memUser, nil)
		//用户初始化完毕
		cl.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Set(user.Uuid, 2)
	}
	//检查其他线程中是否唤起这个初始化用户
	for {
		isInitedUser := cl.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Get(user.Uuid)
		if isInitedUser == nil {
			//其他线程初始化这个用户发生一个致命错误
			err = conn.WriteJSON(gin.H{"action": "SystemMsg", "code": 10006, "msg": "INIT_USER_ERROR"})
			if err != nil {
				conn.Close()
			}
			break
		}
		if isInitedUser == 2 {
			//初始化完成
			break
		}
	}
	cl.M.MemDB.Get(user.Uuid).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).SetIfNotExist(connTag, memUser.Send)
	go cl.WsReadMsgs(memUser)
	go cl.WsWriteMsg(memUser)

}
func (c *Controller) WsWriteMsg(s *models.InitUser) {
	//收到要写入到conn的消息
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-s.Send:
			s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				s.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := s.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(s.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-s.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Controller) WsReadMsgs(s *models.InitUser) {
	defer func() {
		close(s.Send)
		c.WsUnregister(s)
		s.Conn.Close()

	}()
	s.Conn.SetReadLimit(maxMessageSize)
	s.Conn.SetReadDeadline(time.Now().Add(pongWait))
	s.Conn.SetPongHandler(func(string) error { s.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := s.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		var wsMsg WsMsg
		err = json.Unmarshal(message, &wsMsg)
		if err != nil {
			if c.Debug {
				s.Conn.WriteJSON(gin.H{"action": "SystemMsg", "code": 10007, "msg": "ERROR_MSG_FORMAT"})
			} else {
				fmt.Println("错误的消息格式", err)
				break
			}
		}

		c.M.NoSqlDB.SetIfNotExistWithTTL(fmt.Sprintf("%d", wsMsg.MsgId), s.UUID, 5*time.Minute)

		fmt.Println(wsMsg)
		switch wsMsg.Action {
		case "SendP2PMsg":
			go c.SendP2PMsg(s, wsMsg)
		case "PutRecvMsgId":
			go c.PutRecvMsgId(s, wsMsg)
		case "GetP2PMsgsNew":
			go c.GetP2PMsgs(s, wsMsg)
		case "GetP2PMsgsOld":
			go c.GetP2PMsgs(s, wsMsg)
		case "GetP2PMsgsRecent":
			go c.GetP2PMsgs(s, wsMsg)
		case "PutBlack":
			go c.PutBlack(s, wsMsg)
		case "DelBlack":
			go c.DelBlack(s, wsMsg)
		case "GetMyProfile":
			go c.GetUserProfile(s, wsMsg)
		case "GetMyContacts":
			go c.GetMyContacts(s, wsMsg)
		}

		//c.SendUsersMsg(s.UUID, message)
		fmt.Println(s.Model.MemDB.Get(s.UUID))
	}

}

//handle SendP2PMsg

func (c *Controller) SendP2PMsg(s *models.InitUser, wsMsg WsMsg) {
	var errorMsg ErrorMsg
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main", r)
		}
	}()

	//移除消息里面的xss攻击
	wsMsg.Msg = util.XSSRemover(wsMsg.Msg)
	//判断接收方是不是在内存里面
	lock.Lock()
	isInit := c.M.CheckIsInitUser(wsMsg.ToUser)
	if isInit {
		//初始化用户1正在初始化2初始化完毕nil还没有初始化
		c.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Set(wsMsg.ToUser, 1)
	}
	lock.Unlock()
	if isInit {
		//这里开始唤起初始化其他用户
		//获取当前用户的Profile
		//获取要获取的人的类型：个人|企业
		var userType int = 0
		if s.Usertype == 0 {
			userType = 1
		}

		userinfo, err := c.M.GetUserProfileByUUID(models.UserToken{
			UserType: userType,
			Uuid:     wsMsg.ToUser,
		})
		if err != nil {
			c.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Remove(wsMsg.ToUser)
			s.Send <- []byte(err.Error())
			return
		}

		memUser := &models.InitUser{
			UUID: wsMsg.ToUser,
		}

		c.M.InitUser(memUser, &userinfo)

		//用户的profile到全局
		c.M.MemDB.Get("Profile").(*gmap.AnyAnyMap).Set(wsMsg.ToUser, userinfo)

		//读取用户的联系人
		//friendList, err := c.M.GetUserContactsByUUID(wsMsg.ToUser)
		friendList, err := config.ImFriendRepo.ListImFriendByUuid(wsMsg.ToUser)
		if err != nil {
			c.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Remove(wsMsg.ToUser)
			s.Send <- []byte(err.Error())
			return
		}
		//设置联系人和黑名单
		if len(friendList) > 0 {
			for _, v := range friendList {
				if !v.IsBlack {
					c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(v.UserUuid, models.RecvStatus{
						RecvCount:    v.Count,
						NextRecvTime: v.NextTime,
					})
				}
				if v.IsBlack {
					c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Blacklist").(*gmap.AnyAnyMap).Set(v.UserUuid, models.RecvStatus{
						RecvCount:    v.Count,
						NextRecvTime: v.NextTime,
					})
				}
			}
		}

		//用户初始化完毕
		c.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Set(wsMsg.ToUser, 2)
	}

	//检查其他线程中是否唤起这个初始化用户
	for {
		isInitedUser := c.M.MemDB.Get("IsInitUser").(*gmap.AnyAnyMap).Get(wsMsg.ToUser)
		if isInitedUser == nil {
			//其他线程初始化这个用户发生一个致命错误
			errorMsg = ErrorMsg{
				Action: "ErrorMsg",
				Code:   10008,
				MsgId:  wsMsg.MsgId,
				Msg:    "INIT_REMOTE_USER_ERROR",
			}
			message, _ := json.Marshal(&errorMsg)
			s.Send <- message
			return
		}
		if isInitedUser == 2 {
			//初始化完成
			break
		}
	}

	//判断这两个人是不是有过联系
	lock.Lock()
	isNew, err := c.CheckOrSetFriends(s.UUID, wsMsg.ToUser)
	if err != nil {
		s.Send <- []byte(err.Error())
		lock.Unlock()
		return
	}
	lock.Unlock()
	//标记是否可以发送消息
	var canSendMsg bool = false
	var everContacted bool
	// fixme 缓存
	targetFriend, _ := config.ImFriendRepo.GetFriendByUuid(wsMsg.ToUser, s.UUID)
	if targetFriend != nil {
		everContacted = targetFriend.EverContacted
		canSendMsg = targetFriend.EverContacted
	}

	//新的朋友关系不用判断黑名单和可以发送消息的数量
	if !isNew {
		//检查我是不是在它的黑名单里面
		isBlack := c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Blacklist").(*gmap.AnyAnyMap).Get(s.UUID)
		if isBlack != nil {
			errorMsg = ErrorMsg{
				Action: "ErrorMsg",
				Code:   10009,
				MsgId:  wsMsg.MsgId,
				Msg:    "IN_BLACK_LIST",
			}
			message, _ := json.Marshal(&errorMsg)
			s.Send <- message
			return
		}
	}
	sendMsg := &RecvP2PMsg{
		Action:       "RecvP2PMsg",
		MsgId:        wsMsg.MsgId,
		FromUser:     s.UUID,
		FromUserType: s.Usertype,
		ToUser:       wsMsg.ToUser,
		Msg:          wsMsg.Msg,
		//ReadId:       uint64(readId),
	}

	//检查我还能不能发消息给他
	lock.Lock()
	canSendCount := c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Get(s.UUID)
	if canSendCount != nil {
		count := canSendCount.(models.RecvStatus).RecvCount
		nextSendTime := canSendCount.(models.RecvStatus).NextRecvTime

		recvCount := 0
		var nextRecvTime int64 = 0
		if count == 0 && nextSendTime <= time.Now().Unix() {
			//这个是可以发送的
			//这次发送算一次 并重置时间
			recvCount = 1
			nextRecvTime = 0

			c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(s.UUID, models.RecvStatus{
				RecvCount:    1,
				NextRecvTime: 0,
			})
			canSendMsg = true
		}

		if count > 1 {
			recvCount = count - 1
			nextRecvTime = 0
			c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(s.UUID, models.RecvStatus{
				RecvCount:    recvCount,
				NextRecvTime: 0,
			})
			canSendMsg = true
		}

		if count == 1 {
			//还剩下最后一次发送机会
			recvCount = 0
			nextRecvTime = time.Now().Add(config.CAN_SEND_NEXT_TIME).Unix()
			c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(s.UUID, models.RecvStatus{
				RecvCount:    0,
				NextRecvTime: nextRecvTime,
			})
			canSendMsg = true
		}

		if canSendMsg || everContacted {
			//重置它可以给我发消息的次数
			c.M.MemDB.Get(s.UUID).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(wsMsg.ToUser, models.RecvStatus{
				RecvCount:    config.CAN_SEND_COUNT,
				NextRecvTime: 0,
			})

			//更新数据库
			err = c.M.SetRecvCountByTouser(wsMsg.ToUser, s.UUID, recvCount, nextRecvTime)
			if err != nil {
				//聊天发送次数回退
				c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(s.UUID, models.RecvStatus{
					RecvCount:    count + 1,
					NextRecvTime: 0,
				})
				errorMsg = ErrorMsg{
					Action: "ErrorMsg",
					Code:   10020,
					MsgId:  wsMsg.MsgId,
					Msg:    "MANTICORE_UPDATE_TOUSER_RECV_COUNT_ERROR",
				}
				message, _ := json.Marshal(&errorMsg)
				s.Send <- message
				lock.Unlock()
				return
			}

			err = c.M.SetRecvCountByTouser(s.UUID, wsMsg.ToUser, 2, 0)
			if err != nil {
				//聊天发送次数回退
				c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(s.UUID, models.RecvStatus{
					RecvCount:    count + 1,
					NextRecvTime: 0,
				})
				errorMsg = ErrorMsg{
					Action: "ErrorMsg",
					Code:   10021,
					MsgId:  wsMsg.MsgId,
					Msg:    "MANTICORE_UPDATE_UUID_RECV_COUNT_ERROR",
				}
				message, _ := json.Marshal(&errorMsg)
				s.Send <- message
				lock.Unlock()
				return
			}

			//插入聊天记录
			readId, err := c.M.InsertMessages(s, wsMsg.ToUser, wsMsg.Msg, wsMsg.MsgId)
			if err != nil {
				//聊天发送次数回退
				c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(s.UUID, models.RecvStatus{
					RecvCount:    count + 1,
					NextRecvTime: 0,
				})
				errorMsg = ErrorMsg{
					Action: "ErrorMsg",
					Code:   10010,
					MsgId:  wsMsg.MsgId,
					Msg:    "SERVER_INSERT_CHAT_MSG_ERROR",
				}
				message, _ := json.Marshal(&errorMsg)
				s.Send <- message
				lock.Unlock()
				return
			}
			sendMsg.ReadId = uint64(readId)

			// 设置曾经联系过
			// fixme 缓存
			config.ImFriendRepo.UpdateContactStatus(s.UUID, wsMsg.ToUser)
		}

	}
	lock.Unlock()

	if canSendMsg {
		message, _ := json.Marshal(sendMsg)
		//发一遍给自己的所有连接
		c.SendUsersMsg(s.UUID, message)
		//再发一遍给接收人的所有连接
		c.SendUsersMsg(wsMsg.ToUser, message)
	} else {
		errorMsg = ErrorMsg{
			Action: "ErrorMsg",
			Code:   10030,
			MsgId:  wsMsg.MsgId,
			Msg:    "EXCEED_TOUSER_RECV_COUNT_LIMIT",
		}
		message, _ := json.Marshal(&errorMsg)
		c.SendUsersMsg(s.UUID, message)
	}

}

//

func (c *Controller) PutRecvMsgId(s *models.InitUser, wsMsg WsMsg) {
	var errorMsg ErrorMsg
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main", r)
		}
	}()
	//每个线程因为同步问题不可能发送的都是最新的last_read_id
	//所以这里必须继续锁住然后判断
	//获取当前 它<-我 的消息的最后阅读ID
	lock.Lock()
	relatedKey := fmt.Sprintf(`%s<-%s`, wsMsg.ToUser, s.UUID)
	val, err := c.M.NoSqlDB.Get(relatedKey)
	if err != nil {
		errorMsg = ErrorMsg{
			Action: "ErrorMsg",
			Code:   10011,
			MsgId:  wsMsg.MsgId,
			Msg:    "GET_LAST_READ_ID_ERROR",
		}
		message, _ := json.Marshal(&errorMsg)
		s.Send <- message
		lock.Unlock()
		return
	}
	storedReadId, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		errorMsg = ErrorMsg{
			Action: "ErrorMsg",
			Code:   10012,
			MsgId:  wsMsg.MsgId,
			Msg:    "CONVERT_STORED_READ_ID_ERROR",
		}
		message, _ := json.Marshal(&errorMsg)
		s.Send <- message
		lock.Unlock()
		return
	}

	thisReadId, err := strconv.ParseUint(wsMsg.Msg, 10, 64)
	if err != nil {
		errorMsg = ErrorMsg{
			Action: "ErrorMsg",
			Code:   10013,
			MsgId:  wsMsg.MsgId,
			Msg:    "CONVERT_PUT_READ_ID_ERROR",
		}
		message, _ := json.Marshal(&errorMsg)
		s.Send <- message
		lock.Unlock()
		return
	}
	lock.Unlock()
	//比较存储的和需要更新的ID
	if thisReadId > storedReadId {
		c.M.NoSqlDB.Set(relatedKey, fmt.Sprintf("%d", thisReadId))
		recvMsg := &WsMsg{
			Action:   "RecvMsgId",
			ToUser:   wsMsg.ToUser,
			FromUser: s.UUID,
			MsgId:    wsMsg.MsgId,
			Msg:      wsMsg.Msg,
		}
		message, _ := json.Marshal(recvMsg)
		s.Send <- message
	} else {
		recvMsg := &WsMsg{
			Action:   "RecvMsgId",
			ToUser:   wsMsg.ToUser,
			FromUser: s.UUID,
			MsgId:    wsMsg.MsgId,
			Msg:      val,
		}
		message, _ := json.Marshal(recvMsg)
		s.Send <- message
	}
}

// 获取聊天历史消息
func (c *Controller) GetP2PMsgs(s *models.InitUser, wsMsg WsMsg) {
	var (
		errorMsg    ErrorMsg
		recvP2PMsgs RecvP2PMsgs
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main", r)
		}
	}()
	if wsMsg.Action == "GetP2PMsgsNew" {
		startReadId, err := strconv.ParseUint(wsMsg.MsgId, 10, 64)
		if err != nil {
			errorMsg = ErrorMsg{
				Action: "ErrorMsg",
				Code:   10014,
				MsgId:  wsMsg.MsgId,
				Msg:    "CONVERT_P2PMSGS_NEW_READ_ID_ERROR",
			}
			message, _ := json.Marshal(&errorMsg)
			s.Send <- message
			return
		}
		data, err := c.M.GetMessagesByStartId(s.UUID, wsMsg.ToUser, "new", startReadId)
		if err != nil {
			errorMsg = ErrorMsg{
				Action: "ErrorMsg",
				Code:   10015,
				MsgId:  wsMsg.MsgId,
				Msg:    "GET_P2PMSGS_NEW_ERROR",
			}
			message, _ := json.Marshal(&errorMsg)
			s.Send <- message
			return
		}
		recvP2PMsgs = RecvP2PMsgs{
			Action:   "RecvP2PMsgsNew",
			FromUser: s.UUID,
			ToUser:   wsMsg.ToUser,
			MsgId:    wsMsg.MsgId,
			Msg:      data,
		}

	}
	if wsMsg.Action == "GetP2PMsgsOld" {
		startReadId, err := strconv.ParseUint(wsMsg.MsgId, 10, 64)
		if err != nil {
			errorMsg = ErrorMsg{
				Action: "ErrorMsg",
				Code:   10016,
				MsgId:  wsMsg.MsgId,
				Msg:    "CONVERT_P2PMSGS_OLD_READ_ID_ERROR",
			}
			message, _ := json.Marshal(&errorMsg)
			s.Send <- message
			return
		}
		data, err := c.M.GetMessagesByStartId(s.UUID, wsMsg.ToUser, "old", startReadId)
		if err != nil {
			errorMsg = ErrorMsg{
				Action: "ErrorMsg",
				Code:   10017,
				MsgId:  wsMsg.MsgId,
				Msg:    "GET_P2PMSGS_OLD_ERROR",
			}
			message, _ := json.Marshal(&errorMsg)
			s.Send <- message
			return
		}
		recvP2PMsgs = RecvP2PMsgs{
			Action:   "RecvP2PMsgsOld",
			FromUser: s.UUID,
			ToUser:   wsMsg.ToUser,
			MsgId:    wsMsg.MsgId,
			Msg:      data,
		}

	}
	b := wsMsg.Action == "GetP2PMsgsRecent"
	fmt.Println(b)
	if wsMsg.Action == "GetP2PMsgsRecent" {
		data, err := c.M.GetMessagesByStartId(s.UUID, wsMsg.ToUser, "recent", 0)
		if err != nil {
			errorMsg = ErrorMsg{
				Action: "ErrorMsg",
				Code:   10018,
				MsgId:  wsMsg.MsgId,
				Msg:    "GET_P2PMSGS_RECENT_ERROR",
			}
			message, _ := json.Marshal(&errorMsg)
			s.Send <- message
			return
		}
		recvP2PMsgs = RecvP2PMsgs{
			Action:   "RecvP2PMsgsRecent",
			FromUser: s.UUID,
			ToUser:   wsMsg.ToUser,
			MsgId:    wsMsg.MsgId,
			Msg:      data,
		}
	}

	if recvP2PMsgs.Action != "" {
		message, _ := json.Marshal(&recvP2PMsgs)
		s.Send <- message
		return
	}
}

func (c *Controller) PutBlack(s *models.InitUser, wsMsg WsMsg) {
	var errorMsg ErrorMsg
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main", r)
		}
	}()

	lock.Lock()

	//内存移动
	thisUser := c.M.MemDB.Get(s.UUID).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Get(wsMsg.ToUser)
	if thisUser == nil {
		errorMsg = ErrorMsg{
			Action: "ErrorMsg",
			Code:   10019,
			MsgId:  wsMsg.MsgId,
			Msg:    "CAN_NOT_FIND_USER_IN_CONTACTS",
		}
		message, _ := json.Marshal(&errorMsg)
		s.Send <- message
		lock.Unlock()
		return
	}

	//数据库更新
	err := c.M.SetBlacklistByTouser(s.UUID, wsMsg.ToUser, 1)
	if err != nil {
		errorMsg = ErrorMsg{
			Action: "ErrorMsg",
			Code:   10022,
			MsgId:  wsMsg.MsgId,
			Msg:    "MANTICORE_SET_UUID_BLACKLIST_ERROR",
		}
		message, _ := json.Marshal(&errorMsg)
		s.Send <- message
		lock.Unlock()
		return
	}

	//数据库先设置不出错再设置内存
	c.M.MemDB.Get(s.UUID).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Remove(wsMsg.ToUser)
	c.M.MemDB.Get(s.UUID).(*gmap.AnyAnyMap).Get("Blacklist").(*gmap.AnyAnyMap).Set(wsMsg.ToUser, models.RecvStatus{
		RecvCount:    thisUser.(models.RecvStatus).RecvCount,
		NextRecvTime: thisUser.(models.RecvStatus).NextRecvTime,
	})

	lock.Unlock()
	recvMsg := &WsMsg{
		Action:   "RecvBlack",
		ToUser:   wsMsg.ToUser,
		FromUser: s.UUID,
		MsgId:    wsMsg.MsgId,
		Msg:      wsMsg.Msg,
	}
	message, _ := json.Marshal(&recvMsg)
	c.SendUsersMsg(s.UUID, message)

}

func (c *Controller) DelBlack(s *models.InitUser, wsMsg WsMsg) {
	var errorMsg ErrorMsg
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main", r)
		}
	}()
	lock.Lock()

	//内存移动
	thisUser := c.M.MemDB.Get(s.UUID).(*gmap.AnyAnyMap).Get("Blacklist").(*gmap.AnyAnyMap).Get(wsMsg.ToUser)
	if thisUser == nil {
		errorMsg = ErrorMsg{
			Action: "ErrorMsg",
			Code:   10024,
			MsgId:  wsMsg.MsgId,
			Msg:    "CAN_NOT_FIND_USER_IN_BLACKLIST",
		}
		message, _ := json.Marshal(&errorMsg)
		s.Send <- message
		lock.Unlock()
		return
	}

	//数据库更新
	err := c.M.SetBlacklistByTouser(s.UUID, wsMsg.ToUser, 0)
	if err != nil {
		errorMsg = ErrorMsg{
			Action: "ErrorMsg",
			Code:   10023,
			MsgId:  wsMsg.MsgId,
			Msg:    "MANTICORE_DEL_UUID_BLACKLIST_ERROR",
		}
		message, _ := json.Marshal(&errorMsg)
		s.Send <- message
		lock.Unlock()
		return
	}

	//数据库先设置不出错再设置内存
	c.M.MemDB.Get(s.UUID).(*gmap.AnyAnyMap).Get("Blacklist").(*gmap.AnyAnyMap).Remove(wsMsg.ToUser)
	c.M.MemDB.Get(s.UUID).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(wsMsg.ToUser, models.RecvStatus{
		RecvCount:    thisUser.(models.RecvStatus).RecvCount,
		NextRecvTime: thisUser.(models.RecvStatus).NextRecvTime,
	})

	lock.Unlock()
	recvMsg := &WsMsg{
		Action:   "RecvDelBlack",
		ToUser:   wsMsg.ToUser,
		FromUser: s.UUID,
		MsgId:    wsMsg.MsgId,
		Msg:      wsMsg.Msg,
	}
	message, _ := json.Marshal(&recvMsg)
	c.SendUsersMsg(s.UUID, message)
}

// 获取我或其他联系人的基本信息
func (c *Controller) GetUserProfile(s *models.InitUser, wsMsg WsMsg) {
	var errorMsg ErrorMsg
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main", r)
		}
	}()
	//获取我的
	if wsMsg.ToUser == s.UUID {
		thisUser := c.M.MemDB.Get(s.UUID).(*gmap.AnyAnyMap).Get("Profile")
		if thisUser != nil {
			message, _ := json.Marshal(&UserProfile{
				Action:   "RecvMyProfile",
				ToUser:   wsMsg.ToUser,
				FromUser: s.UUID,
				MsgId:    wsMsg.MsgId,
				Msg:      thisUser.(models.UserBasicInfo),
			})
			s.Send <- message
			return
		} else {
			//照道理不会出现这种情况
			user, err := c.M.GetUserProfileByUUID(models.UserToken{
				Uuid:     s.UUID,
				UserType: s.Usertype,
			})
			if err != nil {
				errorMsg = ErrorMsg{
					Action: "ErrorMsg",
					Code:   10025,
					MsgId:  wsMsg.MsgId,
					Msg:    "GET_MY_BASIC_INFO_ERROR",
				}
				message, _ := json.Marshal(&errorMsg)
				s.Send <- message
				return
			}
			c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Profile").(*gmap.AnyAnyMap).Set(s.UUID, user)
			message, _ := json.Marshal(&UserProfile{
				Action:   "RecvMyProfile",
				ToUser:   wsMsg.ToUser,
				FromUser: s.UUID,
				MsgId:    wsMsg.MsgId,
				Msg:      user,
			})
			s.Send <- message
			return
		}
	} else {
		thisUser := c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Profile")
		if thisUser != nil {
			message, _ := json.Marshal(&UserProfile{
				Action:   "RecvUserProfile",
				ToUser:   wsMsg.ToUser,
				FromUser: s.UUID,
				MsgId:    wsMsg.MsgId,
				Msg:      thisUser.(models.UserBasicInfo),
			})
			s.Send <- message
			return
		} else {
			userType := 0
			if s.Usertype == 0 {
				userType = 1
			}
			user, err := c.M.GetUserProfileByUUID(models.UserToken{
				Uuid:     wsMsg.ToUser,
				UserType: userType,
			})
			if err != nil {
				errorMsg = ErrorMsg{
					Action: "ErrorMsg",
					Code:   10026,
					MsgId:  wsMsg.MsgId,
					Msg:    "GET_USER_BASIC_INFO_ERROR",
				}
				message, _ := json.Marshal(&errorMsg)
				s.Send <- message
				return
			}
			c.M.MemDB.Get(wsMsg.ToUser).(*gmap.AnyAnyMap).Get("Profile").(*gmap.AnyAnyMap).Set(wsMsg.ToUser, user)
			message, _ := json.Marshal(&UserProfile{
				Action:   "RecvUserProfile",
				ToUser:   wsMsg.ToUser,
				FromUser: s.UUID,
				MsgId:    wsMsg.MsgId,
				Msg:      user,
			})
			s.Send <- message
			return
		}
	}
}

// 获取联系人名单(含黑名单和头像)
func (c *Controller) GetMyContacts(s *models.InitUser, wsMsg WsMsg) {
	var (
		errorMsg   ErrorMsg
		myContacts []models.MyContacts
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main", r, string(debug.Stack()))
		}
	}()
	if wsMsg.ToUser == s.UUID {
		friendList, err := c.M.GetUserContactsByUUID(s.UUID)
		if err != nil {
			errorMsg = ErrorMsg{
				Action: "ErrorMsg",
				Code:   10027,
				MsgId:  wsMsg.MsgId,
				Msg:    "GET_MY_CONTACTS_INFO_ERROR",
			}
			message, _ := json.Marshal(&errorMsg)
			s.Send <- message
			return
		}
		if len(friendList) > 0 {
			for _, v := range friendList {
				//我的联系人是fromuser touser是我自己
				if c.M.MemDB.Get(v.Fromuser) == nil {
					continue
				}
				thisUser := c.M.MemDB.Get(v.Fromuser).(*gmap.AnyAnyMap).Get("Profile")
				if thisUser != nil {
					myContacts = append(myContacts, models.MyContacts{
						Uuid:       thisUser.(models.UserBasicInfo).Uuid,
						FullNameEn: thisUser.(models.UserBasicInfo).FullNameEn,
						FullNameJa: thisUser.(models.UserBasicInfo).FullNameJa,
						Avatar:     thisUser.(models.UserBasicInfo).Avatar,
						IsBlack:    v.Isblack,
					})

				} else {
					userType := 0
					if s.Usertype == 0 {
						userType = 1
					}
					user, err := c.M.GetUserProfileByUUID(models.UserToken{
						Uuid:     wsMsg.ToUser,
						UserType: userType,
					})
					if err != nil {
						errorMsg = ErrorMsg{
							Action: "ErrorMsg",
							Code:   10028,
							MsgId:  wsMsg.MsgId,
							Msg:    "GET_MY_CONTACTS_USER_BASIC_INFO_ERROR",
						}
						message, _ := json.Marshal(&errorMsg)
						s.Send <- message
						return
					}
					c.M.MemDB.Get(v.Fromuser).(*gmap.AnyAnyMap).Get("Profile").(*gmap.AnyAnyMap).Set(wsMsg.ToUser, user)
					myContacts = append(myContacts, models.MyContacts{
						Uuid:       user.Uuid,
						FullNameEn: user.FullNameEn,
						FullNameJa: user.FullNameJa,
						Avatar:     user.Avatar,
						IsBlack:    v.Isblack,
					})
				}

			}

			message, _ := json.Marshal(&MyContacts{
				Action:   "RecvMyContacts",
				ToUser:   wsMsg.ToUser,
				FromUser: s.UUID,
				MsgId:    wsMsg.MsgId,
				Msg:      myContacts,
			})
			s.Send <- message
			return
		}

	}

}

// 设置联系人关系
func (c *Controller) CheckOrSetFriends(uuid, touser string) (isNew bool, err error) {
	isNew = true
	hash := md5.Sum([]byte(uuid + touser))
	//我和它的唯一关系ID
	friendkey := hex.EncodeToString(hash[:])

	batch := c.M.NoSqlDB.GetBatch()
	//判断是不是有联系人的key
	val, err := batch.Get([]byte(friendkey))
	if err != nil {
		if err.Error() != "key not found in database" {
			batch.Rollback()
			return
		}
	}
	//如果存在我和它的联系人key
	if val != nil {
		isNew = false
		batch.Rollback()
		return
	}
	//它和我的联系人关系
	hash = md5.Sum([]byte(touser + uuid))
	//我和它的唯一关系ID
	friendkey2 := hex.EncodeToString(hash[:])

	if friendkey == friendkey2 {
		err = fmt.Errorf("不能和自己为联系人")
		return
	}
	//照道理这段代码是不应该发生的
	val, err = batch.Get([]byte(friendkey2))
	if err != nil {
		if err.Error() != "key not found in database" {
			batch.Rollback()
			return
		}
	}
	if val != nil {
		isNew = false
		batch.Rollback()
		return
	}
	//
	//设置我和它的联系人互相关系
	err = batch.Put([]byte(friendkey), []byte(""))
	if err != nil {
		batch.Rollback()
		return
	}
	err = batch.Put([]byte(friendkey2), []byte(""))
	if err != nil {
		batch.Rollback()
		return
	}

	//设置互相的最后阅读ID
	relatedKey := fmt.Sprintf(`%s<-%s`, touser, uuid)
	err = batch.Put([]byte(relatedKey), []byte("0"))
	if err != nil {
		batch.Rollback()
		return
	}
	relatedKey2 := fmt.Sprintf(`%s<-%s`, uuid, touser)
	err = batch.Put([]byte(relatedKey2), []byte("0"))
	if err != nil {
		batch.Rollback()
		return
	}

	//设置关系到数据库
	session := c.M.ManticoreDB.NewSession()
	err = session.Begin()
	if err != nil {
		return
	}
	_, err = session.Exec("BEGIN")
	if err != nil {
		return
	}
	existFriend, err := config.ImFriendRepo.GetFriendByUuid(uuid, touser)
	if existFriend == nil {
		sql := fmt.Sprintf(`insert into im_friend_list (fromuser,touser,isblack,count,status,created,nexttime) values ('%s','%s',%d,%d,%d,%d,%d)`, uuid, touser, 0, config.CAN_SEND_COUNT, 1, time.Now().UnixMilli(), 0)
		_, err = session.Exec(sql)
		if err != nil {
			batch.Rollback()
			session.Rollback()
			return
		}
	}
	anotherExistFriend, err := config.ImFriendRepo.GetFriendByUuid(touser, uuid)
	if anotherExistFriend == nil {
		sql := fmt.Sprintf(`insert into im_friend_list (fromuser,touser,isblack,count,status,created,nexttime) values ('%s','%s',%d,%d,%d,%d,%d)`, touser, uuid, 0, config.CAN_SEND_COUNT, 1, time.Now().UnixMilli(), 0)
		_, err = session.Exec(sql)
		if err != nil {
			batch.Rollback()
			session.Rollback()
			return
		}
	}
	//一致性提交
	err = batch.Commit()
	if err != nil {
		session.Rollback()
		return
	}
	//session.Rollback()
	err = session.Commit()
	if err != nil {
		batch.Rollback()
		return
	}
	//更新我的好友里面有它 它的好友里面有我

	c.M.MemDB.Get(uuid).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(touser, models.RecvStatus{
		RecvCount:    2,
		NextRecvTime: 0,
	})
	c.M.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("Contacts").(*gmap.AnyAnyMap).Set(uuid, models.RecvStatus{
		RecvCount:    2,
		NextRecvTime: 0,
	})

	return
}

func (c *Controller) SendUsersMsg(uuid string, msg []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main", r)
		}
	}()
	connTags := c.M.MemDB.Get(uuid).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).Keys()
	for _, v := range connTags {
		if this := c.M.MemDB.Get(uuid).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).Get(v); this != nil {
			this.(chan []byte) <- msg
		}
	}
}

func (c *Controller) WsUnregister(s *models.InitUser) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main", r)
		}
	}()
	c.M.MemDB.Get(s.UUID).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).Remove(s.ConnTag)
}
