package persistence

import (
	"fmt"
	"imserver/internal/domain/immessage"
	"time"
	"xorm.io/xorm"
)

// region struct
type imMessagePO struct {
	Id        int64                 `json:"id,omitempty"`
	Sessionid string                `json:"sessionid,omitempty"`
	Touser    string                `json:"touser,omitempty"`
	Fromuser  string                `json:"fromuser,omitempty"`
	Msg       string                `json:"msg,omitempty"`
	Msgtype   immessage.MessageType `json:"msgtype,omitempty"`
	MsgCode   string                `json:"msg_code,omitempty"`
	Totype    int                   `json:"totype,omitempty"`
	Fromtype  int                   `json:"fromtype,omitempty"`
	Created   int64                 `json:"created,omitempty"`
	Msgid     string                `json:"msgid,omitempty"`
}

func convertImMessageEntity(po *imMessagePO) *immessage.ImMessage {
	return &immessage.ImMessage{
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

// endregion

// region implement

func NewManticoreImMessageRepo(engine *xorm.Engine) immessage.ImMessageRepository {
	return &ManticoreMessageRepo{
		ManticoreDB: engine,
	}
}

type ManticoreMessageRepo struct {
	ManticoreDB *xorm.Engine
}

func (r ManticoreMessageRepo) ListMessageRecentBy(sessionId string, size int, userUuid string) ([]*immessage.ImMessage, error) {
	var querySize = size
	if querySize <= 0 {
		querySize = 10
	}
	var data []imMessagePO
	sql := fmt.Sprintf(`select * from im_message where sessionid='%s' and touser='%s' order by created desc limit %d`, sessionId, userUuid, querySize)
	err := r.ManticoreDB.SQL(sql).Find(&data)
	if err != nil {
		return nil, err
	}
	var result []*immessage.ImMessage
	for _, item := range data {
		result = append(result, convertImMessageEntity(&item))
	}
	return result, nil
}

func (r ManticoreMessageRepo) ListMessageRecent(sessionId string, size int) ([]*immessage.ImMessage, error) {
	var querySize = size
	if querySize <= 0 {
		querySize = 10
	}
	var data []imMessagePO
	sql := fmt.Sprintf(`select * from im_message where match('@sessionid %s') order by created desc limit %d`, sessionId, querySize)
	err := r.ManticoreDB.SQL(sql).Find(&data)
	if err != nil {
		return nil, err
	}
	var result []*immessage.ImMessage
	for _, item := range data {
		result = append(result, convertImMessageEntity(&item))
	}
	return result, nil
}

func (r ManticoreMessageRepo) ListMessageBeforeCreateTime(sessionId string, createTime int64) ([]*immessage.ImMessage, error) {
	var data []imMessagePO
	sql := fmt.Sprintf(`select * from im_message where match('@sessionid %s') and created<%d order by created asc`, sessionId, createTime)
	err := r.ManticoreDB.SQL(sql).Find(&data)
	if err != nil {
		return nil, err
	}
	var result []*immessage.ImMessage
	for _, item := range data {
		result = append(result, convertImMessageEntity(&item))
	}
	return result, nil
}

func (r ManticoreMessageRepo) ListMessageAfterCreateTime(sessionId string, createTime int64) ([]*immessage.ImMessage, error) {
	var data []imMessagePO
	sql := fmt.Sprintf(`select * from im_message where match('@sessionid %s') and created>%d order by created asc`, sessionId, createTime)
	err := r.ManticoreDB.SQL(sql).Find(&data)
	if err != nil {
		return nil, err
	}
	var result []*immessage.ImMessage
	for _, item := range data {
		result = append(result, convertImMessageEntity(&item))
	}
	return result, nil
}

func (r ManticoreMessageRepo) GetMessageByClientMsgId(sessionId string, clientMsgId string) (*immessage.ImMessage, error) {
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

func (r ManticoreMessageRepo) ListMessageAfterMsgId(sessionId string, id int64) ([]*immessage.ImMessage, error) {
	var data []imMessagePO
	sql := fmt.Sprintf(`select * from im_message where match('@sessionid %s') and id>%s order by id asc`, sessionId, id)
	err := r.ManticoreDB.SQL(sql).Find(&data)
	if err != nil {
		return nil, err
	}
	var result []*immessage.ImMessage
	for _, item := range data {
		result = append(result, convertImMessageEntity(&item))
	}
	return result, nil
}

func (r ManticoreMessageRepo) LatestImMessageBySessionId(sessionIds []string) ([]*immessage.ImMessage, error) {
	var result []*immessage.ImMessage
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

func (r ManticoreMessageRepo) SaveImMessage(imMessage immessage.ImMessage) (int64, error) {
	sql := fmt.Sprintf(`insert into im_message (sessionid,touser,fromuser,msg,msgtype,totype,fromtype,created,msgid) values ('%s','%s','%s','%s',%d,%d,%d,%d,'%s')`,
		imMessage.SessionId, imMessage.ToUser, imMessage.FromUser, imMessage.Msg, imMessage.MsgType, imMessage.ToType, imMessage.FromType, time.Now().UnixMicro(), imMessage.MsgId)
	res, err := r.ManticoreDB.Exec(sql)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// endregion
