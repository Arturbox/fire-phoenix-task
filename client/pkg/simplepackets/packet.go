package simplepacket

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type MessagePacket struct {
	Length int16
	Body   struct {
		Address byte
		Command byte
		Message []byte
	}
}

func (msg *MessagePacket) ValidateMessage() error {
	if int(msg.Length)-2 != len(msg.Body.Message) {
		return fmt.Errorf("client message has invalid length: %d, expected: %d", len(msg.Body.Message)+2, msg.Length)
	}
	return nil
}

func (msg *MessagePacket) MarshalMessage(text string) ([]byte, error) {
	var err error
	msg.Body.Message = []byte(text)
	msg.Length = int16(len(msg.Body.Message) + 2)
	var writer bytes.Buffer
	enc := gob.NewEncoder(&writer)

	err = enc.Encode(msg)

	if err != nil {
		return nil, fmt.Errorf("error encoding message: %s", err)
	}

	return writer.Bytes(), nil
}
