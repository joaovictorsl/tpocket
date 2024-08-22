package messages

type UnchokeMessage struct {
}

func NewUnchokeMessage() *UnchokeMessage {
	return &UnchokeMessage{}
}

func FromBytesUnchokeMessage() *UnchokeMessage {
	return &UnchokeMessage{}
}

func (msg *UnchokeMessage) Type() int {
	return UNCHOKE
}

func (msg *UnchokeMessage) ToBytes() []byte {
	return []byte{0, 0, 0, 1, UNCHOKE}
}
