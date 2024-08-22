package messages

type ChokeMessage struct {
}

func NewChokeMessage() *ChokeMessage {
	return &ChokeMessage{}
}

func FromBytesChokeMessage() *ChokeMessage {
	return &ChokeMessage{}
}

func (msg *ChokeMessage) Type() int {
	return CHOKE
}

func (msg *ChokeMessage) ToBytes() []byte {
	return []byte{0, 0, 0, 1, CHOKE}
}
