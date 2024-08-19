package user

import (
	"imserver/internal/application"
	"imserver/internal/domain/imfriend"
	"imserver/internal/domain/immessage"
	"imserver/internal/domain/user"
	"time"
)

type AppService struct {
	imFriendRepo  imfriend.ImFriendRepository
	userRepo      user.IUserRepository
	imMessageRepo immessage.ImMessageRepository
}

func NewUserAppService(imFriendRepo imfriend.ImFriendRepository,
	userRepo user.IUserRepository,
	imMessageRepo immessage.ImMessageRepository) AppService {
	return AppService{
		imFriendRepo:  imFriendRepo,
		userRepo:      userRepo,
		imMessageRepo: imMessageRepo,
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
	imMessageMap := make(map[string]*immessage.ImMessage)
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
	imFriend, err := s.imFriendRepo.GetFriendByUuid(cmd.Uuid, cmd.FriendUuid)
	if err != nil {
		return application.SingleRespFail[any]("sync last read id failed: " + err.Error())
	}
	if imFriend == nil {
		return application.SingleRespFail[any]("friend not found")
	}
	sessionId := imFriend.SessionId()
	currentMessage, err := s.imMessageRepo.GetMessageByClientMsgId(sessionId, cmd.ClientMsgId)
	if err != nil {
		return application.SingleRespFail[any]("sync last read id failed: " + err.Error())
	}
	if currentMessage == nil {
		return application.SingleRespFail[any]("message not found")
	}

	var lastClientMsgIdCreateTime int64
	if imFriend.LastReadMsgId != "" {
		lastReadMsg, err := s.imMessageRepo.GetMessageByClientMsgId(sessionId, imFriend.LastReadMsgId)
		if err == nil && lastReadMsg != nil {
			lastClientMsgIdCreateTime = lastReadMsg.Created
		}
	}
	if currentMessage.Created > lastClientMsgIdCreateTime {
		err := s.imFriendRepo.UpdateLastReadClientMsgId(cmd.Uuid, cmd.FriendUuid, cmd.ClientMsgId)
		if err != nil {
			return application.SingleRespFail[any]("sync last read id failed: " + err.Error())
		}
	}
	return application.SingleRespOk[any]()
}

func (s AppService) AddImFriend(cmd *AddImFriendCmd) application.SingleResp[any] {
	imFriend, err := s.imFriendRepo.GetFriendByUuid(cmd.Uuid, cmd.FriendUuid)
	if err != nil {
		return application.SingleRespFail[any]("add friend failed: " + err.Error())
	}
	if imFriend != nil {
		return application.SingleRespFail[any]("friend already exists")
	}
	imFriend = &imfriend.ImFriend{
		UserUuid:   cmd.Uuid,
		FriendUuid: cmd.FriendUuid,
		Count:      2,
		Created:    time.Now().UnixMicro(),
	}
	err = s.imFriendRepo.AddImFriend(*imFriend)
	if err != nil {
		return application.SingleRespFail[any]("add friend failed: " + err.Error())
	}
	return application.SingleRespOk[any]()
}

func (s AppService) ListImMessageRecent(qry *ListImMessageRecentQry) application.MultiResp[ImMessageDTO] {
	friend, err := s.imFriendRepo.GetFriendByUuid(qry.Uuid, qry.FriendUuid)
	if err != nil {
		return application.MultiRespFail[ImMessageDTO]("get friend failed: " + err.Error())
	}
	if friend == nil {
		return application.MultiRespFail[ImMessageDTO]("get friend failed: friend not exist")
	}

	imMessages, err := s.imMessageRepo.ListMessageRecent(friend.SessionId(), qry.Size)
	if err != nil {
		return application.MultiRespFail[ImMessageDTO]("get message failed: " + err.Error())
	}

	var messageDTOs []*ImMessageDTO
	for _, message := range imMessages {
		// 系统消息只有自己能看
		if message.MsgType == immessage.SendMessageLimit && message.FromUser != qry.Uuid {
			continue
		}
		messageDTOs = append(messageDTOs, buildImMessageDTO(message))
	}

	return application.MultiRespOf[ImMessageDTO](messageDTOs, "get message success")

}

func (s AppService) ListImMessageBeforeClientMsgId(qry *ListImMessageBeforeClientMsgQry) application.MultiResp[ImMessageDTO] {
	friend, err := s.imFriendRepo.GetFriendByUuid(qry.Uuid, qry.FriendUuid)
	if err != nil {
		return application.MultiRespFail[ImMessageDTO]("get friend failed: " + err.Error())
	}
	if friend == nil {
		return application.MultiRespFail[ImMessageDTO]("get friend failed: friend not exist")
	}

	curMessage, err := s.imMessageRepo.GetMessageByClientMsgId(friend.SessionId(), qry.ClientMsgId)
	if err != nil {
		return application.MultiRespFail[ImMessageDTO]("get message failed: " + err.Error())
	}
	if curMessage == nil {
		return application.MultiRespFail[ImMessageDTO]("get message failed: message not exist")
	}

	imMessages, err := s.imMessageRepo.ListMessageBeforeCreateTime(friend.SessionId(), curMessage.Created)
	if err != nil {
		return application.MultiRespFail[ImMessageDTO]("get message failed: " + err.Error())
	}

	var messageDTOs []*ImMessageDTO
	for _, message := range imMessages {
		// 系统消息只有自己能看
		if message.MsgType == immessage.SendMessageLimit && message.FromUser != qry.Uuid {
			continue
		}
		messageDTOs = append(messageDTOs, buildImMessageDTO(message))
	}

	return application.MultiRespOf[ImMessageDTO](messageDTOs, "get message success")

}

func (s AppService) ListImMessageAfterClientMsgId(qry *ListImMessageAfterClientMsgQry) application.MultiResp[ImMessageDTO] {
	friend, err := s.imFriendRepo.GetFriendByUuid(qry.Uuid, qry.FriendUuid)
	if err != nil {
		return application.MultiRespFail[ImMessageDTO]("get friend failed: " + err.Error())
	}
	if friend == nil {
		return application.MultiRespFail[ImMessageDTO]("get friend failed: friend not exist")
	}

	curMessage, err := s.imMessageRepo.GetMessageByClientMsgId(friend.SessionId(), qry.ClientMsgId)
	if err != nil {
		return application.MultiRespFail[ImMessageDTO]("get message failed: " + err.Error())
	}
	if curMessage == nil {
		return application.MultiRespFail[ImMessageDTO]("get message failed: message not exist")
	}

	imMessages, err := s.imMessageRepo.ListMessageAfterCreateTime(friend.SessionId(), curMessage.Created)
	if err != nil {
		return application.MultiRespFail[ImMessageDTO]("get message failed: " + err.Error())
	}

	var messageDTOs []*ImMessageDTO
	for _, message := range imMessages {
		// 系统消息只有自己能看
		if message.MsgType == immessage.SendMessageLimit && message.FromUser != qry.Uuid {
			continue
		}
		messageDTOs = append(messageDTOs, buildImMessageDTO(message))
	}

	return application.MultiRespOf[ImMessageDTO](messageDTOs, "get message success")
}

func (s AppService) AddSendMessageLimitMessage(cmd *AddSendMessageLimitMessageCmd) application.SingleResp[any] {
	imFriend, err := s.imFriendRepo.GetFriendByUuid(cmd.Uuid, cmd.FriendUuid)
	if err != nil {
		return application.SingleRespFail[any]("add system message failed: " + err.Error())
	}
	if imFriend == nil {
		return application.SingleRespFail[any]("add system message failed: friend not exist")
	}

	imMessage := &immessage.ImMessage{
		SessionId: imFriend.SessionId(),
		ToUser:    cmd.FriendUuid,
		FromUser:  cmd.Uuid,
		Msg:       cmd.Msg,
		MsgType:   immessage.SendMessageLimit,
		ToType:    0,
		FromType:  0,
		Created:   time.Now().UnixMicro(),
		MsgId:     cmd.SystemMsgId,
	}
	_, err = s.imMessageRepo.SaveImMessage(*imMessage)
	if err != nil {
		return application.SingleRespFail[any]("add system message failed: " + err.Error())
	}
	return application.SingleRespOk[any]()
}

func (s AppService) AddBlacklistMessage(cmd *AddBlacklistMessageCmd) application.SingleResp[any] {
	imFriend, err := s.imFriendRepo.GetFriendByUuid(cmd.Uuid, cmd.FriendUuid)
	if err != nil {
		return application.SingleRespFail[any]("add system message failed: " + err.Error())
	}
	if imFriend == nil {
		return application.SingleRespFail[any]("add system message failed: friend not exist")
	}

	imMessage := &immessage.ImMessage{
		SessionId: imFriend.SessionId(),
		ToUser:    cmd.FriendUuid,
		FromUser:  cmd.Uuid,
		Msg:       cmd.Msg,
		MsgType:   immessage.Blacklist,
		ToType:    0,
		FromType:  0,
		Created:   time.Now().UnixMicro(),
		MsgId:     cmd.SystemMsgId,
	}
	_, err = s.imMessageRepo.SaveImMessage(*imMessage)
	if err != nil {
		return application.SingleRespFail[any]("add system message failed: " + err.Error())
	}
	return application.SingleRespOk[any]()
}

func (s AppService) BlacklistFriend(cmd *BlacklistFriendCmd) application.SingleResp[any] {
	friend, err := s.imFriendRepo.GetFriendByUuid(cmd.Uuid, cmd.FriendUuid)
	if err != nil {
		return application.SingleRespFail[any]("blacklist friend failed: " + err.Error())
	}
	if friend == nil {
		return application.SingleRespFail[any]("blacklist friend failed: friend not exist")
	}
	if friend.IsBlack {
		return application.SingleRespOk[any]()
	}
	friend.BlacklistFriend()
	err = s.imFriendRepo.UpdateBlacklistStatus(friend)
	if err != nil {
		return application.SingleRespFail[any]("blacklist friend failed: " + err.Error())
	}
	return application.SingleRespOk[any]()
}

func (s AppService) CancelBlacklistFriend(cmd *CancelBlacklistFriendCmd) application.SingleResp[any] {
	friend, err := s.imFriendRepo.GetFriendByUuid(cmd.Uuid, cmd.FriendUuid)
	if err != nil {
		return application.SingleRespFail[any]("cancel blacklist friend failed: " + err.Error())
	}
	if friend == nil {
		return application.SingleRespFail[any]("cancel blacklist friend failed: friend not exist")
	}
	if !friend.IsBlack {
		return application.SingleRespOk[any]()
	}
	friend.CancelBlacklistFriend()
	err = s.imFriendRepo.UpdateBlacklistStatus(friend)
	if err != nil {
		return application.SingleRespFail[any]("cancel blacklist friend failed: " + err.Error())
	}
	return application.SingleRespOk[any]()
}

func buildImMessageDTO(message *immessage.ImMessage) *ImMessageDTO {
	return &ImMessageDTO{
		Id:        message.Id,
		Sessionid: message.SessionId,
		Touser:    message.ToUser,
		Fromuser:  message.FromUser,
		Msg:       message.Msg,
		Msgtype:   int(message.MsgType),
		Totype:    message.ToType,
		Fromtype:  message.FromType,
		Created:   message.Created,
		Msgid:     message.MsgId,
	}
}
