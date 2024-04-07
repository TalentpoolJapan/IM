package controller

import (
	"bytes"
	"encoding/json"
	"imserver/config"
	"imserver/model"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/deatil/go-cryptobin/cryptobin/crypto"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gorilla/websocket"
	gonanoid "github.com/matoous/go-nanoid/v2"
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
	newline = []byte{'\n'}
	space   = []byte{' '}
	lock    sync.RWMutex
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *Controller) GetNewToken(ctx *gin.Context) {
	var (
		_userProfile UserToken
		_userToken   UserToken
	)

	err := ctx.ShouldBindJSON(&_userProfile)
	if err != nil {
		ctx.JSON(200, gin.H{"code": 10002, "msg": err.Error()})
		return
	}

	auth := ctx.GetHeader("Authorization")
	decrypt := crypto.FromBase64String(auth).SetKey(config.TOKEN_KEY).Aes().ECB().PKCS7Padding().Decrypt()
	decryptErr := decrypt.Error()
	if decryptErr != nil {
		ctx.JSON(200, gin.H{"code": 10003, "msg": decryptErr.Error()})
		return
	}
	cyptde := decrypt.ToBytes()
	err = json.Unmarshal(cyptde, &_userToken)
	if err != nil {
		ctx.JSON(200, gin.H{"code": 10004, "msg": err.Error()})
		return
	}

	data, _ := json.Marshal(_userToken)

	token := crypto.
		FromBytes(data).
		SetKey(config.TOKEN_KEY).
		Aes().
		ECB().
		PKCS7Padding().
		Encrypt().
		ToBase64String()
	ctx.JSON(200, gin.H{"code": 0, "msg": "ok", "data": token})
}

func (c *Controller) WsUnregister(s *model.MemInitUser) {
	c.M.WsUnregister(s)
	close(s.Send)
}

type WsMsg struct {
	Action     string          `json:"action"`
	Msg        model.ImMessage `json:"msg,omitempty"`
	Touser     string          `json:"touser,omitempty"`
	LastReadId int64           `json:"lastreadid,omitempty"`
	MsgId      string          `json:"msgid,omitempty"`
}

type WsErrMsg struct {
	Action string `json:"action"`
	Msg    ErrMsg `json:"msg"`
}
type ErrMsg struct {
	Msgid  string `json:"msgid,omitempty"`
	ErrMsg string `json:"errmsg,omitempty"`
}

func (c *Controller) WsReadMsg(s *model.MemInitUser) {
	defer func() {
		close(s.Send)
		c.WsUnregister(s)
		s.Conn.Close()
		lock.Unlock()
	}()
	s.Conn.SetReadLimit(maxMessageSize)
	s.Conn.SetReadDeadline(time.Now().Add(pongWait))
	s.Conn.SetPongHandler(func(string) error { s.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := s.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
				break
			}
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		var _WsMsg WsMsg
		err = json.Unmarshal(message, &_WsMsg)
		if err != nil {
			errMsg, _ := json.Marshal(&WsErrMsg{
				Action: "ErrMsg",
				Msg: ErrMsg{
					Msgid:  _WsMsg.Msg.Msgid,
					ErrMsg: err.Error(),
				},
			})
			select {
			case s.Send <- errMsg:
			default:
				return
			}
			break
		}
		fromuser := s.Touser

		if _WsMsg.Action == "chat" {
			//check touser is in memory map
			lock.Lock()
			isSet := s.Model.MemDB.Get(_WsMsg.Msg.Touser)
			if isSet == nil {
				//offline user
				s.Model.InitMemOfflineUser(&model.MemInitUser{
					Touser: _WsMsg.Msg.Touser,
				})
			}
			touser := s.Model.MemGetUserProfile(_WsMsg.Touser)
			fromuserProfile := s.Model.MemGetUserProfile(s.Touser)
			rest, err := s.Model.CheckOrSetFriends(model.ImFreindList{
				Touser:   _WsMsg.Msg.Touser,
				Fromuser: fromuser,
			})
			if err != nil {
				errMsg, _ := json.Marshal(&WsErrMsg{
					Action: "ErrMsg",
					Msg: ErrMsg{
						Msgid:  _WsMsg.Msg.Msgid,
						ErrMsg: err.Error(),
					},
				})
				select {
				case s.Send <- errMsg:
				default:
					return
				}
				lock.Unlock()
				continue
			}
			//new friend
			if rest {
				//set new friend
				s.Model.MemSetContactByTouser(s.Touser, fromuser, 0)
				s.Model.MemSetRecvCountDirectByTouser(s.Touser, fromuser, 2)

			}
			//check is in blacklist
			isInBlack := s.Model.MemIsInBlacklist(s.Touser, fromuser)
			if isInBlack {
				errMsg, _ := json.Marshal(&WsErrMsg{
					Action: "ErrMsg",
					Msg: ErrMsg{
						Msgid:  _WsMsg.Msg.Msgid,
						ErrMsg: "user in blacklist",
					},
				})
				select {
				case s.Send <- errMsg:
				default:
					return
				}
				lock.Unlock()
				continue
			}
			//check is exceed send limit
			rest, err = s.Model.MemSetGetRecvCountThresholdByTouser(s.Touser, fromuser)
			if err != nil {
				errMsg, _ := json.Marshal(&WsErrMsg{
					Action: "ErrMsg",
					Msg: ErrMsg{
						Msgid:  _WsMsg.Msg.Msgid,
						ErrMsg: err.Error(),
					},
				})
				select {
				case s.Send <- errMsg:
				default:
					return
				}
				lock.Unlock()
				continue
			}
			//can send
			if rest {
				//check is repeated msgid
				isSet, err := s.Model.SetMsgIdWithTTL(_WsMsg.Msg.Msgid)
				if err != nil {
					errMsg, _ := json.Marshal(&WsErrMsg{
						Action: "ErrMsg",
						Msg: ErrMsg{
							Msgid:  _WsMsg.Msg.Msgid,
							ErrMsg: err.Error(),
						},
					})
					select {
					case s.Send <- errMsg:
					default:
						return
					}
					lock.Unlock()
					continue
				}

				if isSet {
					totype := 0
					fromtype := s.Usertype
					if fromtype == 0 {
						totype = 1
					} else {
						totype = 0
					}
					lastid, err := s.Model.InsertMessages(model.ImMessage{
						Touser:   _WsMsg.Msg.Touser,
						Fromuser: fromuser,
						Msg:      _WsMsg.Msg.Msg,
						Fromtype: fromtype,
						Totype:   totype,
						Msgtype:  1,
					})
					if err != nil {
						errMsg, _ := json.Marshal(&WsErrMsg{
							Action: "ErrMsg",
							Msg: ErrMsg{
								Msgid:  _WsMsg.Msg.Msgid,
								ErrMsg: err.Error(),
							},
						})
						select {
						case s.Send <- errMsg:
						default:
							return
						}
						lock.Unlock()
						continue
					}
					//brodcast to fromuser and touser
					//source == 1 own send
					//source == 2 remoute send
					msg, _ := json.Marshal(&model.ImMessage{
						Msgid:            _WsMsg.Msg.Msgid,
						Touser:           _WsMsg.Msg.Touser,
						Fromuser:         fromuser,
						Msg:              _WsMsg.Msg.Msg,
						Fromtype:         fromtype,
						Totype:           totype,
						Msgtype:          1,
						ReadId:           lastid,
						Source:           1,
						TouserFullNameEn: touser.FullNameEn,
						TouserFullNameJa: touser.FullNameJa,
						TouserAvatar:     touser.Avatar,

						FromUserFullNameEn: fromuserProfile.FullNameEn,
						FromUserFullNameJa: fromuserProfile.FullNameJa,
						FromUserAvatar:     fromuserProfile.Avatar,
					})

					s.Model.MemDB.Get(fromuser).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).Iterator(func(k, v interface{}) bool {
						if k.(string) != s.SessionId {
							select {
							case v.(*model.MemInitUser).Send <- msg:
							default:
								close(v.(*model.MemInitUser).Send)
								c.WsUnregister(v.(*model.MemInitUser))
							}
						}
						return true
					})
					msg, _ = json.Marshal(&model.ImMessage{
						Msgid:            _WsMsg.Msg.Msgid,
						Touser:           _WsMsg.Msg.Touser,
						Fromuser:         fromuser,
						Msg:              _WsMsg.Msg.Msg,
						Fromtype:         fromtype,
						Totype:           totype,
						Msgtype:          1,
						ReadId:           lastid,
						Source:           2,
						TouserFullNameEn: touser.FullNameEn,
						TouserFullNameJa: touser.FullNameJa,
						TouserAvatar:     touser.Avatar,

						FromUserFullNameEn: fromuserProfile.FullNameEn,
						FromUserFullNameJa: fromuserProfile.FullNameJa,
						FromUserAvatar:     fromuserProfile.Avatar,
					})

					s.Model.MemDB.Get(_WsMsg.Msg.Touser).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).Iterator(func(k, v interface{}) bool {
						select {
						case v.(*model.MemInitUser).Send <- msg:
						default:
							close(v.(*model.MemInitUser).Send)
							c.WsUnregister(v.(*model.MemInitUser))
						}
						return true
					})
				}

			}
			successMsg, _ := json.Marshal(&WsMsg{
				Action: "MsgAck",
				MsgId:  _WsMsg.MsgId,
			})

			select {
			case s.Send <- successMsg:
			default:
				return
			}

			lock.Unlock()
		}

		//chat end
		if _WsMsg.Action == "chatack" {
			lock.Lock()
			fromuser := s.Touser
			isExists := s.Model.MemDB.Get(fromuser).(*gmap.AnyAnyMap).Get("Contact").(*gmap.AnyAnyMap).Get(_WsMsg.Touser)
			if isExists == nil {
				//不存在不可能ack
				errMsg, _ := json.Marshal(&WsErrMsg{
					Action: "ErrMsg",
					Msg: ErrMsg{
						Msgid:  _WsMsg.Msg.Msgid,
						ErrMsg: "err.Error()",
					},
				})
				select {
				case s.Send <- errMsg:
				default:
					return
				}
				lock.Unlock()
				continue
			}

			err = s.Model.SetLastReadId(model.SetLastReadId{
				Touser:   _WsMsg.Touser,
				Fromuser: fromuser,
				Id:       _WsMsg.LastReadId,
			})
			if err != nil {
				errMsg, _ := json.Marshal(&WsErrMsg{
					Action: "ErrMsg",
					Msg: ErrMsg{
						Msgid:  _WsMsg.Msg.Msgid,
						ErrMsg: "err.Error()",
					},
				})
				select {
				case s.Send <- errMsg:
				default:
					return
				}
				lock.Unlock()
				continue
			}
			successMsg, _ := json.Marshal(&WsMsg{
				Action: "ReadAck",
				MsgId:  _WsMsg.MsgId,
			})

			select {
			case s.Send <- successMsg:
			default:
				return
			}
			lock.Unlock()
		}

		//TODO get contact list
		//TODO move to black list
		//TODO get recent msgs

	}
}

func (c *Controller) WsWriteMsg(s *model.MemInitUser) {
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
				// The hub closed the channel.
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

type UserToken struct {
	Uuid     string
	Expired  int64
	UserType int
}

func (c *Controller) WsHandler(g *gin.Context) {
	//Check Header
	var (
		isChecked = true
		errMsg    = ""
	)

	auth := g.Request.Header.Get("Authorization")
	cyptde := crypto.FromBase64String(auth).SetKey(config.TOKEN_KEY).Aes().ECB().PKCS7Padding().Decrypt().ToString()
	var _userToken UserToken
	err := json.Unmarshal([]byte(cyptde), &_userToken)
	if err != nil {
		isChecked = false
		errMsg = err.Error()
	}

	conn, err := upgrader.Upgrade(g.Writer, g.Request, nil)
	if err != nil {
		return
	}

	if !isChecked {
		conn.WriteJSON(gin.H{"code": 10000, "msg": errMsg})
		conn.Close()
		return
	}

	//每个连接都有一个标识
	sessionId, _ := gonanoid.New(128)
	s := &model.MemInitUser{
		Touser:    _userToken.Uuid,
		Usertype:  _userToken.UserType,
		SessionId: sessionId,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Model:     c.M,
	}
	lock.Lock()
	err = c.M.MemAddNewUser(s)
	if err != nil {
		conn.WriteJSON(gin.H{"code": 1, "msg": "error msg"})
		conn.Close()
		lock.Unlock()
		return
	}
	lock.Unlock()

	go c.WsReadMsg(s)
	go c.WsWriteMsg(s)

}
