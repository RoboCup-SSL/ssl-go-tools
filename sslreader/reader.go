package sslreader

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"

	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type LogReader struct {
	file   *os.File
	reader *bufio.Reader
}

type LogMessage struct {
	timestamp   int64
	messageType int
	message     []byte
}

const (
	MESSAGE_BLANK           = 0 //(ignore message)
	MESSAGE_UNKNOWN         = 1 //(try to guess message type by parsing the data)
	MESSAGE_SSL_VISION_2010 = 2
	MESSAGE_SSL_REFBOX_2013 = 3
	MESSAGE_SSL_VISION_2014 = 4
)

func (l *LogReader) CreateVisionWrapperChannel(channel chan *SSL_WrapperPacket) {
	logMessageChannel := make(chan *LogMessage, 100)
	go l.CreateLogMessageChannel(logMessageChannel)

	for logMessage := range logMessageChannel {
		if logMessage.messageType == MESSAGE_SSL_VISION_2014 {
			visionMsg := parseVision2014(logMessage)
			channel <- visionMsg
		}
	}
	close(channel)
	return
}

func (l *LogReader) CreateVisionDetectionChannel(channel chan *SSL_DetectionFrame) {
	logMessageChannel := make(chan *LogMessage, 100)
	go l.CreateLogMessageChannel(logMessageChannel)

	for logMessage := range logMessageChannel {
		if logMessage.messageType == MESSAGE_SSL_VISION_2014 {
			visionMsg := parseVision2014(logMessage)
			if visionMsg.Detection != nil {
				channel <- visionMsg.Detection
			}
		}
	}
	close(channel)
	return
}

func (l *LogReader) CreateRefereeChannel(channel chan *SSL_Referee) {
	logMessageChannel := make(chan *LogMessage, 100)
	go l.CreateLogMessageChannel(logMessageChannel)

	for logMessage := range logMessageChannel {
		if logMessage.messageType == MESSAGE_SSL_REFBOX_2013 {
			refereeMsg := parseReferee2013(logMessage)
			channel <- refereeMsg
		}
	}
	close(channel)
	return
}

func (l *LogReader) CreateLogMessageChannel(channel chan *LogMessage) (err error) {
	for l.HasMessage() {
		msg, err := l.ReadMessage()
		if err != nil {
			break
		}

		channel <- msg
	}
	close(channel)
	return
}

func NewLogReader(filename string) (logReader *LogReader, err error) {
	logReader = new(LogReader)
	logReader.file, err = os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open log file: "+filename)
	}

	if filename[len(filename)-2:] == "gz" {
		gzipReader, err := gzip.NewReader(logReader.file)
		if err != nil {
			return nil, err
		}
		logReader.reader = bufio.NewReader(gzipReader)
	} else {
		logReader.reader = bufio.NewReader(logReader.file)
	}
	logReader.verifyLogFile()
	return
}

func (l *LogReader) Close() error {
	return l.file.Close()
}

func parseVision2014(msg *LogMessage) *SSL_WrapperPacket {
	packet := new(SSL_WrapperPacket)
	parseMessage(msg.message, packet)
	return packet
}

func parseReferee2013(msg *LogMessage) *SSL_Referee {
	packet := new(SSL_Referee)
	parseMessage(msg.message, packet)
	return packet
}

func parseMessage(data []byte, message proto.Message) error {

	if err := proto.Unmarshal(data, message); err != nil {
		return errors.Wrap(err, "unable to unmarshal data")
	}
	return nil
}

func (l *LogReader) HasMessage() bool {
	_, err := l.reader.Peek(1)
	return err != io.EOF
}

func (l *LogReader) verifyLogFile() error {
	header, err := l.readString(12)
	if err != nil {
		return err
	}
	if header != "SSL_LOG_FILE" {
		return errors.New("header validation failed. Found: " + header)
	}

	version, err := l.readInt32()
	if err != nil {
		return err
	}
	if version != 1 {
		return errors.New(fmt.Sprintf("unsupported version: %d", version))
	}
	return err
}

func (l *LogReader) ReadMessage() (msg *LogMessage, err error) {
	msg = new(LogMessage)
	msg.timestamp, err = l.readInt64()
	if err != nil {
		return
	}
	messageType, err := l.readInt32()
	msg.messageType = int(messageType)
	if err != nil {
		return
	}
	length, err := l.readInt32()
	if err != nil {
		return
	}
	msg.message, err = l.readBytes(int(length))
	if err != nil {
		return
	}
	return
}

func (l *LogReader) readBytes(length int) ([]byte, error) {
	byteSlice := make([]byte, length)
	_, err := io.ReadAtLeast(l.reader, byteSlice, length)

	return byteSlice, err
}

func (l *LogReader) readString(length int) (string, error) {
	s, err := l.readBytes(length)
	return string(s), err
}

func (l *LogReader) readInt32() (ret int32, err error) {
	err = binary.Read(l.reader, binary.BigEndian, &ret)
	return
}

func (l *LogReader) readInt64() (ret int64, err error) {
	err = binary.Read(l.reader, binary.BigEndian, &ret)
	return
}
