package persistence

import (
	"fmt"
	"imserver/internal/domain/user"
	"xorm.io/xorm"
)

// region struct
type imMessagePO struct {
	Id        int64  `json:"id,omitempty"`
	SessionId string `json:"sessionid,omitempty"`
	ToUser    string `json:"touser,omitempty"`
	FromUser  string `json:"fromuser,omitempty"`
	Msg       string `json:"msg,omitempty"`
	MsgType   int    `json:"msgtype,omitempty"`
	ToType    int    `json:"totype,omitempty"`
	FromType  int    `json:"fromtype,omitempty"`
	Created   int64  `json:"created,omitempty"`
	MsgId     string `json:"msgid,omitempty"`
}

func convertImMessageEntity(po *imMessagePO) *user.ImMessage {
	return &user.ImMessage{
		Id:        po.Id,
		SessionId: po.SessionId,
		ToUser:    po.ToUser,
		FromUser:  po.FromUser,
		Msg:       po.Msg,
		MsgType:   po.MsgType,
		ToType:    po.ToType,
		FromType:  po.FromType,
		Created:   po.Created,
		MsgId:     po.MsgId,
	}
}
func convertImMessagePO(msg *user.ImMessage) *imMessagePO {
	return &imMessagePO{
		Id:        msg.Id,
		SessionId: msg.SessionId,
		ToUser:    msg.ToUser,
		FromUser:  msg.FromUser,
		Msg:       msg.Msg,
		MsgType:   msg.MsgType,
		ToType:    msg.ToType,
		FromType:  msg.FromType,
		Created:   msg.Created,
		MsgId:     msg.MsgId,
	}
}

// endregion

// region implement

func NewManticoreImMessageRepo(engine *xorm.Engine) user.ImMessageRepository {
	return &ManticoreMessageRepo{
		ManticoreDB: engine,
	}
}

type ManticoreMessageRepo struct {
	ManticoreDB *xorm.Engine
}

func (r ManticoreMessageRepo) LatestImMessageBySessionId(sessionIds []string) ([]*user.ImMessage, error) {
	var result []*user.ImMessage
	for _, sessionId := range sessionIds {
		var imMessage []*imMessagePO
		sql := fmt.Sprintf(`select * from im_message where match('@sessionid %s') order by id desc limit 1`, sessionId)
		err := r.ManticoreDB.SQL(sql).Find(&imMessage)
		if err != nil && len(imMessage) > 0 {
			result = append(result, convertImMessageEntity(imMessage[0]))
		}
	}
	return result, nil
}

// endregion
