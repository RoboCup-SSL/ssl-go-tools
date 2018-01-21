package protoconn

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"io"
	"net"
)

// read the data length from message header
// The header is a 4 byte big endian uint32
func readDataLength(conn net.Conn) (length uint32, err error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return 0, errors.Wrap(err, "unable to read data length")
	}
	length = binary.BigEndian.Uint32(header)
	return
}

// write the data length to the message header
// The header is a 4 byte big endian uint32
func writeDataLength(conn net.Conn, dataLength int) error {
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(dataLength))
	n, err := conn.Write(header)
	if err != nil {
		return errors.Wrap(err, "unable to write data length")
	}
	if n != 4 {
		return errors.New("invalid size written")
	}
	return nil
}

func SendMessage(conn net.Conn, message proto.Message) error {

	data, err := proto.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "marshaling error")
	}

	if err = writeDataLength(conn, len(data)); err != nil {
		return errors.Wrap(err, "unable to write data length")
	}
	if _, err = conn.Write(data); err != nil {
		return errors.Wrap(err, "unable to write data")
	}

	return nil
}

func ReceiveMessage(conn net.Conn, message proto.Message) error {

	dataLength, err := readDataLength(conn)
	if err != nil {
		return errors.Wrap(err, "unable to read data length")
	}

	data := make([]byte, dataLength)
	if _, err = io.ReadFull(conn, data); err != nil {
		return errors.Wrap(err, "unable to read data")
	}

	if err = proto.Unmarshal(data, message); err != nil {
		return errors.Wrap(err, "unable to unmarshal data")
	}
	return nil
}
