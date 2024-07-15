package user

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

type ImFriend struct {
	Id         int64
	UserUuid   string
	FriendUuid string
	IsBlack    bool
	//=== 下面字段暂时不知道有什么用的
	Count    int
	Status   int //预留添加好友字段
	Created  int64
	NextTime int64
}

func (f ImFriend) SessionId() string {
	var idstr string
	rst := strings.Compare(f.FriendUuid, f.UserUuid)
	//小的放在前面就行了
	//sessionId.Touser<sessionId.Fromuser
	if rst == -1 {
		idstr = f.FriendUuid + f.UserUuid
	}
	if rst == 1 {
		idstr = f.UserUuid + f.FriendUuid
	}

	hash := md5.Sum([]byte(idstr))
	md5Str := hex.EncodeToString(hash[:])
	return md5Str
}

type ImFriendRepository interface {
	ListImFriendByUuid(uuid string) ([]*ImFriend, error)
}
