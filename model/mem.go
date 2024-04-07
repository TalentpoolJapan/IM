package model

import (
	"errors"
	"fmt"

	mysqldb "imserver/db/mysql"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gorilla/websocket"
)

type MemInitUser struct {
	Touser string
	//用户类型
	Usertype int
	//sessionId
	SessionId string
	//Conn
	Conn *websocket.Conn
	//Send
	Send chan []byte
	//Model
	Model *Model
}

func (m *Model) MemAddNewUser(user *MemInitUser) (err error) {
	isInitUser := m.MemDB.Get(user.Touser).(*gmap.AnyAnyMap).Get("InitUser")
	if isInitUser != nil {
		if isInitUser.(int) == 1 {
			m.MemSetConnByTouser(user)
			return
		}
	}
	m.MemDB.SetIfNotExistFuncLock(user.Touser, func() interface{} {
		var (
			node1 = gmap.New(true)
			node2 = gmap.New(true)
			node3 = gmap.New(true)
			node4 = gmap.New(true)
			node5 = gmap.New(true)
		)
		//存放所有的conn连接指针 node2:sessionId,conn
		node1.Set("Conn", node2)
		// //存放所有的黑名单
		node1.Set("Blacklist", node3)
		//存放我还能接收你发送的消息的数量
		node1.Set("RecvCount", node4)
		//存放所有联系人的uuid
		node1.Set("Contact", node5)
		//存放当前用户的profile
		node1.Set("Profile", user)
		//还没有初始化的用户
		node1.Set("InitUser", 0)
		return node1
	})
	m.MemSetUserProfile(user)
	err = m.MemInitTouserFromDB(user.Touser)
	if err != nil {
		return
	}
	m.MemDB.Get(user.Touser).(*gmap.AnyAnyMap).Set("InitUser", 1)

	m.MemSetConnByTouser(user)
	return
}

func (m *Model) InitMemOfflineUser(user *MemInitUser) (err error) {
	isInitUser := m.MemDB.Get(user.Touser).(*gmap.AnyAnyMap).Get("InitUser")
	if isInitUser != nil {
		if isInitUser.(int) == 1 {
			return
		}
	}
	m.MemDB.SetIfNotExistFuncLock(user.Touser, func() interface{} {
		var (
			node1 = gmap.New(true)
			node2 = gmap.New(true)
			node3 = gmap.New(true)
			node4 = gmap.New(true)
			node5 = gmap.New(true)
		)
		//存放所有的conn连接指针 node2:sessionId,conn
		node1.Set("Conn", node2)
		// //存放所有的黑名单
		node1.Set("Blacklist", node3)
		//存放我还能接收你发送的消息的数量
		node1.Set("RecvCount", node4)
		//存放所有联系人的uuid
		node1.Set("Contact", node5)
		//存放当前用户的profile
		node1.Set("Profile", user)
		//还没有初始化的用户
		node1.Set("InitUser", 0)
		return node1
	})
	m.MemSetUserProfile(user)
	err = m.MemInitTouserFromDB(user.Touser)
	if err != nil {
		return
	}
	m.MemDB.Get(user.Touser).(*gmap.AnyAnyMap).Set("InitUser", 1)
	return
}

func (m *Model) MemGetUserProfile(touser string) mysqldb.UserBasicInfo {
	return m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("Profile").(mysqldb.UserBasicInfo)
}

func (m *Model) MemSetUserProfile(user *MemInitUser) (err error) {
	data, err := m.MysqlDB.GetUserProfileByUser(mysqldb.UserProfile{
		UserType: user.Usertype,
		Uuid:     []string{user.Touser},
	})
	if err != nil {
		return
	}
	if len(data) > 0 {
		m.MemDB.Get(user.Touser).(*gmap.AnyAnyMap).Set("Profile", data[0])
	}
	return errors.New("can not find this user")
}

func (m *Model) MemInitTouserFromDB(touser string) (err error) {
	var im_friend_list []ImFreindList
	res, err := m.FulltextDB.Query(fmt.Sprintf(`select * from im_friend_list where match('@touser %s');`, touser), &im_friend_list)
	if err != nil {
		return
	}
	for _, v := range *res.(*[]ImFreindList) {
		m.MemSetContactByTouser(v.Touser, v.Fromuser, v.Isblack)
		if v.Isblack == 1 {
			m.MemSetBlacklistByTouser(v.Touser, v.Fromuser)
		}
		m.MemSetRecvCountDirectByTouser(v.Touser, v.Fromuser, v.Count)
	}
	return
}

func (m *Model) MemSetContactByTouser(touser, fromuser string, isBlack int) {
	m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("Contact").(*gmap.AnyAnyMap).Set(fromuser, isBlack)
}

func (m *Model) MemSetBlacklistByTouser(touser, fromuser string) {
	m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("Blacklist").(*gmap.AnyAnyMap).Set(fromuser, 1)
}

// 黑名单是1 不在是0
func (m *Model) MemMoveFriendToBlacklistByTouser(touser, fromuser string) {
	m.MemSetContactByTouser(touser, fromuser, 1)
	m.MemSetBlacklistByTouser(touser, fromuser)
}

func (m *Model) MemMoveBlacklistToFriendByTouser(touser, fromuser string) {
	m.MemSetContactByTouser(touser, fromuser, 0)
	m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("Blacklist").(*gmap.AnyAnyMap).Remove(fromuser)
}

func (m *Model) MemIsInBlacklist(touser, fromuser string) bool {
	return m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("Blacklist").(*gmap.AnyAnyMap).Contains(fromuser)
}

func (m *Model) MemSetRecvCountDirectByTouser(touser, fromuser string, count int) {
	m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("RecvCount").(*gmap.AnyAnyMap).Set(fromuser, count)
}

// 设置获取单个对话消息数量阈值
func (m *Model) MemSetGetRecvCountThresholdByTouser(touser, fromuser string) (iSet bool, err error) {
	var (
		count    int  = 0
		isUpdate bool = true
	)
	recvCount := m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("RecvCount").(*gmap.AnyAnyMap)
	recvCount.RLockFunc(func(s map[interface{}]interface{}) {
		for k, v := range s {
			if fromuser == k.(string) {
				count = v.(int) - 1
				if count >= 0 {
					recvCount.Set(fromuser, count)
					err = m.SetRecvCountByTouser(touser, fromuser, count)
					if err != nil {
						//rollback
						recvCount.Set(fromuser, count+1)
					}
					isUpdate = true
				}
			}
		}
	})
	if isUpdate {
		return true, nil
	}
	return !isUpdate, err
}

func (m *Model) MemSetConnByTouser(user *MemInitUser) {
	m.MemDB.Get(user.Touser).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).Set(user.SessionId, user.Conn)
}

func (m *Model) MemRemoveConnByTouser(touser, sessionId string) {
	m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).Remove(sessionId)
}
