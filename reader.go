package main

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"io"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type LogReader struct {
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

func main() {
	filename := "2017-07-27-erforce-src.log"
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var bufferedReader *bufio.Reader
	if filename[:len(filename)-2] == "gz" {
		gzipReader, err := gzip.NewReader(file)
		assertNoError(err)
		bufferedReader = bufio.NewReader(gzipReader)
	} else {
		bufferedReader = bufio.NewReader(file)
	}

	logReader := LogReader{bufferedReader}
	logReader.VerifyLogFile(bufferedReader)

	for logReader.HasMessage() {
		msg, err := logReader.ReadMessage()
		if err != nil {
			break
		}

		switch msg.messageType {
		case MESSAGE_SSL_REFBOX_2013:
			refereePkg := parseReferee2013(msg)
			if refereePkg.GetBlue().GetScore() > 0 {
				log.Print("blue yellow card")
			}
			if refereePkg.GetYellow().GetScore() > 0 {
				log.Print("blue yellow card")
			}
		case MESSAGE_SSL_VISION_2014:
		}
	}
}

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func parseVision2014(msg LogMessage) *SSL_WrapperPacket {
	packet := new(SSL_WrapperPacket)
	parseMessage(msg.message, packet)
	return packet
}

func parseReferee2013(msg LogMessage) *SSL_Referee {
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

func (l *LogReader) VerifyLogFile(reader *bufio.Reader) error {
	header, err := l.readString(12)
	assertNoError(err)
	if header != "SSL_LOG_FILE" {
		log.Fatal("header validation failed. Found: ", header)
	}

	version, err := l.readInt32()
	assertNoError(err)
	if version != 1 {
		log.Fatal("unsupported version: ", version)
	}
	return err
}

func (l *LogReader) ReadMessage() (msg LogMessage, err error) {
	msg.timestamp, err = l.readInt64()
	assertNoError(err)
	messageType, err := l.readInt32()
	msg.messageType = int(messageType)
	assertNoError(err)
	length, err := l.readInt32()
	assertNoError(err)
	msg.message, err = l.readBytes(int(length))
	assertNoError(err)
	return
}

func (r *LogReader) readBytes(length int) ([]byte, error) {
	byteSlice := make([]byte, length)
	_, err := io.ReadAtLeast(r.reader, byteSlice, length)

	return byteSlice, err
}

func (r *LogReader) readString(length int) (string, error) {
	s, err := r.readBytes(length)
	return string(s), err
}

func (r *LogReader) readInt32() (ret int32, err error) {
	err = binary.Read(r.reader, binary.BigEndian, &ret)
	return
}

func (r *LogReader) readInt64() (ret int64, err error) {
	err = binary.Read(r.reader, binary.BigEndian, &ret)
	return
}
