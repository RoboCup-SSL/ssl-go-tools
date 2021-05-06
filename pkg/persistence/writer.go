package persistence

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"github.com/pkg/errors"
	"os"
)

const fileType = "SSL_LOG_FILE"
const indexedMarker = "INDEXED"
const HeaderSize = 16

type Writer struct {
	file       *os.File
	writer     *bufio.Writer
	gzipWriter *gzip.Writer
}

func NewWriter(filename string) (logWriter *Writer, err error) {
	logWriter = new(Writer)
	newLogfile := false
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		newLogfile = true
	}
	logWriter.file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		err = errors.Wrap(err, "Could not create log file: "+filename)
		return
	}

	if filename[len(filename)-2:] == "gz" {
		logWriter.gzipWriter = gzip.NewWriter(logWriter.file)
		logWriter.writer = bufio.NewWriter(logWriter.gzipWriter)
	} else {
		logWriter.writer = bufio.NewWriter(logWriter.file)
	}

	if newLogfile {
		err = logWriter.writeHeader()
		if err != nil {
			err = errors.Wrap(err, "Could not write header")
		}
	}

	return
}

func (l *Writer) writeHeader() error {
	_, err := l.writer.WriteString(fileType)
	if err != nil {
		return err
	}
	err = l.writeInt32(1)
	return err
}

func (l *Writer) Close() error {
	if l.writer == nil {
		// not open
		return nil
	}
	err := l.writer.Flush()
	if err != nil {
		return err
	}
	if l.gzipWriter != nil {
		err = l.gzipWriter.Close()
		if err != nil {
			return err
		}
	}
	return l.file.Close()
}

func (l *Writer) Write(msg *Message) (err error) {
	err = l.writeInt64(msg.Timestamp)
	if err != nil {
		return
	}
	err = l.writeInt32(int32(msg.MessageType.Id))
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

func (l *Writer) WriteIndex(offsets []int64) (err error) {
	payloadLen := len(offsets) * 8
	trailingSize := 8 + len(indexedMarker)
	msgLen := payloadLen + 16 + trailingSize
	timestamp := int64(0)

	err = l.writeInt64(timestamp)
	if err != nil {
		return
	}
	err = l.writeInt32(int32(MessageIndex2021))
	if err != nil {
		return
	}
	err = l.writeInt32(int32(payloadLen + trailingSize))
	if err != nil {
		return
	}
	err = l.writeInt64Array(offsets)
	if err != nil {
		return
	}
	// write backwards offset to last (this) message
	err = l.writeInt64(int64(msgLen))
	if err != nil {
		return
	}
	// mark the file as indexed
	err = l.writeString(indexedMarker)
	if err != nil {
		return
	}
	return
}

func (l *Writer) writeBytes(data []byte) error {
	_, err := l.writer.Write(data)
	return err
}

func (l *Writer) writeString(data string) error {
	_, err := l.writer.WriteString(data)
	return err
}

func (l *Writer) writeInt32(data int32) error {
	err := binary.Write(l.writer, binary.BigEndian, data)
	return err
}

func (l *Writer) writeInt64(data int64) error {
	err := binary.Write(l.writer, binary.BigEndian, data)
	return err
}

func (l *Writer) writeInt64Array(data []int64) error {
	err := binary.Write(l.writer, binary.BigEndian, data)
	return err
}
