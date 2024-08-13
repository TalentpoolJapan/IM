package imfriend

type ImFriendRepository interface {
	GetFriendByUuid(uuid string, friendUuid string) (*ImFriend, error)
	ListImFriendByUuid(uuid string) ([]*ImFriend, error)
	AddImFriend(friend ImFriend) error
	UpdateLastReadClientMsgId(uuid string, friendUuid string, lastReadMsgId string) error
	UpdateContactStatus(uuid string, friendUuid string) error
}
