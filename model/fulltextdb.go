package model

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Describe struct {
	Field      string
	Type       string
	Properties string
}

func (m *Model) GetDescribe(table string) (*[]Describe, error) {
	var data []Describe
	res, err := m.FulltextDB.Query(fmt.Sprintf("DESCRIBE %s", table), &data)
	if err != nil {
		return res.(*[]Describe), err
	}
	return res.(*[]Describe), nil
}

//需要转译的字符串：!    "    $    '    (    )    -    /    <    @    \    ^    |    ~ 转译为空字符串

// CREATE TABLE im_message (
// 	id bigint,
//  sessionid text
// 	touser text,
// 	fromuser text,
// 	msg text,
// 	msgtype integer,
// 	totype integer,
// 	fromtype integer,
// 	created bigint
// 	) prefix_fields='msg' min_prefix_len='2'  ngram_len='1' ngram_chars='cjk'

type ImMessage struct {
	Id               int64  `json:"id,omitempty"`
	SessionId        string `json:"sessionId,omitempty"`
	Touser           string `json:"touser,omitempty"`
	Fromuser         string `json:"fromuser,omitempty"`
	Msg              string `json:"msg,omitempty"`
	Msgtype          int    `json:"msgtype,omitempty"`
	Totype           int    `json:"totype,omitempty"`
	Fromtype         int    `json:"fromtype,omitempty"`
	Created          int64  `json:"created,omitempty"`
	Msgid            string `json:"msgid,omitempty"`
	ReadId           int64  `json:"readid,omitempty"`
	Source           int    `json:"source,omitempty"`
	TouserFullNameEn string `json:"touserfullnameen,omitempty"`
	TouserFullNameJa string `json:"touserfullnameja,omitempty"`
	TouserAvatar     string `json:"touseravatar,omitempty"`

	FromUserFullNameEn string `json:"fromuserfullnameen,omitempty"`
	FromUserFullNameJa string `json:"fromuserfullnameja,omitempty"`
	FromUserAvatar     string `json:"fromuseravatar,omitempty"`
}

func (m *Model) InsertMessages(msg ImMessage) (lastId int64, err error) {
	sessionId, err := m.GetSessionId(GetSessionId{
		Touser:   msg.Touser,
		Fromuser: msg.Fromuser,
	})
	if err != nil {
		return 0, err
	}

	sql := fmt.Sprintf(`insert into im_message (sessionid,touser,fromuser,msg,msgtype,totype,fromtype,created,msgid) values ('%s','%s','%s','%s',%d,%d,%d,%d,'%s')`,
		sessionId, msg.Touser, msg.Fromuser, msg.Msg, msg.Msgtype, msg.Totype, msg.Fromtype, msg.Created, msg.Msgid)
	res, err := m.FulltextDB.Exec(sql)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type MessagesByStartId struct {
	Id       int64
	Touser   string
	Fromuser string
	Method   string
}

func (m *Model) GetMessagesByStartId(msgId MessagesByStartId) (*[]ImMessage, error) {
	var (
		data []ImMessage
		sql  string
	)
	sessionId, err := m.GetSessionId(GetSessionId{
		Touser:   msgId.Touser,
		Fromuser: msgId.Fromuser,
	})
	if err != nil {
		return &data, err
	}

	if msgId.Method == "new" {
		sql = fmt.Sprintf(`select * from im_message where match('@sessionid %s') and id>%d order by id asc`, sessionId, msgId.Id)
	} else if msgId.Method == "old" {
		sql = fmt.Sprintf(`select * from im_message where match('@sessionid %s') and id<%d order by id desc limit 10`, sessionId, msgId.Id)
	} else if msgId.Method == "recent" {
		sql = fmt.Sprintf(`select * from im_message where match('@sessionid %s') order by id desc limit 10`, sessionId)
	} else {
		return &data, fmt.Errorf("this method->%s is not supported", msgId.Method)
	}
	res, err := m.FulltextDB.Query(sql, &data)
	if err != nil {
		return res.(*[]ImMessage), err
	}
	return res.(*[]ImMessage), nil
}

func (m *Model) GetAllMessages() (*[]ImMessage, error) {
	var data []ImMessage
	res, err := m.FulltextDB.Query("select * from im_message", &data)
	if err != nil {
		return res.(*[]ImMessage), err
	}
	return res.(*[]ImMessage), nil
}

func (m *Model) ClearAllMessages() (err error) {
	msgs, err := m.GetAllMessages()
	if err != nil {
		return err
	}
	session := m.FulltextDB.GetSession()
	defer session.Close()
	err = session.Begin()
	for _, v := range *msgs {
		_, err = m.FulltextDB.Exec(fmt.Sprintf(`delete from im_message where id=%d`, v.Id))
		if err != nil {
			session.Rollback()
			return
		}
	}
	err = session.Commit()
	return
}

type GetSessionId struct {
	Touser   string
	Fromuser string
}

func (m *Model) GetSessionId(sessionId GetSessionId) (string, error) {
	var idstr string
	rst := strings.Compare(sessionId.Touser, sessionId.Fromuser)
	//小的放在前面就行了
	//sessionId.Touser<sessionId.Fromuser
	if rst == -1 {
		idstr = sessionId.Touser + sessionId.Fromuser
	}
	if rst == 1 {
		idstr = sessionId.Fromuser + sessionId.Touser
	}

	if rst == 0 {
		return "", errors.New("touser == fromuser")
	}
	hash := md5.Sum([]byte(idstr))
	md5Str := hex.EncodeToString(hash[:])
	return md5Str, nil
}

// isblack 1=好友 2=黑名单
// status 计数器 我还能接收你发的多少条消息 我<-你
type ImFreindList struct {
	Id       int64
	Touser   string
	Fromuser string
	Isblack  int
	Count    int
	// Status 预留添加好友字段
	Status  int
	Created int64
}

func (m *Model) GetAllFreinds() (*[]ImFreindList, error) {
	var data []ImFreindList
	res, err := m.FulltextDB.Query("select * from im_friend_list", &data)
	if err != nil {
		return res.(*[]ImFreindList), err
	}
	return res.(*[]ImFreindList), nil
}

// TODO
func (m *Model) SearchP2PChatMsg(touser, fromuser, msg string) (*[]ImMessage, error) {
	var (
		data []ImMessage
	)
	sessionId, err := m.GetSessionId(GetSessionId{
		Touser:   touser,
		Fromuser: fromuser,
	})
	if err != nil {
		return &data, err
	}
	sql := fmt.Sprintf(`select * from im_message where match('@msg *%s* @sessionid %s') group by id within group order by id desc`,
		msg, sessionId)
	res, err := m.FulltextDB.Query(sql, &data)
	if err != nil {
		return res.(*[]ImMessage), err
	}

	return res.(*[]ImMessage), nil

}
func (m *Model) SearchChatMsg(touser, msg string) (*[]ImMessage, error) {
	var (
		data   []ImMessage
		data_f []ImMessage
		data_s []ImMessage
	)
	sql := fmt.Sprintf(`select * from im_message where match('@msg *%s* @touser %s') group by id within group order by id desc`,
		msg, touser)
	res_f, err := m.FulltextDB.Query(sql, &data_f)
	if err != nil {
		return res_f.(*[]ImMessage), err
	}
	sql = fmt.Sprintf(`select * from im_message where match('@msg *%s* @fromuser %s') group by id within group order by id desc`,
		msg, touser)
	res_s, err := m.FulltextDB.Query(sql, &data_s)
	if err != nil {
		return res_s.(*[]ImMessage), err
	}
	//搜索第一遍我<-你
	for _, v := range *res_f.(*[]ImMessage) {
		data = append(data, v)
	}
	//搜索第二遍我->你
	for _, v := range *res_s.(*[]ImMessage) {
		data = append(data, v)
	}
	//根据ID从大到小排序
	sort.SliceStable(data, func(i, j int) bool {
		return data[i].Id > data[j].Id
	})
	return &data, nil
}

func (m *Model) SetBlacklistByTouser(touser, fromuser string, isBlack int) (err error) {
	sql := fmt.Sprintf(`update im_friend_list set isblack=%d where match('@touser %s fromuser %s')`, isBlack, touser, fromuser)
	_, err = m.FulltextDB.Exec(sql)
	return
}

func (m *Model) SetRecvCountByTouser(touser, fromuser string, count int) (err error) {
	sql := fmt.Sprintf(`update im_friend_list set count=%d where match('@touser %s fromuser %s')`, count, touser, fromuser)
	_, err = m.FulltextDB.Exec(sql)
	return
}
