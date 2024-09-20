package torrent

import (
	"encoding/binary"
	"net"
)

type trackerResponse struct {
	announcer string
	interval  int
	peers     []net.Addr
}

func (tr trackerResponse) Announcer() string {
	return tr.announcer
}

func (tr trackerResponse) Interval() int {
	return tr.interval
}

func (tr trackerResponse) Peers() []net.Addr {
	return tr.peers
}

func NewTrackerResponse(announcer string, interval int, peers []net.Addr) ITrackerResponse {
	return trackerResponse{
		announcer: announcer,
		interval:  interval,
		peers:     peers,
	}
}

func TrackerResponseFrom(announce string, source map[string]interface{}) (ITrackerResponse, error) {
	interval, err := getField[int]("interval", source)
	if err != nil {
		return nil, err
	}

	strPeers, err := getField[string]("peers", source)
	if err != nil {
		return nil, err
	}

	peersBytes := []byte(strPeers)
	peers := make([]net.Addr, 0)

	for i := 0; i < len(peersBytes); i += 6 {
		currPeer := peersBytes[i : i+6]
		ip := currPeer[:4]
		portBytes := currPeer[4:]
		port := binary.BigEndian.Uint16(portBytes)

		peers = append(peers, &net.TCPAddr{
			IP:   ip,
			Port: int(port),
		})
	}

	return &trackerResponse{
		announcer: announce,
		interval:  interval,
		peers:     peers,
	}, nil
}
