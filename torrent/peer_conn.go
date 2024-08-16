package torrent

import (
	"bytes"
	"encoding/binary"
	"net"
	"slices"
	"time"

	"github.com/joaovictorsl/mytorrent/torrent/messages"
)

type PeerConn struct {
	addr         net.Addr
	conn         net.Conn
	msgLengthBuf []byte
	payloadBuf   []byte
	pieceLen     uint32
}

func NewPeerConn(peerAddr net.Addr, pieceLen uint32) *PeerConn {
	return &PeerConn{
		addr:         peerAddr,
		msgLengthBuf: make([]byte, 4),
		payloadBuf:   make([]byte, 16*1024+9), // 9 bytes from msg_id, idx, and begin. 16 *1024 is max block size
		pieceLen:     pieceLen,
	}
}

func (pc *PeerConn) Handshake(infoHash []byte) error {
	msgBytes := bytes.NewBuffer(make([]byte, 0))
	msgBytes.Write([]byte{19})
	msgBytes.Write([]byte("BitTorrent protocol"))
	msgBytes.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	msgBytes.Write(infoHash)
	msgBytes.Write([]byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}) // TODO: Make this id configurable

	conn, err := net.DialTimeout("tcp", pc.addr.String(), 20*time.Second) // TODO: Make this timeout configurable
	if err != nil {
		return err
	}

	_, err = conn.Write(msgBytes.Bytes())
	if err != nil {
		conn.Close()
		return err
	}

	buf := make([]byte, 1024)

	_, err = conn.Read(buf)
	if err != nil {
		conn.Close()
		return err
	}

	pc.conn = conn

	return nil
}

func (pc *PeerConn) SendInterest() error {
	req := messages.NewInterestedMessage()
	_, err := pc.conn.Write(req.ToBytes())
	return err
}

func (pc *PeerConn) SendRequest(idx, begin, pieceLen uint32) error {
	blockLen := uint32(16 * 1024)
	length := pieceLen - begin
	if length > blockLen {
		length = blockLen
	}

	req := messages.NewRequestMessage(idx, begin, length)
	_, err := pc.conn.Write(req.ToBytes())
	return err
}

func (pc *PeerConn) HashMatches(piece []byte, hash []byte) bool {
	return slices.Compare(calcHash(piece), hash) == 0
}

func (pc *PeerConn) SendHave(idx uint32) error {
	msg := messages.NewHaveMessage(idx)
	_, err := pc.conn.Write(msg.ToBytes())
	return err
}

func (pc *PeerConn) ReadMessage() (messages.PeerMessage, error) {
	if _, err := pc.conn.Read(pc.msgLengthBuf); err != nil {
		return nil, err
	}

	msgBytes := int(binary.BigEndian.Uint32(pc.msgLengthBuf))
	bytesRead := 0

	for bytesRead != msgBytes {
		n, err := pc.conn.Read(pc.payloadBuf[bytesRead:])
		if err != nil {
			return nil, err
		}

		bytesRead += n
	}

	return messages.FromBytes(pc.payloadBuf)
}

func (pc *PeerConn) Close() error {
	return pc.conn.Close()
}
