package discover

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/joaovictorsl/bencoding"
	"github.com/joaovictorsl/tpocket/torrent"
)

type PeerDiscoverWorker struct {
	announce string
	infoHash []byte
	length   uint64

	inputCh  chan struct{}
	outputCh chan net.Addr
	buff     []byte

	log     *log.Logger
	logFile *os.File
}

type PeerDiscoverTask struct {
	Announce string
	InfoHash []byte
	Length   uint64
}

func NewPeerDiscoverWorker(announce string, infoHash []byte, length uint64, inputCh chan struct{}, outputCh chan net.Addr) *PeerDiscoverWorker {
	// I make this buffer big enough so I can get up to 5000 peers on a response
	// 20 bytes for seeders bla bla
	// 30_000 for peer addresses
	pd := &PeerDiscoverWorker{
		announce: announce,
		infoHash: infoHash,
		length:   length,
		inputCh:  inputCh,
		outputCh: outputCh,
		buff:     make([]byte, 30020),
	}

	return pd
}

func (w *PeerDiscoverWorker) startLogger() error {
	// TODO: Create log folder
	f, err := os.OpenFile("./log/announce_"+strings.Split(w.announce, "/")[2]+"_log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.log = log.New(f, "[PeerDiscoverWorker] ", log.Flags())
	w.logFile = f

	return nil
}

func (w *PeerDiscoverWorker) ID() string {
	return w.announce
}

func (w *PeerDiscoverWorker) SignalRemoval() {
	// We're not using this right now
}

func (w *PeerDiscoverWorker) Process() {
	if err := w.startLogger(); err != nil {
		panic(err)
	}
	defer w.logFile.Close()

	failures := 0

	t := time.NewTimer(-1)
	for failures < 5 {
		select {
		case <-t.C:
		case _, ok := <-w.inputCh:
			t.Stop()
			if !ok {
				return
			}
		}

		w.log.Println("Getting peers")

		tr, err := w.discoverPeer()
		if err != nil {
			w.log.Println(err, reflect.TypeOf(err))
			t.Reset(-1)
			failures++
			continue
		}

		for _, p := range tr.Peers() {
			w.outputCh <- p
		}

		t.Reset(time.Duration(tr.Interval()) * time.Second)
		w.log.Println("Rerequest", time.Now().Add(time.Duration(tr.Interval())*time.Second))
	}
}

func (w *PeerDiscoverWorker) discoverPeer() (torrent.ITrackerResponse, error) {
	if strings.Contains(w.announce, "udp") {
		return w.discoverUDP()
	}

	return w.discoverHTTP()
}

func (w *PeerDiscoverWorker) discoverHTTP() (torrent.ITrackerResponse, error) {
	params := url.Values{}
	params.Add("info_hash", string(w.infoHash))
	params.Add("peer_id", "00112233445566778899")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", string(w.length))
	params.Add("compact", "1")

	req, err := http.NewRequest(http.MethodGet, w.announce+"?"+params.Encode(), nil)
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

	tr, err := torrent.TrackerResponseFrom(w.announce, data)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

func (w *PeerDiscoverWorker) discoverUDP() (torrent.ITrackerResponse, error) {
	url := strings.Split(w.announce[6:], "/")[0] // Removes "url://"
	conn, err := net.Dial("udp", url)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	connectRes, err := w.connectRequest(conn)
	if err != nil {
		return nil, err
	}

	tmp := strings.Split(conn.LocalAddr().String(), ":")
	port, err := strconv.ParseUint(tmp[len(tmp)-1], 10, 16)
	if err != nil {
		return nil, err
	}

	announceRes, err := w.announceRequest(conn, connectRes.ConnectionId, w.infoHash, w.length, uint16(port))
	if err != nil {
		return nil, err
	}

	return torrent.NewTrackerResponse(w.announce, int(announceRes.Interval), announceRes.Peers), nil
}

func (w *PeerDiscoverWorker) connectRequest(conn net.Conn) (*ConnectResponse, error) {
	var res *ConnectResponse
	for i := 0; i < 2; i++ {
		timeout := time.Duration(15*math.Pow(2, float64(i))) * time.Second

		req := NewConnectRequest()
		_, err := conn.Write(req.ToBytes())
		if err != nil {
			return nil, err
		}

		conn.SetReadDeadline(time.Now().Add(timeout))
		n, err := conn.Read(w.buff)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				w.log.Println("timeout")
				if i == 1 {
					return nil, fmt.Errorf("timed out for real connect")
				}
				continue
			}
			return nil, err
		} else if n < 16 {
			return nil, fmt.Errorf("bytes read on connect request should be >= 16")
		}

		res = NewConnectResponse(w.buff[:n])
		if res.TransactionId != req.TransactionId {
			return nil, fmt.Errorf("res.TransactionId != req.TransactionId")
		} else if res.Action != CONNECT {
			return nil, fmt.Errorf("res.Action != CONNECT")
		}
		break
	}

	return res, nil
}

func (w *PeerDiscoverWorker) announceRequest(conn net.Conn, connId uint64, infoHash []byte, length uint64, port uint16) (*AnnounceResponse, error) {
	w.log.Println("announceRequest")
	var res *AnnounceResponse
	for i := 0; i < 2; i++ {
		timeout := time.Duration(15*math.Pow(2, float64(i))) * time.Second
		announceReq := NewAnnounceRequest(connId, infoHash, length, port)
		_, err := conn.Write(announceReq.ToBytes())
		if err != nil {
			return nil, err
		}

		conn.SetReadDeadline(time.Now().Add(timeout))
		n, err := conn.Read(w.buff)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				w.log.Println("timeout")
				if i == 1 {
					return nil, fmt.Errorf("timedout for real")
				}
				continue
			}
			return nil, err
		} else if n < 20 {
			return nil, fmt.Errorf("bytes read on announce request should be >= 20")
		} else if w.buff[3] == 3 {
			w.log.Println(string(w.buff[8:n]))
			return nil, fmt.Errorf("announce response with error")
		}

		w.log.Println("n", n)
		w.log.Println("pd.buff", w.buff[:n])
		res = NewAnnounceResponse(w.buff[:n])
		break
	}

	return res, nil
}
