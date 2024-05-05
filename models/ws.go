package models

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gorilla/websocket"
)

// 检查是否初始化用户
func (m *Model) CheckIsInitUser(userUUID string) bool {
	isInitUser := m.MemDB.Get(userUUID)
	fmt.Println("检查初始化用户:", isInitUser)
	return isInitUser == nil
}

type InitUser struct {
	ConnTag string
	Conn    *websocket.Conn
	UUID    string
	//用户类型
	Usertype int
	//Send
	Send chan []byte
	//Model
	Model *Model
}

type RecvStatus struct {
	//还能接收的消息数量
	RecvCount int
	//如果接收数量为0那么下次可继续接收消息的时间
	NextRecvTime int64
}

// 初始化用户 true 初始化成功 false 初始化失败 golang bool默认值是false
func (m *Model) InitUser(user *InitUser) (isSuccess bool, err error) {
	m.MemDB.SetIfNotExistFuncLock(user.UUID, func() interface{} {
		var (
			node1 = gmap.New(true)
			node2 = gmap.New(true)
			node3 = gmap.New(true)
			node4 = gmap.New(true)
			//node5 = gmap.New(true)
		)
		//存放所有的conn连接指针 node2:sessionId,conn
		node1.Set("Conn", node2)
		// //存放所有的黑名单
		node1.Set("Blacklist", node3)
		//存放我还能接收你发送的消息的数量 //node4.Set(remoteUUID,UserRecvMsg)
		node1.Set("Contacts", node4)
		//存放所有联系人的uuid
		//node1.Set("Contacts", node5)
		// //存放当前用户的profile
		// node1.Set("Profile", UserProfile{
		// 	UUID: user.UUID,
		// })
		return node1
	})

	//读取用户的Profile

	//UUID <- REMOTE_UUID RecvCount
	return true, nil
}

type UserBasicInfo struct {
	Uuid       string `json:"uuid" xorm:"not null default '''' unique CHAR(36)"`
	FullNameEn string `json:"full_name_en" xorm:"default 'NULL' VARCHAR(64)"`
	FullNameJa string `json:"full_name_ja" xorm:"default 'NULL' VARCHAR(64)"`
	Avatar     string `json:"avatar" xorm:"default '''' VARCHAR(128)"`
}
type UserToken struct {
	Uuid     string
	Expired  int64
	UserType int
}

func (m *Model) GetUserProfileByUUID(user UserToken) (userinfo UserBasicInfo, err error) {
	var ok bool
	if user.UserType == 0 {
		ok, err = m.MySQLDB.SQL("select uuid,company_name_en as full_name_en,company_name_ja as full_name_ja,logo as avatar from enterprise_account_basic_info where uuid=?", user.Uuid).Get(&userinfo)
	}
	if user.UserType == 1 {
		ok, err = m.MySQLDB.SQL("select uuid,nick_name_oversea as full_name_en,nick_name_ja as full_name_ja,logo as avatar from job_seeker_basic_info where uuid=?", user.Uuid).Get(&userinfo)
	}
	if err != nil {
		return
	}

	if !ok {
		err = fmt.Errorf("NONE_USER")
	}

	return

}

type ImFreindList struct {
	Id       int64
	Touser   string
	Fromuser string
	Isblack  int
	Count    int
	// Status 预留添加好友字段
	Status   int
	Created  int64
	Nexttime int64
}

// 获取用户联系人
func (m *Model) GetUserContactsByUUID(uuid string) (friendList []ImFreindList, err error) {
	sql := fmt.Sprintf(`select * from im_friend_list where match('@touser %s') order by isblack asc;`, uuid)
	err = m.ManticoreDB.SQL(sql).Find(&friendList)
	return
}

// 聊天sessionID
func (m *Model) GetSessionId(uuid, touser string) (string, error) {
	var idstr string
	rst := strings.Compare(touser, uuid)
	//小的放在前面就行了
	//sessionId.Touser<sessionId.Fromuser
	if rst == -1 {
		idstr = touser + uuid
	}
	if rst == 1 {
		idstr = uuid + touser
	}

	if rst == 0 {
		return "", errors.New("touser == fromuser")
	}
	hash := md5.Sum([]byte(idstr))
	md5Str := hex.EncodeToString(hash[:])
	return md5Str, nil
}

// 插入聊天记录
func (m *Model) InsertMessages(s *InitUser, touser, msg, msgid string) (lastId int64, err error) {
	sessionId, err := m.GetSessionId(s.UUID, touser)
	if err != nil {
		return 0, err
	}
	toType := 0
	if s.Usertype == 0 {
		toType = 1
	}

	sql := fmt.Sprintf(`insert into im_message (sessionid,touser,fromuser,msg,msgtype,totype,fromtype,created,msgid) values ('%s','%s','%s','%s',%d,%d,%d,%d,'%s')`,
		sessionId, touser, s.UUID, msg, 1, toType, s.Usertype, time.Now().Unix(), msgid)
	res, err := m.ManticoreDB.Exec(sql)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// 直接赋值最后消息阅读ID
func (m *Model) SetLastReadId(uuid, touser string, lastid uint) {
	relatedKey := fmt.Sprintf(`%s<-%s`, touser, uuid)
	lastReadid := fmt.Sprintf("%d", lastid)

	m.NoSqlDB.Set(relatedKey, lastReadid)
}

type ImMessage struct {
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

// 我和它的聊天消息，新/旧/最近，起始ID
func (m *Model) GetMessagesByStartId(uuid, touser, method string, readId uint64) (data []ImMessage, err error) {
	var (
		sql string
	)
	sessionId, err := m.GetSessionId(touser, uuid)
	if err != nil {
		return
	}

	if method == "new" {
		sql = fmt.Sprintf(`select * from im_message where match('@sessionid %s') and id>%d order by id asc`, sessionId, readId)
	} else if method == "old" {
		sql = fmt.Sprintf(`select * from im_message where match('@sessionid %s') and id<%d order by id desc limit 10`, sessionId, readId)
	} else if method == "recent" {
		sql = fmt.Sprintf(`select * from im_message where match('@sessionid %s') order by id desc limit 10`, sessionId)
	} else {
		err = fmt.Errorf("this method->%s is not supported", method)
		return
	}
	err = m.ManticoreDB.SQL(sql).Find(&data)
	if err != nil {
		return
	}
	return
}

// 更新数据库发送次数
func (m *Model) SetRecvCountByTouser(touser, fromuser string, count int, nextRecvTime int64) (err error) {
	sql := fmt.Sprintf(`update im_friend_list set count=%d,nexttime=%d where match('@touser %s fromuser %s')`, count, nextRecvTime, touser, fromuser)
	_, err = m.ManticoreDB.Exec(sql)
	return
}

// 更新黑名单
func (m *Model) SetBlacklistByTouser(touser, fromuser string, isBlack int) (err error) {
	sql := fmt.Sprintf(`update im_friend_list set isblack=%d where match('@touser %s fromuser %s')`, isBlack, touser, fromuser)
	_, err = m.ManticoreDB.Exec(sql)
	return
}

type MyContacts struct {
	Uuid       string `json:"uuid" xorm:"not null default '''' unique CHAR(36)"`
	FullNameEn string `json:"full_name_en" xorm:"default 'NULL' VARCHAR(64)"`
	FullNameJa string `json:"full_name_ja" xorm:"default 'NULL' VARCHAR(64)"`
	Avatar     string `json:"avatar" xorm:"default '''' VARCHAR(128)"`
	IsBlack    int    `json:"isblack"`
}
