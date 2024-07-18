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
