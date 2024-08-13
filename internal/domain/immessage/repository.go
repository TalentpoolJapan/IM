package immessage

type ImMessageRepository interface {
	LatestImMessageBySessionId(sessionId []string) ([]*ImMessage, error)
	ListMessageAfterMsgId(sessionId string, id int64) ([]*ImMessage, error)
	ListMessageAfterCreateTime(sessionId string, createTime int64) ([]*ImMessage, error)
	ListMessageBeforeCreateTime(sessionId string, createTime int64) ([]*ImMessage, error)
	ListMessageRecent(sessionId string, size int) ([]*ImMessage, error)
	ListMessageRecentBy(sessionId string, size int, userUuid string) ([]*ImMessage, error)
	GetMessageByClientMsgId(sessionId string, clientMsgId string) (*ImMessage, error)

	SaveImMessage(msg ImMessage) (int64, error)
}
