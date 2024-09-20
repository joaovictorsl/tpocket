package main

import (
	"bufio"
	"context"

	"github.com/joaovictorsl/bencoding"
	"github.com/joaovictorsl/tpocket/discover"
	"github.com/joaovictorsl/tpocket/download"
	"github.com/joaovictorsl/tpocket/torrent"
)

func Download(ctx context.Context, torrentFile *bufio.Reader) error {
	td, err := decodeTorrentData(torrentFile)
	if err != nil {
		return err
	}

	discovery := discover.NewPeerDiscovery(context.Background(), td)
	downloader := download.NewDownloader(context.Background(), discovery.AddrCh(), td.Info())
	downloadMonitor := download.NewDownloadMonitor(discovery, downloader)

	downloadMonitor.Start()
	downloader.Start()
	discovery.Start()

	downloader.AssemblyFile()

	return nil
}

func decodeTorrentData(torrentFile *bufio.Reader) (torrent.ITorrentData, error) {
	data, err := bencoding.DecodeTo[map[string]interface{}](torrentFile)
	if err != nil {
		return nil, err
	}

	return torrent.TorrentDataFrom(data)
}
