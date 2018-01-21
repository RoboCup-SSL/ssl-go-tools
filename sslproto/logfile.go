package sslproto

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type LogMessage struct {
	Timestamp   int64
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

func parseVision2014(msg *LogMessage) *SSL_WrapperPacket {
	packet := new(SSL_WrapperPacket)
	parseMessage(msg.Message, packet)
	return packet
}

func parseReferee2013(msg *LogMessage) *SSL_Referee {
	packet := new(SSL_Referee)
	parseMessage(msg.Message, packet)
	return packet
}

func parseMessage(data []byte, message proto.Message) error {

	if err := proto.Unmarshal(data, message); err != nil {
		return errors.Wrap(err, "unable to unmarshal data")
	}
	return nil
}

func (l *LogWriter) writeBytes(data []byte) error {
	_, err := l.writer.Write(data)
	return err
}

func (l *LogWriter) writeString(data string) error {
	_, err := l.writer.WriteString(data)
	return err
}

func (l *LogWriter) writeInt32(data int32) error {
	err := binary.Write(l.writer, binary.BigEndian, data)
	return err
}

func (l *LogWriter) writeInt64(data int64) error {
	err := binary.Write(l.writer, binary.BigEndian, data)
	return err
}
