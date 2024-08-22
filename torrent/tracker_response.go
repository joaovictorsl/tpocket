package torrent

import (
	"encoding/binary"
	"net"
)

type TrackerResponse struct {
	Interval int
	Peers    []net.Addr
}

func TrackerResponseFrom(source map[string]interface{}) (*TrackerResponse, error) {
	tr := &TrackerResponse{}

	interval, err := getField[int]("interval", source)
	if err != nil {
		return tr, err
	}

	strPeers, err := getField[string]("peers", source)
	if err != nil {
		return tr, err
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

	tr.Interval = interval
	tr.Peers = peers

	return tr, nil
}
