package model

import (
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
