package user

type GetMyContactsQry struct {
	Uuid string `json:"uuid"`
}

type UnreadMessageStateQry struct {
	Uuid string `json:"uuid"`
}
