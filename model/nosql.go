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

func (n *Model) SetLastReadId(lastId SetLastReadId) error {
	relatedKey := fmt.Sprintf(`%s<-%s`, lastId.Touser, lastId.Fromuser)
	lastReadid := fmt.Sprintf("%d", lastId.Id)
	return n.NoSqlDB.Set(relatedKey, lastReadid)
}

func (n *Model) GetLastReadId(lastId GetLastReadId) (string, error) {
	relatedKey := fmt.Sprintf(`%s<-%s`, lastId.Touser, lastId.Fromuser)
	return n.NoSqlDB.Get(relatedKey)
}
