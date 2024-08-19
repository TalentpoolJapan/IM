package user

type SyncLastReadClientMsgIdCmd struct {
	Uuid        string `json:"uuid"`
	FriendUuid  string `json:"friend_uuid"`
	ClientMsgId string `json:"client_msg_id"`
}

type AddImFriendCmd struct {
	Uuid       string `json:"uuid"`
	FriendUuid string `json:"friend_uuid"`
}

type AddSendMessageLimitMessageCmd struct {
	Uuid        string `json:"uuid"`
	FriendUuid  string `json:"friend_uuid"`
	Msg         string `json:"msg"`
	SystemMsgId string `json:"system_msg_id"`
}

type AddBlacklistMessageCmd struct {
	Uuid        string `json:"uuid"`
	FriendUuid  string `json:"friend_uuid"`
	Msg         string `json:"msg"`
	SystemMsgId string `json:"system_msg_id"`
}

type BlacklistFriendCmd struct {
	Uuid       string `json:"uuid"`
	FriendUuid string `json:"friend_uuid"`
}

type CancelBlacklistFriendCmd struct {
	Uuid       string `json:"uuid"`
	FriendUuid string `json:"friend_uuid"`
}
