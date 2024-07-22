package user

type GetMyContactsQry struct {
	Uuid string `json:"uuid"`
}

type UnreadMessageStateQry struct {
	Uuid string `json:"uuid"`
}

type ListImMessageRecentQry struct {
	Uuid       string `json:"uuid"`
	FriendUuid string `json:"friend_uuid"`
	Size       int    `json:"size"`
}

type ListImMessageBeforeClientMsgQry struct {
	Uuid        string `json:"uuid"`
	FriendUuid  string `json:"friend_uuid"`
	ClientMsgId string `json:"client_msg_id"`
}

type ListImMessageAfterClientMsgQry struct {
	Uuid        string `json:"uuid"`
	FriendUuid  string `json:"friend_uuid"`
	ClientMsgId string `json:"client_msg_id"`
}
