package persistence

import (
	"fmt"
	"imserver/internal/domain/user"
	"xorm.io/xorm"
)

// region struct
type imMessagePO struct {
	Id        int64  `json:"id,omitempty"`
	Sessionid string `json:"sessionid,omitempty"`
	Touser    string `json:"touser,omitempty"`
	Fromuser  string `json:"fromuser,omitempty"`
	Msg       string `json:"msg,omitempty"`
	Msgtype   int    `json:"msgtype,omitempty"`
	Totype    int    `json:"totype,omitempty"`
	Fromtype  int    `json:"fromtype,omitempty"`
	Created   int64  `json:"created,omitempty"`
	Msgid     string `json:"msgid,omitempty"`
}

func convertImMessageEntity(po *imMessagePO) *user.ImMessage {
	return &user.ImMessage{
		Id:        po.Id,
		SessionId: po.Sessionid,
		ToUser:    po.Touser,
		FromUser:  po.Fromuser,
		Msg:       po.Msg,
		MsgType:   po.Msgtype,
		ToType:    po.Totype,
		FromType:  po.Fromtype,
		Created:   po.Created,
		MsgId:     po.Msgid,
	}
}
func convertImMessagePO(msg *user.ImMessage) *imMessagePO {
	return &imMessagePO{
		Id:        msg.Id,
		Sessionid: msg.SessionId,
		Touser:    msg.ToUser,
		Fromuser:  msg.FromUser,
		Msg:       msg.Msg,
		Msgtype:   msg.MsgType,
		Totype:    msg.ToType,
		Fromtype:  msg.FromType,
		Created:   msg.Created,
		Msgid:     msg.MsgId,
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

func (r ManticoreMessageRepo) GetMessageByClientMsgId(sessionId string, clientMsgId string) (*user.ImMessage, error) {
	var data []imMessagePO
	sql := fmt.Sprintf(`select * from im_message where match('@sessionid %s') and msgid='%s'`, sessionId, clientMsgId)
	err := r.ManticoreDB.SQL(sql).Find(&data)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		return convertImMessageEntity(&data[0]), nil
	}
	return nil, nil
}

func (r ManticoreMessageRepo) ListMessageAfterMsgId(sessionId string, id int64) ([]user.ImMessage, error) {
	var data []imMessagePO
	sql := fmt.Sprintf(`select * from im_message where match('@sessionid %s') and id>%s order by id asc`, sessionId, id)
	err := r.ManticoreDB.SQL(sql).Find(&data)
	if err != nil {
		return nil, err
	}
	var result []user.ImMessage
	for _, item := range data {
		result = append(result, *convertImMessageEntity(&item))
	}
	return result, nil
}

func (r ManticoreMessageRepo) LatestImMessageBySessionId(sessionIds []string) ([]*user.ImMessage, error) {
	var result []*user.ImMessage
	for _, sessionId := range sessionIds {
		var imMessage []*imMessagePO
		sql := fmt.Sprintf(`select * from im_message where match('@sessionid %s') order by id desc limit 1`, sessionId)
		err := r.ManticoreDB.SQL(sql).Find(&imMessage)
		if err == nil && len(imMessage) > 0 {
			result = append(result, convertImMessageEntity(imMessage[0]))
		}
	}
	return result, nil
}

// endregion
