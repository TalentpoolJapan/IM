package user

type ImFriendRepository interface {
	ListImFriendByUuid(uuid string) ([]*ImFriend, error)
}
