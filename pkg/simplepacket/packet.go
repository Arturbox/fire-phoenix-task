package simplepacket

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	PING                   = 0
	GET_CLIENT_LIST        = 1
	SEND_PRIVATE_MESSAGE   = 2
	SEND_BROADCAST_MESSAGE = 3
	DISCONNECT             = 4
)

const (
	OFFSET    = 6
	SERVER_ID = 0
)

type MessagePacket struct {
	Length uint16
	Body   struct {
		Address uint32
		Command uint16
		Message string
	}
}

func (msg *MessagePacket) Marshal() []byte {
	msg.Length = uint16(len(msg.Body.Message) + OFFSET)

	raw := make([]byte, 8)

	binary.BigEndian.PutUint16(raw[0:], msg.Length)
	binary.BigEndian.PutUint32(raw[2:], msg.Body.Address)
	binary.BigEndian.PutUint16(raw[6:], uint16(msg.Body.Command))

	byteString := []byte(msg.Body.Message)

	//slice message in case original message
	//being greater than maximum value of uint16
	raw = append(raw, byteString[0:uint16(len(msg.Body.Message))]...)

	return raw
}

func (msg *MessagePacket) Unmarshal(raw []byte, received int) error {
	if len(raw) < 2 {
		return fmt.Errorf("invalid message packet")
	}

	msg.Length = binary.BigEndian.Uint16(raw[0:2])

	if int(msg.Length)+2 != received {
		return fmt.Errorf("client message has invalid length: %d, expected: %d", len(raw), msg.Length+2)
	}

	msg.Body.Address = binary.BigEndian.Uint32(raw[2:6])
	msg.Body.Command = binary.BigEndian.Uint16(raw[6:8])
	msg.Body.Message = string(raw[8:received])

	return nil
}

func (mp *MessagePacket) GetMessage(conn net.Conn) (bool, error) {
	raw := make([]byte, 1024)
	n, err := bufio.NewReader(conn).Read(raw)

	if err == io.EOF || n == 0 {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("error reading request")
	}

	err = mp.Unmarshal(raw, n)

	if err != nil {
		return false, fmt.Errorf("error unmarshaling request: %s", err)
	}

	return true, nil
}
