package user

type ImMessage struct {
	Id        int64  `json:"id,omitempty"`
	SessionId string `json:"sessionid,omitempty"`
	ToUser    string `json:"touser,omitempty"`
	FromUser  string `json:"fromuser,omitempty"`
	Msg       string `json:"msg,omitempty"`
	MsgType   int    `json:"msgtype,omitempty"`
	ToType    int    `json:"totype,omitempty"`
	FromType  int    `json:"fromtype,omitempty"`
	Created   int64  `json:"created,omitempty"`
	MsgId     string `json:"msgid,omitempty"`
}

type ImMessageRepository interface {
	LatestImMessageBySessionId(sessionId []string) ([]*ImMessage, error)
	ListMessageAfterMsgId(sessionId string, id int64) ([]ImMessage, error)
	ListMessageAfterCreateTime(sessionId string, createTime int64) ([]ImMessage, error)
	GetMessageByClientMsgId(sessionId string, clientMsgId string) (*ImMessage, error)
}
