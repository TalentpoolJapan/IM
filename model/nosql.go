package model

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

func (n *Model) SetMsgIdWithTTL(msgid string) (bool, error) {
	return n.NoSqlDB.SetIfNotExistWithTTL(msgid, "1", 5*time.Second)
}

type SetLastReadId struct {
	Id       int64
	Touser   string
	Fromuser string
}

type GetLastReadId struct {
	Touser   string
	Fromuser string
}

// 我<-你 你发给我的/我看过的/最后一条消息ID
func (n *Model) SetLastReadId(lastId SetLastReadId) error {
	relatedKey := fmt.Sprintf(`%s<-%s`, lastId.Touser, lastId.Fromuser)
	lastReadid := fmt.Sprintf("%d", lastId.Id)
	return n.NoSqlDB.Set(relatedKey, lastReadid)
}

// 如果没有说明没有发过消息
func (n *Model) GetLastReadId(lastId GetLastReadId) (string, error) {
	relatedKey := fmt.Sprintf(`%s<-%s`, lastId.Touser, lastId.Fromuser)
	val, err := n.NoSqlDB.Get(relatedKey)
	if err != nil {
		if err.Error() != "key not found in database" {
			return "", err
		}
	}
	return val, err
}

// 检查是否有好友记录
func (n *Model) CheckOrSetFriends(friend ImFreindList) (bool, error) {
	hash := md5.Sum([]byte(friend.Touser + friend.Fromuser))
	friendkey := hex.EncodeToString(hash[:])

	batch := n.NoSqlDB.GetBatch()
	val, err := batch.Get([]byte(friendkey))
	if err != nil {
		if err.Error() != "key not found in database" {
			batch.Rollback()
			return false, err
		}
	}

	if val != nil {
		batch.Rollback()
		return false, nil
	}

	err = batch.Put([]byte(friendkey), []byte(friend.Fromuser))
	if err != nil {
		batch.Rollback()
		return false, err
	}
	/////////////////////////////////////////////////////////////////
	hash2 := md5.Sum([]byte(friend.Fromuser + friend.Touser))
	if hash == hash2 {
		return false, errors.New("can not send msg to own")
	}
	friendkey = hex.EncodeToString(hash2[:])
	val, err = batch.Get([]byte(friendkey))
	if err != nil {
		if err.Error() != "key not found in database" {
			batch.Rollback()
			return false, err
		}
	}

	if val != nil {
		batch.Rollback()
		return false, nil
	}

	err = batch.Put([]byte(friendkey), []byte(friend.Touser))
	if err != nil {
		batch.Rollback()
		return false, err
	}
	///////////////////////////////////////////////////////////////
	session := n.FulltextDB.GetSession()
	defer session.Close()
	err = session.Begin()
	if err != nil {
		return false, err
	}
	sql := fmt.Sprintf(`insert into im_friend_list (touser,fromuser,isblack,status,created) values ('%s','%s',%d,%d,%d)`, friend.Touser, friend.Fromuser, 1, 2, time.Now().UnixNano())
	_, err = n.FulltextDB.Exec(sql)
	if err != nil {
		batch.Rollback()
		session.Rollback()
		return false, err
	}
	sql = fmt.Sprintf(`insert into im_friend_list (touser,fromuser,isblack,status,created) values ('%s','%s',%d,%d,%d)`, friend.Fromuser, friend.Touser, 1, 2, time.Now().UnixNano())
	_, err = n.FulltextDB.Exec(sql)
	if err != nil {
		batch.Rollback()
		session.Rollback()
		return false, err
	}
	//true
	// session, err := n.InsertFriends(friend)
	// if err != nil {
	// 	batch.Rollback()
	// }
	//一致性提交
	err = batch.Commit()
	if err != nil {
		session.Rollback()
		return false, err
	}
	err = session.Commit()
	if err != nil {
		batch.Rollback()
		return false, err
	}
	return true, nil
}
