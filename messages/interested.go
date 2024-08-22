package messages

type InterestedMessage struct {
}

func NewInterestedMessage() *InterestedMessage {
	return &InterestedMessage{}
}

func FromBytesInterestedMessage() *InterestedMessage {
	return &InterestedMessage{}
}

func (msg *InterestedMessage) Type() int {
	return INTERESTED
}

func (msg *InterestedMessage) ToBytes() []byte {
	return []byte{0, 0, 0, 1, INTERESTED}
}
