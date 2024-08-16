package messages

import (
	"fmt"
)

type PeerMessage interface {
	ToBytes() []byte
	Type() int
}

func FromBytes(b []byte) (PeerMessage, error) {
	var msg PeerMessage
	msgId := b[0]

	switch msgId {
	case CHOKE:
		msg = FromBytesChokeMessage()
	case UNCHOKE:
		msg = FromBytesUnchokeMessage()
	case INTERESTED:
		msg = FromBytesInterestedMessage()
	case BITFIELD:
		msg = FromBytesBitfieldMessage(b[1:])
	case REQUEST:
		msg = FromBytesRequestMessage(b[1:])
	case PIECE:
		msg = FromBytesPieceMessage(b[1:])
	default:
		return nil, fmt.Errorf("id not implemented: %d", msgId)
	}

	return msg, nil
}
