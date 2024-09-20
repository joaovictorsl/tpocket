package messages

import (
	"bytes"
	"encoding/binary"
)

type BitfieldMessage struct {
	bitfield []byte
}

func NewBitfieldMessage(b []byte) *BitfieldMessage {
	return FromBytesBitfieldMessage(b)
}

func FromBytesBitfieldMessage(b []byte) *BitfieldMessage {
	return &BitfieldMessage{
		bitfield: b,
	}
}

func (msg *BitfieldMessage) Type() int {
	return BITFIELD
}

func (msg *BitfieldMessage) ToBytes() []byte {
	b := bytes.NewBuffer(make([]byte, 0))
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, uint32(len(msg.bitfield)+1))

	b.Write(tmp)
	b.WriteByte(BITFIELD)
	b.Write(msg.bitfield)

	return b.Bytes()
}

func (msg BitfieldMessage) Bitfield() []byte {
	return msg.bitfield
}
