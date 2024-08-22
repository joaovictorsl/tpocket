package discover

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joaovictorsl/bencoding"
	"github.com/joaovictorsl/tpocket/torrent"
)

type PeerDiscoverWorker struct {
	taskCh chan PeerDiscoverTask
	addrCh chan net.Addr
	buff   []byte
}

type PeerDiscoverTask struct {
	Announce string
	InfoHash []byte
	Length   uint64
}

func NewPeerDiscoverWorker(taskCh chan PeerDiscoverTask, addrCh chan net.Addr) *PeerDiscoverWorker {
	// I make this buffer big enough so I can get up to 5000 peers on a response
	// 20 bytes for seeders bla bla
	// 30_000 for peer addresses
	return &PeerDiscoverWorker{
		taskCh: taskCh,
		addrCh: addrCh,
		buff:   make([]byte, 30020),
	}
}

func (pd *PeerDiscoverWorker) Process() {
	var discoverFn func(string, []byte, uint64) (*torrent.TrackerResponse, error)
	for t := range pd.taskCh {
		if strings.Contains(t.Announce, "udp") {
			discoverFn = pd.discoverUDP
		} else {
			discoverFn = pd.discoverHTTP
		}

		tr, err := discoverFn(t.Announce, t.InfoHash, t.Length)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, p := range tr.Peers {
			pd.addrCh <- p
		}
	}
}

func (pd *PeerDiscoverWorker) discoverHTTP(announce string, infoHash []byte, length uint64) (*torrent.TrackerResponse, error) {
	params := url.Values{}
	params.Add("info_hash", string(infoHash))
	params.Add("peer_id", "00112233445566778899")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", string(length))
	params.Add("compact", "1")

	req, err := http.NewRequest(http.MethodGet, announce+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := bencoding.DecodeTo[map[string]interface{}](bufio.NewReader(res.Body))
	if err != nil {
		return nil, err
	}

	tr, err := torrent.TrackerResponseFrom(data)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

func (pd *PeerDiscoverWorker) discoverUDP(announce string, infoHash []byte, length uint64) (*torrent.TrackerResponse, error) {
	url := strings.Split(announce[6:], "/")[0] // Removes "url://"
	conn, err := net.Dial("udp", url)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	connectRes, err := pd.connectRequest(conn)
	if err != nil {
		return nil, err
	}

	tmp := strings.Split(conn.LocalAddr().String(), ":")
	port, err := strconv.ParseUint(tmp[len(tmp)-1], 10, 16)
	if err != nil {
		return nil, err
	}

	announceRes, err := pd.announceRequest(conn, connectRes.ConnectionId, infoHash, length, uint16(port))
	if err != nil {
		return nil, err
	}

	return &torrent.TrackerResponse{
		Interval: int(announceRes.Interval),
		Peers:    announceRes.Peers,
	}, nil
}

func (pd *PeerDiscoverWorker) connectRequest(conn net.Conn) (*ConnectResponse, error) {
	var res *ConnectResponse
	for i := 0; i < 2; i++ {
		timeout := time.Duration(15*math.Pow(2, float64(i))) * time.Second

		req := NewConnectRequest()
		_, err := conn.Write(req.ToBytes())
		if err != nil {
			return nil, err
		}

		conn.SetReadDeadline(time.Now().Add(timeout))
		n, err := conn.Read(pd.buff)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println("timeout")
				if i == 1 {
					return nil, fmt.Errorf("timed out for real connect")
				}
				continue
			}
			return nil, err
		} else if n < 16 {
			return nil, fmt.Errorf("bytes read on connect request should be >= 16")
		}

		res = NewConnectResponse(pd.buff[:n])
		if res.TransactionId != req.TransactionId {
			return nil, fmt.Errorf("res.TransactionId != req.TransactionId")
		} else if res.Action != CONNECT {
			return nil, fmt.Errorf("res.Action != CONNECT")
		}
	}

	return res, nil
}

func (pd *PeerDiscoverWorker) announceRequest(conn net.Conn, connId uint64, infoHash []byte, length uint64, port uint16) (*AnnounceResponse, error) {
	fmt.Println("announceRequest")
	var res *AnnounceResponse
	for i := 0; i < 2; i++ {
		timeout := time.Duration(15*math.Pow(2, float64(i))) * time.Second
		announceReq := NewAnnounceRequest(connId, infoHash, length, port)
		_, err := conn.Write(announceReq.ToBytes())
		if err != nil {
			return nil, err
		}

		conn.SetReadDeadline(time.Now().Add(timeout))
		n, err := conn.Read(pd.buff)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println("timeout")
				if i == 1 {
					return nil, fmt.Errorf("timedout for real")
				}
				continue
			}
			return nil, err
		} else if n < 20 {
			return nil, fmt.Errorf("bytes read on announce request should be >= 20")
		} else if pd.buff[3] == 3 {
			fmt.Println(string(pd.buff[8:n]))
			return nil, fmt.Errorf("announce response with error")
		}

		res = NewAnnounceResponse(pd.buff[:n])
	}

	return res, nil
}
