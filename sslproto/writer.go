package sslproto

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"github.com/pkg/errors"
	"os"
)

type LogWriter struct {
	file   *os.File
	writer *bufio.Writer
}

func NewLogWriter(filename string) (logWriter *LogWriter, err error) {
	logWriter = new(LogWriter)
	logWriter.file, err = os.Create(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create log file: "+filename)
	}

	if filename[len(filename)-2:] == "gz" {
		gzipWriter := gzip.NewWriter(logWriter.file)
		logWriter.writer = bufio.NewWriter(gzipWriter)
	} else {
		logWriter.writer = bufio.NewWriter(logWriter.file)
	}
	logWriter.writeHeader()
	return
}

func (l *LogWriter) writeHeader() error {
	_, err := l.writer.WriteString("SSL_LOG_FILE")
	if err != nil {
		return err
	}
	err = l.writeInt32(1)
	return err
}

func (l *LogWriter) Close() error {
	err := l.writer.Flush()
	if err != nil {
		return err
	}
	return l.file.Close()
}

func (l *LogWriter) WriteMessage(msg *LogMessage) (err error) {
	err = l.writeInt64(msg.Timestamp)
	if err != nil {
		return
	}
	err = l.writeInt32(int32(msg.MessageType))
	if err != nil {
		return
	}
	err = l.writeInt32(int32(len(msg.Message)))
	if err != nil {
		return
	}
	err = l.writeBytes(msg.Message)
	if err != nil {
		return
	}
	return
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
