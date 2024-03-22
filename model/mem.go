package model

import (
	"fmt"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gorilla/websocket"
)

//MARK sessionId = encypt key

type MemInitUser struct {
	Touser string
}

func (m *Model) MemAddNewUser(user MemInitUser) (err error) {
	exist := m.MemDB.SetIfNotExistFuncLock(user.Touser, func() interface{} {
		var (
			node1 = gmap.New(true)
			node2 = gmap.New(true)
			node3 = gmap.New(true)
			node4 = gmap.New(true)
			node5 = gmap.New(true)
		)
		//存放所有的conn连接指针
		node1.Set("Conn", node2)
		// //存放所有的黑名单
		node1.Set("Blacklist", node3)
		//存放还能接收消息的数量
		node1.Set("RecvCount", node4)
		//存放所有联系人的uuid
		node1.Set("Contact", node5)
		return node1
	})
	if !exist {
		err = m.MemInitTouserFromDB(user.Touser)
		if err != nil {
			return
		}
	}
	return
}

func (m *Model) MemSetAllUserBasicInfo() {
	//TODO GET ALL users from DB
	m.MemDB.SetIfNotExistFuncLock("UserList", func() interface{} {
		var (
			node1 = gmap.New(true)
			node2 = gmap.New(true)
		)
		node1.Set("Conn", node2)
		return node1
	})
}

func (m *Model) MemGetUserProfile(touser string) string {
	return m.MemDB.Get(touser).(string)
}

func (m *Model) MemSetUserProfile(touserjson string) {
	m.MemDB.Set("User", touserjson)
}

func (m *Model) MemGetAllContactsProfiles(touser string) {

}
func (m *Model) MemGetContactProfileByFromUser(fromuser string) {

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
		m.MemSetRecvCountByTouser(v.Touser, v.Fromuser, v.Count)
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

func (m *Model) MemSetRecvCountByTouser(touser, fromuser string, count int) {
	m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("RecvCount").(*gmap.AnyAnyMap).Set(fromuser, count)
}

func (m *Model) MemSetConnByTouser(touser, sessionId string, conn *websocket.Conn) {
	m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).Set(sessionId, conn)
}

func (m *Model) MemRemoveConnByTouser(touser, sessionId string) {
	m.MemDB.Get(touser).(*gmap.AnyAnyMap).Get("Conn").(*gmap.AnyAnyMap).Remove(sessionId)
}
