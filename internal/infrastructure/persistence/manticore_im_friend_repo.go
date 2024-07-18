package persistence

import (
	"fmt"
	"imserver/internal/domain/user"
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
}

func convertPOToImFriend(po *imFriendPO) *user.ImFriend {
	return &user.ImFriend{
		Id:            po.Id,
		UserUuid:      po.Touser,
		FriendUuid:    po.Fromuser,
		IsBlack:       po.Isblack == 1,
		Count:         po.Count,
		Status:        po.Status,
		Created:       po.Created,
		NextTime:      po.Nexttime,
		LastReadMsgId: po.LastReadMsgId,
	}
}
func convertImFriendToPO(friend *user.ImFriend) *imFriendPO {
	isBlack := 0
	if friend.IsBlack {
		isBlack = 1
	}
	return &imFriendPO{
		Id:            friend.Id,
		Touser:        friend.UserUuid,
		Fromuser:      friend.FriendUuid,
		Isblack:       isBlack,
		Count:         friend.Count,
		Status:        friend.Status,
		Created:       friend.Created,
		Nexttime:      friend.NextTime,
		LastReadMsgId: friend.LastReadMsgId,
	}
}

// endregion

// region implement

func NewManticoreImFriendRepo(engine *xorm.Engine) user.ImFriendRepository {
	return &ManticoreImFriendRepo{
		ManticoreDB: engine,
	}
}

type ManticoreImFriendRepo struct {
	ManticoreDB *xorm.Engine
}

func (r ManticoreImFriendRepo) GetFriendByUuid(uuid string, friendUuid string) (*user.ImFriend, error) {
	var friendPO imFriendPO
	sql := fmt.Sprintf(`select * from im_friend_list where match('@fromuser %s @touser %s');`, friendUuid, uuid)
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
	sql := fmt.Sprintf(`update im_friend_list set last_read_msg_id = '%s' where match('@fromuser %s @touser %s');`, lastReadMsgId, friendUuid, uuid)
	_, err := r.ManticoreDB.Exec(sql)
	return err
}

func (r ManticoreImFriendRepo) ListImFriendByUuid(uuid string) (friends []*user.ImFriend, err error) {
	var friendPOs []*imFriendPO
	sql := fmt.Sprintf(`select * from im_friend_list where match('@touser %s') order by isblack asc;`, uuid)
	err = r.ManticoreDB.SQL(sql).Find(&friendPOs)
	for _, friendPO := range friendPOs {
		friends = append(friends, convertPOToImFriend(friendPO))
	}
	return
}

// endregion
