package persistence

import (
	"fmt"
	"imserver/internal/domain/imfriend"
	"time"
	"xorm.io/xorm"
)

// region struct
type imFriendPO struct {
	Id       int64
	Touser   string
	Fromuser string
	Isblack  int
	Count    int
	// Status 预留添加好友字段
	Status        int
	Created       int64
	Nexttime      int64
	LastReadMsgId string `json:"last_read_msg_id" xorm:"last_read_msg_id"`
	EverContacted bool   `json:"ever_contacted" xorm:"ever_contacted"`
}

func convertPOToImFriend(po *imFriendPO) *imfriend.ImFriend {
	return &imfriend.ImFriend{
		Id: po.Id,

		UserUuid:      po.Fromuser,
		FriendUuid:    po.Touser,
		IsBlack:       po.Isblack == 1,
		Count:         po.Count,
		Status:        po.Status,
		Created:       po.Created,
		NextTime:      po.Nexttime,
		LastReadMsgId: po.LastReadMsgId,
		EverContacted: po.EverContacted,
	}
}
func convertImFriendToPO(friend *imfriend.ImFriend) *imFriendPO {
	isBlack := 0
	if friend.IsBlack {
		isBlack = 1
	}
	return &imFriendPO{
		Id:            friend.Id,
		Fromuser:      friend.UserUuid,
		Touser:        friend.FriendUuid,
		Isblack:       isBlack,
		Count:         friend.Count,
		Status:        friend.Status,
		Created:       friend.Created,
		Nexttime:      friend.NextTime,
		LastReadMsgId: friend.LastReadMsgId,
		EverContacted: friend.EverContacted,
	}
}

// endregion

// region implement

func NewManticoreImFriendRepo(engine *xorm.Engine) imfriend.ImFriendRepository {
	return &ManticoreImFriendRepo{
		ManticoreDB: engine,
	}
}

type ManticoreImFriendRepo struct {
	ManticoreDB *xorm.Engine
}

func (r ManticoreImFriendRepo) UpdateBlacklistStatus(friend *imfriend.ImFriend) error {
	var blacklist = 0
	if friend.IsBlack {
		blacklist = 1
	}
	sql := fmt.Sprintf(`update im_friend_list set isblack = %d where match('@fromuser %s @touser %s');`, blacklist, friend.UserUuid, friend.FriendUuid)
	_, err := r.ManticoreDB.Exec(sql)
	return err
}

func (r ManticoreImFriendRepo) AddImFriend(friend imfriend.ImFriend) error {
	sql := fmt.Sprintf(`insert into im_friend_list (fromuser,touser,isblack,count,status,created,nexttime) values ('%s','%s',%d,%d,%d,%d,%d)`, friend.UserUuid, friend.FriendUuid, 0, 2, 1, time.Now().UnixMilli(), 0)
	_, err := r.ManticoreDB.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func (r ManticoreImFriendRepo) UpdateContactStatus(uuid string, friendUuid string) error {
	sql := fmt.Sprintf(`update im_friend_list set ever_contacted = 1 where match('@fromuser %s @touser %s');`, uuid, friendUuid)
	_, err := r.ManticoreDB.Exec(sql)
	return err
}

func (r ManticoreImFriendRepo) GetFriendByUuid(uuid string, friendUuid string) (*imfriend.ImFriend, error) {
	var friendPO imFriendPO
	sql := fmt.Sprintf(`select * from im_friend_list where match('@fromuser %s @touser %s');`, uuid, friendUuid)
	has, err := r.ManticoreDB.SQL(sql).Get(&friendPO)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return convertPOToImFriend(&friendPO), nil
}

func (r ManticoreImFriendRepo) UpdateLastReadClientMsgId(uuid string, friendUuid string, lastReadMsgId string) error {
	sql := fmt.Sprintf(`update im_friend_list set last_read_msg_id = '%s' where match('@fromuser %s @touser %s');`, lastReadMsgId, uuid, friendUuid)
	_, err := r.ManticoreDB.Exec(sql)
	return err
}

func (r ManticoreImFriendRepo) ListImFriendByUuid(uuid string) (friends []*imfriend.ImFriend, err error) {
	var friendPOs []*imFriendPO
	sql := fmt.Sprintf(`select * from im_friend_list where match('@fromuser %s') order by isblack asc;`, uuid)
	err = r.ManticoreDB.SQL(sql).Find(&friendPOs)
	for _, friendPO := range friendPOs {
		friends = append(friends, convertPOToImFriend(friendPO))
	}
	return
}

// endregion
