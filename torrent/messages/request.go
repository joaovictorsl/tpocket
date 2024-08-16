package messages

import (
	"bytes"
	"encoding/binary"
)

type RequestMessage struct {
	Idx    uint32
	Begin  uint32
	Length uint32
}

func NewRequestMessage(idx, begin, length uint32) *RequestMessage {
	return &RequestMessage{
		Idx:    idx,
		Begin:  begin,
		Length: length,
	}
}

func FromBytesRequestMessage(b []byte) *RequestMessage {
	return &RequestMessage{
		Idx:    binary.BigEndian.Uint32(b[0:4]),
		Begin:  binary.BigEndian.Uint32(b[4:8]),
		Length: binary.BigEndian.Uint32(b[8:12]),
	}
}

func (msg *RequestMessage) Type() int {
	return REQUEST
}

func (msg *RequestMessage) ToBytes() []byte {
	b := bytes.NewBuffer(make([]byte, 0))
	tmp := make([]byte, 4)

	b.Write([]byte{0, 0, 0, 13}) // Payload length
	b.WriteByte(REQUEST)         // Message id

	binary.BigEndian.PutUint32(tmp, msg.Idx) // Piece index
	b.Write(tmp)

	binary.BigEndian.PutUint32(tmp, msg.Begin) // Block begin
	b.Write(tmp)

	binary.BigEndian.PutUint32(tmp, msg.Length) // Block length
	b.Write(tmp)

	return b.Bytes()
}
