package controller

import (
	"bytes"
	"fmt"
	"imserver/model"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *Controller) WsUnregister(s *model.MemInitUser) {
	c.M.WsUnregister(s)
}

func (c *Controller) WsReadMsg(s *model.MemInitUser) {
	defer func() {
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
		//TODO 逻辑
		//Test Echo
		s.Conn.WriteJSON(gin.H{"msg": string(message)})
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
			//TODO解析发送
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

func (c *Controller) WsHandler(g *gin.Context) {
	fmt.Println(g.Request.Header)
	conn, err := upgrader.Upgrade(g.Writer, g.Request, nil)
	if err != nil {
		return
	}

	//TODO
	//Check Header
	var (
		isChecked = true
	)
	if !isChecked {
		conn.WriteJSON(gin.H{"code": 1, "msg": "error msg"})
		conn.Close()
		return
	}

	s := &model.MemInitUser{
		Touser:    "uuid",
		Fullname:  "fullname",
		Avatar:    "avatar",
		Usertype:  0,
		SessionId: "TODO BUILD SESSIONID",
		Conn:      conn,
		Send:      make(chan []byte, 256),
	}

	err = c.M.MemAddNewUser(s)
	if err != nil {
		conn.WriteJSON(gin.H{"code": 1, "msg": "error msg"})
		conn.Close()
		return
	}

	go c.WsReadMsg(s)
	go c.WsWriteMsg(s)

}
