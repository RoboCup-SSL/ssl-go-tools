package sslproto

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
)

type LogReader struct {
	file   *os.File
	reader *bufio.Reader
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
	msg.Timestamp, err = l.readInt64()
	if err != nil {
		return
	}
	messageType, err := l.readInt32()
	msg.MessageType = int(messageType)
	if err != nil {
		return
	}
	length, err := l.readInt32()
	if err != nil {
		return
	}
	msg.Message, err = l.readBytes(int(length))
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
