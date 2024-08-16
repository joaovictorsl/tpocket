package messages

import "encoding/binary"

type HaveMessage struct {
	Idx uint32
}

func NewHaveMessage(idx uint32) *HaveMessage {
	return &HaveMessage{
		Idx: idx,
	}
}

func FromBytesHaveMessage(b []byte) *HaveMessage {
	return &HaveMessage{
		Idx: binary.BigEndian.Uint32(b),
	}
}

func (msg *HaveMessage) Type() int {
	return HAVE
}

func (msg *HaveMessage) ToBytes() []byte {
	b := []byte{0, 0, 0, 5, HAVE}
	b = binary.BigEndian.AppendUint32(b, msg.Idx)
	return b
}
