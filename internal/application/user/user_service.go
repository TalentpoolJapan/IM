package user

import (
	"imserver/internal/application"
	"imserver/internal/domain/user"
)

type AppService struct {
	imFriendRepo user.ImFriendRepository
}

func NewUserAppService(imFriendRepo user.ImFriendRepository) AppService {
	return AppService{
		imFriendRepo: imFriendRepo,
	}
}

type GetMyContactsQry struct {
	Uuid string `json:"uuid"`
}

type MyContacts struct {
	Uuid       string `json:"uuid"`
	FullNameEn string `json:"full_name_en"`
	FullNameJa string `json:"full_name_ja"`
	Avatar     string `json:"avatar"`
	IsBlack    int    `json:"is_black"`
}

func (s *AppService) GetMyContacts(qry *GetMyContactsQry) application.MultiResp[MyContacts] {
	// 1. get my friends
	friends, err := s.imFriendRepo.ListImFriendByUuid(qry.Uuid)
	if err != nil {
		return application.MultiRespFail[MyContacts]("get my friends failed: " + err.Error())
	}

	// 2. convert to response
	var myContacts []*MyContacts
	for _, friend := range friends {
		myContacts = append(myContacts, &MyContacts{
			Uuid:       friend.FriendUuid,
			FullNameEn: "full_name_en",
			FullNameJa: "full_name_ja",
			Avatar:     "avatar",
			IsBlack:    0,
		})
	}

	return application.MultiRespOf[MyContacts](myContacts, "get my friends success")
}
