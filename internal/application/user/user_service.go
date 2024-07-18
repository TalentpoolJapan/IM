package user

import (
	"imserver/internal/application"
	"imserver/internal/domain/user"
)

type AppService struct {
	imFriendRepo       user.ImFriendRepository
	userRepo           user.IUserRepository
	imMessageRepo      user.ImMessageRepository
	imFriendDomainServ user.ImFriendDomainService
}

func NewUserAppService(imFriendRepo user.ImFriendRepository,
	userRepo user.IUserRepository,
	imMessageRepo user.ImMessageRepository,
	imFriendDomainServ user.ImFriendDomainService) AppService {
	return AppService{
		imFriendRepo:       imFriendRepo,
		userRepo:           userRepo,
		imMessageRepo:      imMessageRepo,
		imFriendDomainServ: imFriendDomainServ,
	}
}

func (s AppService) GetMyContacts(qry *GetMyContactsQry) application.MultiResp[MyContacts] {
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
		message, existMessage := imMessageMap[friend.SessionId()]
		if existMessage && message != nil {
			contact.LatestMessage = message.Msg
			contact.LatestContactTime = message.Created
		}
		myContacts = append(myContacts, contact)
	}

	return application.MultiRespOf[MyContacts](myContacts, "get my friends success")
}

func (s AppService) GetUnreadMessageState(qry *UnreadMessageStateQry) application.SingleResp[UnreadMessageState] {

	// get contact
	friends, err := s.imFriendRepo.ListImFriendByUuid(qry.Uuid)
	if err != nil {
		return application.SingleRespFail[UnreadMessageState]("get my friends failed: " + err.Error())
	}

	if len(friends) == 0 {
		return application.SingleRespOk[UnreadMessageState]()
	}
	// search contact last read message
	var totalCount int
	unreadContacts := make([]UnreadContact, 0, len(friends))
	for _, friend := range friends {
		// get unread message count
		lastReadMsg, err := s.imMessageRepo.GetMessageByClientMsgId(friend.SessionId(), friend.LastReadMsgId)
		if err != nil || lastReadMsg == nil {
			continue
		}
		unreadMessages, err := s.imMessageRepo.ListMessageAfterCreateTime(friend.SessionId(), lastReadMsg.Created)
		if err != nil {
			continue
		}
		// filter self message
		var unreadMessageCount int
		for _, message := range unreadMessages {
			if message.FromUser != qry.Uuid {
				unreadMessageCount = unreadMessageCount + 1
			}
		}
		if unreadMessageCount == 0 {
			continue
		}
		unreadContacts = append(unreadContacts, UnreadContact{ContactUuid: friend.FriendUuid, Count: unreadMessageCount})
		totalCount += unreadMessageCount
	}

	return application.SingleRespOf[UnreadMessageState](
		UnreadMessageState{Total: totalCount, UnreadContacts: unreadContacts},
		"get unread message state success")
}

func (s AppService) SyncLastReadClientMsgId(cmd *SyncLastReadClientMsgIdCmd) application.SingleResp[any] {
	_, err := s.imFriendDomainServ.SyncLastReadClientMsgId(cmd.Uuid, cmd.FriendUuid, cmd.ClientMsgId)
	if err != nil {
		return application.SingleRespFail[any]("sync last read id failed: " + err.Error())
	}
	return application.SingleRespOk[any]()
}
