package persistence

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
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
)

func (m *Message) ParseVisionWrapper() (*sslproto.SSL_WrapperPacket, error) {
	packet := new(sslproto.SSL_WrapperPacket)
	err := ParseMessage(m.Message, packet)
	return packet, err
}

func (m *Message) ParseReferee() (*sslproto.Referee, error) {
	packet := new(sslproto.Referee)
	err := ParseMessage(m.Message, packet)
	return packet, err
}

func ParseMessage(data []byte, message proto.Message) error {

	if err := proto.Unmarshal(data, message); err != nil {
		return errors.Wrap(err, "unable to unmarshal data")
	}
	return nil
}

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
	default:
		return strconv.Itoa(int(m))
	}
}
