package sslconn

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"io"
	"net"
)

// readDataLength reads the data length from message header
// The header is a 4 byte big endian uint32
func readDataLength(conn net.Conn) (length uint32, err error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return 0, err
	}
	length = binary.BigEndian.Uint32(header)
	return
}

// writeDataLength writes the data length to the message header
// The header is a 4 byte big endian uint32
func writeDataLength(conn net.Conn, dataLength int) error {
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(dataLength))
	n, err := conn.Write(header)
	if err != nil {
		return err
	}
	if n != 4 {
		return errors.New("invalid size written")
	}
	return nil
}

// SendMessage sends a protobuf message to the given connection
func SendMessage(conn net.Conn, message proto.Message) error {

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	if err = writeDataLength(conn, len(data)); err != nil {
		return err
	}
	if _, err = conn.Write(data); err != nil {
		return err
	}

	return nil
}

// ReceiveMessage reads a protobuf message and the preceding size from the given connection
func ReceiveMessage(conn net.Conn, message proto.Message) error {

	dataLength, err := readDataLength(conn)
	if err != nil {
		return err
	}

	data := make([]byte, dataLength)
	if _, err = io.ReadFull(conn, data); err != nil {
		return err
	}

	if err = proto.Unmarshal(data, message); err != nil {
		return err
	}
	return nil
}
