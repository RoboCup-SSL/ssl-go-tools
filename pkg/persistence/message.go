package persistence

import (
	"strconv"
)

type MessageId int

type MessageType struct {
	Id   MessageId
	Name string
}

type Message struct {
	Timestamp   int64 // Receiver timestamp in ns
	MessageType MessageType
	Message     []byte
}

const (
	MessageBlank                MessageId = 0 //(ignore message)
	MessageUnknown              MessageId = 1 //(try to guess message type by parsing the data)
	MessageSslVision2010        MessageId = 2
	MessageSslRefbox2013        MessageId = 3
	MessageSslVision2014        MessageId = 4
	MessageSslVisionTracker2020 MessageId = 5
	MessageIndex2021            MessageId = 6
)

func (m MessageId) String() string {
	switch m {
	case MessageBlank:
		return "MessageBlank"
	case MessageUnknown:
		return "MessageUnknown"
	case MessageSslVision2010:
		return "MessageSslVision2010"
	case MessageSslRefbox2013:
		return "MessageSslRefbox2013"
	case MessageSslVision2014:
		return "MessageSslVision2014"
	case MessageSslVisionTracker2020:
		return "MessageSslVisionTracker2020"
	case MessageIndex2021:
		return "MessageIndex2021"
	default:
		return strconv.Itoa(int(m))
	}
}
