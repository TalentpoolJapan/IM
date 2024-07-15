package user

type UserType int

const (
	JobSeeker  = UserType(1)
	Enterprise = UserType(2)
)

type UserBasicInfo struct {
	Uuid       string   `json:"uuid"`
	UserType   UserType `json:"user_type"`
	FullNameEn string   `json:"full_name_en"`
	FullNameJa string   `json:"full_name_ja"`
	Avatar     string   `json:"avatar"`
}

type IUserRepository interface {
	ListUserBasicInfoByUuid(uuids []string) ([]*UserBasicInfo, error)
}
