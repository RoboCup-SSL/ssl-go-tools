package sslconn

import (
	"bufio"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
)

// readDataLength reads the data length from message header
// The header is a 4 byte big endian uint32
func readDataLength(reader io.ByteReader) (length uint32, err error) {

	length64, err := binary.ReadUvarint(reader)
	length = uint32(length64)

	return
}

// SendMessage sends a protobuf message to the given connection
func SendMessage(conn net.Conn, message proto.Message) error {

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	size := uint64(len(data))
	data = append(proto.EncodeVarint(size), data...)

	if _, err = conn.Write(data); err != nil {
		return err
	}

	return nil
}

// ReceiveMessage reads a protobuf message and the preceding size from the given connection
func ReceiveMessage(conn net.Conn, message proto.Message) error {

	reader := bufio.NewReaderSize(conn, 1)
	dataLength, err := readDataLength(reader)
	if err != nil {
		return err
	}

	data := make([]byte, dataLength)
	if _, err = io.ReadFull(reader, data); err != nil {
		return err
	}

	if err = proto.Unmarshal(data, message); err != nil {
		return err
	}
	return nil
}
