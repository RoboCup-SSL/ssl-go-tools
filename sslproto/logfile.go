package sslproto

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type LogMessage struct {
	Timestamp   int64 // Receiver timestamp in ns
	MessageType int
	Message     []byte
}

const (
	MESSAGE_BLANK                = 0 //(ignore message)
	MESSAGE_UNKNOWN              = 1 //(try to guess message type by parsing the data)
	MESSAGE_SSL_VISION_2010      = 2
	MESSAGE_SSL_REFBOX_2013      = 3
	MESSAGE_SSL_VISION_2014      = 4
	MESSAGE_SSL_REFBOX_RCON_2018 = 5
)

func (m *LogMessage) ParseVisionWrapper() (*SSL_WrapperPacket, error) {
	packet := new(SSL_WrapperPacket)
	err := ParseMessage(m.Message, packet)
	return packet, err
}

func (m *LogMessage) ParseReferee() (*SSL_Referee, error) {
	packet := new(SSL_Referee)
	err := ParseMessage(m.Message, packet)
	return packet, err
}

func (m *LogMessage) ParseRefereeRemoteControlRequest() (*SSL_RefereeRemoteControlRequest, error) {
	packet := new(SSL_RefereeRemoteControlRequest)
	err := ParseMessage(m.Message, packet)
	return packet, err
}

func ParseMessage(data []byte, message proto.Message) error {

	if err := proto.Unmarshal(data, message); err != nil {
		return errors.Wrap(err, "unable to unmarshal data")
	}
	return nil
}
