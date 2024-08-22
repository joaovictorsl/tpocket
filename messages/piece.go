package messages

import (
	"bytes"
	"encoding/binary"
)

type PieceMessage struct {
	Idx   uint32
	Begin uint32
	Block []byte
}

func NewPieceMessage(idx, begin uint32, block []byte) *PieceMessage {
	return &PieceMessage{
		Idx:   idx,
		Begin: begin,
		Block: block,
	}
}

func FromBytesPieceMessage(b []byte) *PieceMessage {
	return &PieceMessage{
		Idx:   binary.BigEndian.Uint32(b[0:4]),
		Begin: binary.BigEndian.Uint32(b[4:8]),
		Block: b[8:],
	}
}

func (msg *PieceMessage) Type() int {
	return PIECE
}

func (msg *PieceMessage) ToBytes() []byte {
	b := bytes.NewBuffer(make([]byte, 0))
	a := make([]byte, 4)

	binary.BigEndian.PutUint32(a, uint32(9+len(msg.Block))) // Message Length

	b.WriteByte(PIECE) // Message id

	binary.BigEndian.PutUint32(a, msg.Idx) // Piece index
	b.Write(a)

	binary.BigEndian.PutUint32(a, msg.Begin) // Block begin
	b.Write(a)

	b.Write(msg.Block) // Block data

	return b.Bytes()
}
