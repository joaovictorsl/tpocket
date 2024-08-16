package torrent

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"

	"github.com/joaovictorsl/bencoding"
)

type Client struct {
}

func (c *Client) Download(torrentFile *bufio.Reader) error {
	td, err := c.decodeTorrentData(torrentFile)
	if err != nil {
		return err
	}

	infoHash := calcHash([]byte(td.Info.Encode()))

	tr, err := c.discoverPeers(td.Announce, string(infoHash), fmt.Sprint(td.Info.Length))
	if err != nil {
		return err
	}

	pm := NewPieceManager(td.Info.Pieces)
	for _, peer := range tr.Peers {
		go NewDownloadWorker(peer, infoHash, uint32(td.Info.PieceLength), pm).Process()
	}

	<-pm.done

	return nil
}

func (c *Client) decodeTorrentData(torrentFile *bufio.Reader) (*TorrentData, error) {
	data, err := bencoding.DecodeTo[map[string]interface{}](torrentFile)
	if err != nil {
		return nil, err
	}

	return torrentDataFrom(data)
}

func (c *Client) discoverPeers(announce, infoHash, length string) (*TrackerResponse, error) {
	params := url.Values{}
	params.Add("info_hash", infoHash)
	params.Add("peer_id", "00112233445566778899")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", length)
	params.Add("compact", "1")
	params.Encode()

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

	tr, err := trackerResponseFrom(data)
	if err != nil {
		return nil, err
	}

	return tr, nil
}
