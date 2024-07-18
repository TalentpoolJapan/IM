package user

type SyncLastReadClientMsgIdCmd struct {
	Uuid        string `json:"uuid"`
	FriendUuid  string `json:"friend_uuid"`
	ClientMsgId string `json:"client_msg_id"`
}
