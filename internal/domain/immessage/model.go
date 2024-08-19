package immessage

type MessageType int

const (
	Text             = 1 << iota
	SendMessageLimit = 2
	Blacklist        = 3
)

type ImMessage struct {
	Id        int64       `json:"id,omitempty"`
	SessionId string      `json:"sessionid,omitempty"`
	ToUser    string      `json:"touser,omitempty"`
	FromUser  string      `json:"fromuser,omitempty"`
	Msg       string      `json:"msg,omitempty"`
	MsgType   MessageType `json:"msgtype,omitempty"`
	ToType    int         `json:"totype,omitempty"`
	FromType  int         `json:"fromtype,omitempty"`
	Created   int64       `json:"created,omitempty"`
	MsgId     string      `json:"msgid,omitempty"`
}
