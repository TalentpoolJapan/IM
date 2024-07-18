package user

import "errors"

type ImFriendDomainService struct {
	imFriendRepo        ImFriendRepository
	imMessageRepository ImMessageRepository
}

func NewImFriendDomainService(imFriendRepo ImFriendRepository, imMessageRepository ImMessageRepository) ImFriendDomainService {
	return ImFriendDomainService{
		imFriendRepo:        imFriendRepo,
		imMessageRepository: imMessageRepository,
	}
}

func (s *ImFriendDomainService) SyncLastReadClientMsgId(userUuid string, friendUuid string, clientMsgId string) (bool, error) {
	imFriend, err := s.imFriendRepo.GetFriendByUuid(userUuid, friendUuid)
	if err != nil {
		return false, err
	}
	if imFriend == nil {
		return false, errors.New("friend not found")
	}
	sessionId := imFriend.SessionId()
	currentMessage, err := s.imMessageRepository.GetMessageByClientMsgId(sessionId, clientMsgId)
	if err != nil {
		return false, err
	}
	if currentMessage == nil {
		return false, errors.New("message not found")
	}

	var lastClientMsgIdCreateTime int64
	if imFriend.LastReadMsgId != "" {
		lastReadMsg, err := s.imMessageRepository.GetMessageByClientMsgId(sessionId, imFriend.LastReadMsgId)
		if err == nil && lastReadMsg != nil {
			lastClientMsgIdCreateTime = lastReadMsg.Created
		}
	}
	if currentMessage.Created > lastClientMsgIdCreateTime {
		s.imFriendRepo.UpdateLastReadClientMsgId(userUuid, friendUuid, clientMsgId)
	}
	return true, nil
}
