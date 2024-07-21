package user

type MyContacts struct {
	Uuid              string `json:"uuid"`
	FullNameEn        string `json:"full_name_en"`
	FullNameJa        string `json:"full_name_ja"`
	Avatar            string `json:"avatar"`
	IsBlack           bool   `json:"is_black"`
	LatestMessage     string `json:"latest_message"`
	LatestContactTime int64  `json:"latest_contact_time"`
}

type UnreadMessageState struct {
	Total          int             `json:"total"`
	UnreadContacts []UnreadContact `json:"unread_contacts"`
}

type UnreadContact struct {
	ContactUuid string `json:"contact_uuid"`
	Count       int    `json:"count"`
}

type ImMessageDTO struct {
	Id        int64  `json:"id,omitempty"`
	Sessionid string `json:"sessionid,omitempty"`
	Touser    string `json:"touser,omitempty"`
	Fromuser  string `json:"fromuser,omitempty"`
	Msg       string `json:"msg,omitempty"`
	Msgtype   int    `json:"msgtype,omitempty"`
	Totype    int    `json:"totype,omitempty"`
	Fromtype  int    `json:"fromtype,omitempty"`
	Created   int64  `json:"created,omitempty"`
	Msgid     string `json:"msgid,omitempty"`
}
