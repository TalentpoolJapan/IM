package user

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
