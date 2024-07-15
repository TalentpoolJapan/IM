package user

import (
	"imserver/internal/application"
	"imserver/internal/domain/user"
)

type AppService struct {
	imFriendRepo  user.ImFriendRepository
	userRepo      user.IUserRepository
	imMessageRepo user.ImMessageRepository
}

func NewUserAppService(imFriendRepo user.ImFriendRepository,
	userRepo user.IUserRepository,
	repository user.ImMessageRepository) AppService {
	return AppService{
		imFriendRepo:  imFriendRepo,
		userRepo:      userRepo,
		imMessageRepo: repository,
	}
}

type GetMyContactsQry struct {
	Uuid string `json:"uuid"`
}

type MyContacts struct {
	Uuid              string `json:"uuid"`
	FullNameEn        string `json:"full_name_en"`
	FullNameJa        string `json:"full_name_ja"`
	Avatar            string `json:"avatar"`
	IsBlack           bool   `json:"is_black"`
	LatestMessage     string `json:"latest_message"`
	LatestContactTime int64  `json:"latest_contact_time"`
}

func (s *AppService) GetMyContacts(qry *GetMyContactsQry) application.MultiResp[MyContacts] {
	// get my friends
	friends, err := s.imFriendRepo.ListImFriendByUuid(qry.Uuid)
	if err != nil {
		return application.MultiRespFail[MyContacts]("get my friends failed: " + err.Error())
	}

	friendUuids := make([]string, len(friends))
	friendSessionIds := make([]string, len(friends))
	for i, friend := range friends {
		friendUuids[i] = friend.FriendUuid
		friendSessionIds[i] = friend.SessionId()
	}
	userBasicInfos, err := s.userRepo.ListUserBasicInfoByUuid(friendUuids)

	userInfoMap := make(map[string]*user.UserBasicInfo)
	for _, userInfo := range userBasicInfos {
		userInfoMap[userInfo.Uuid] = userInfo
	}

	imMessage, err := s.imMessageRepo.LatestImMessageBySessionId(friendSessionIds)
	imMessageMap := make(map[string]*user.ImMessage)
	for _, message := range imMessage {
		imMessageMap[message.SessionId] = message
	}

	// convert to response
	var myContacts []*MyContacts
	for _, friend := range friends {
		contact := &MyContacts{
			Uuid:              friend.FriendUuid,
			IsBlack:           friend.IsBlack,
			LatestContactTime: friend.Created,
		}
		userBasicInfo, existUser := userInfoMap[friend.FriendUuid]
		if existUser && userBasicInfo != nil {
			contact.FullNameEn = userBasicInfo.FullNameEn
			contact.FullNameJa = userBasicInfo.FullNameJa
			contact.Avatar = userBasicInfo.Avatar
		}
		message, existMessage := imMessageMap[friend.FriendUuid]
		if existMessage && message != nil {
			contact.LatestMessage = message.Msg
			contact.LatestContactTime = message.Created
		}
		myContacts = append(myContacts, contact)
	}

	return application.MultiRespOf[MyContacts](myContacts, "get my friends success")
}
