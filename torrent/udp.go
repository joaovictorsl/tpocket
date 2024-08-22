package torrent

import (
	"encoding/binary"
	"math/rand"
	"net"
)

const (
	PROTOCOL_ID = 4497486125440
	CONNECT     = 0
	ANNOUNCE    = 1
)

type ConnectRequest struct {
	ProtocolId    uint64
	Action        uint32
	TransactionId uint32
}

func NewConnectRequest() *ConnectRequest {
	return &ConnectRequest{
		ProtocolId:    PROTOCOL_ID,
		Action:        CONNECT,
		TransactionId: rand.Uint32(),
	}
}

func (c *ConnectRequest) ToBytes() []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b, c.ProtocolId)
	binary.BigEndian.PutUint32(b[8:], c.Action)
	binary.BigEndian.PutUint32(b[12:], c.TransactionId)
	return b
}

type ConnectResponse struct {
	Action        uint32
	TransactionId uint32
	ConnectionId  uint64
}

func NewConnectResponse(b []byte) *ConnectResponse {
	return &ConnectResponse{
		Action:        CONNECT,
		TransactionId: binary.BigEndian.Uint32(b[4:8]),
		ConnectionId:  binary.BigEndian.Uint64(b[8:16]),
	}
}

type AnnounceRequest struct {
	ConnectionId  uint64
	Action        uint32
	TransactionId uint32
	InfoHash      []byte
	PeerId        []byte
	Downloaded    uint64
	Left          uint64
	Uploaded      uint64
	Event         uint32
	IPAddr        uint32
	Key           uint32
	NumWant       int32
	Port          uint16
}

func NewAnnounceRequest(connId uint64, infoHash []byte, length uint64, port uint16) *AnnounceRequest {
	return &AnnounceRequest{
		ConnectionId:  connId,
		Action:        ANNOUNCE,
		TransactionId: rand.Uint32(),
		InfoHash:      infoHash,
		PeerId:        []byte("00112233445566778899"),
		Downloaded:    0,
		Left:          length,
		Uploaded:      0,
		Event:         0,
		IPAddr:        0,
		Key:           0, // What is this?
		NumWant:       -1,
		Port:          port,
	}
}

func (a *AnnounceRequest) ToBytes() []byte {
	b := make([]byte, 0)
	b = binary.BigEndian.AppendUint64(b, a.ConnectionId)
	b = binary.BigEndian.AppendUint32(b, a.Action)
	b = binary.BigEndian.AppendUint32(b, a.TransactionId)
	b = append(b, a.InfoHash...)
	b = append(b, a.PeerId...)
	b = binary.BigEndian.AppendUint64(b, a.Downloaded)
	b = binary.BigEndian.AppendUint64(b, a.Left)
	b = binary.BigEndian.AppendUint64(b, a.Uploaded)
	b = binary.BigEndian.AppendUint32(b, a.Event)
	b = binary.BigEndian.AppendUint32(b, a.IPAddr)
	b = binary.BigEndian.AppendUint32(b, a.Key)
	b = binary.BigEndian.AppendUint32(b, uint32(a.NumWant))
	b = binary.BigEndian.AppendUint16(b, a.Port)
	return b
}

type AnnounceResponse struct {
	Action        uint32
	TransactionId uint32
	Interval      uint32
	Leechers      uint32
	Seeders       uint32
	Peers         []net.Addr
}

func NewAnnounceResponse(b []byte) *AnnounceResponse {
	res := &AnnounceResponse{
		Action:        ANNOUNCE,
		TransactionId: binary.BigEndian.Uint32(b[4:8]),
		Interval:      binary.BigEndian.Uint32(b[8:12]),
		Leechers:      binary.BigEndian.Uint32(b[12:16]),
		Seeders:       binary.BigEndian.Uint32(b[16:20]),
	}

	addrs := make([]net.Addr, res.Seeders)
	for i := uint32(0); i < res.Seeders; i++ {
		ipStart := 6*i + 20
		portStart := ipStart + 4
		addrs[i] = &net.TCPAddr{
			IP:   b[ipStart : ipStart+4],
			Port: int(binary.BigEndian.Uint16(b[portStart : portStart+2])),
		}
	}

	res.Peers = addrs

	return res
}
