package persistence

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
)

const channelBufferSize = 100

type Reader struct {
	file       *os.File
	reader     *bufio.Reader
	gzipReader *gzip.Reader
}

func NewReader(filename string) (logReader *Reader, err error) {
	logReader = new(Reader)
	logReader.file, err = os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open log file: "+filename)
	}

	if filename[len(filename)-2:] == "gz" {
		logReader.gzipReader, err = gzip.NewReader(logReader.file)
		if err != nil {
			return nil, err
		}
		logReader.reader = bufio.NewReader(logReader.gzipReader)
	} else {
		logReader.reader = bufio.NewReader(logReader.file)
	}
	err = logReader.verifyLogFile()
	return
}

func (l *Reader) Close() error {
	if l.gzipReader != nil {
		err := l.gzipReader.Close()
		if err != nil {
			return err
		}
	}
	return l.file.Close()
}

func (l *Reader) HasMessage() bool {
	_, err := l.reader.Peek(1)
	return err != io.EOF
}

func (l *Reader) ReadMessage() (msg *Message, err error) {
	msg = new(Message)
	msg.Timestamp, err = l.readInt64()
	if err != nil {
		return
	}
	messageId, err := l.readInt32()
	msg.MessageType.Id = MessageId(messageId)
	if err != nil {
		return
	}
	length, err := l.readInt32()
	if err != nil {
		return
	}
	if length < 0 {
		err = errors.New(fmt.Sprintf("Length is invalid: %d", length))
		return
	}
	msg.Message, err = l.readBytes(int(length))
	if err != nil {
		return
	}
	return
}

func (l *Reader) SkipMessage() (bytesRead int, err error) {
	headerBytes := 12
	if n, err := l.reader.Discard(headerBytes); err != nil {
		if err == io.EOF {
			return 0, nil
		}
		log.Fatalf("Could not discard %v header bytes. Discarded: %v bytes: %v", headerBytes, n, err)
	}

	length, err := l.readInt32()
	if err != nil {
		return -1, err
	}

	if n, err := l.reader.Discard(int(length)); err != nil {
		log.Fatalf("Could not discard %v message bytes. Discarded: %v bytes: %v", length, n, err)
	}
	return headerBytes + 4 + int(length), nil
}

func (l *Reader) CreateChannel() (channel chan *Message) {
	channel = make(chan *Message, channelBufferSize)
	go l.readToChannel(channel)
	return
}

func (l *Reader) readToChannel(channel chan *Message) {
	for l.HasMessage() {
		msg, err := l.ReadMessage()
		if err != nil {
			break
		}

		channel <- msg
	}
	close(channel)
}

func (l *Reader) verifyLogFile() error {
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

func (l *Reader) readBytes(length int) ([]byte, error) {
	byteSlice := make([]byte, length)
	_, err := io.ReadAtLeast(l.reader, byteSlice, length)

	return byteSlice, err
}

func (l *Reader) readString(length int) (string, error) {
	s, err := l.readBytes(length)
	return string(s), err
}

func (l *Reader) readInt32() (ret int32, err error) {
	err = binary.Read(l.reader, binary.BigEndian, &ret)
	return
}

func (l *Reader) readInt64() (ret int64, err error) {
	err = binary.Read(l.reader, binary.BigEndian, &ret)
	return
}
