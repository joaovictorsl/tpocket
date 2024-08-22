package main

import (
	"bufio"
	"context"
	"net"

	"github.com/joaovictorsl/bencoding"
	"github.com/joaovictorsl/gorkpool"
	"github.com/joaovictorsl/tpocket/discover"
	"github.com/joaovictorsl/tpocket/download"
	"github.com/joaovictorsl/tpocket/torrent"
)

func Download(ctx context.Context, torrentFile *bufio.Reader) error {
	td, err := decodeTorrentData(torrentFile)
	if err != nil {
		return err
	}

	length := td.Info.Length
	if length == 0 {
		for _, f := range td.Info.Files {
			length += f.Length
		}
	}

	peerAddrCh := make(chan net.Addr)
	discoverPool := gorkpool.NewBoundedGorkPool(ctx, 10, func(c chan discover.PeerDiscoverTask) gorkpool.BoundedGorkWorker[discover.PeerDiscoverTask] {
		return discover.NewPeerDiscoverWorker(c, peerAddrCh)
	})
	fileDownload := download.NewFileDownload(context.Background(), peerAddrCh, td.Info)
	fileDownload.Start()

	for _, announcer := range td.Announcers {
		discoverPool.AddTask(discover.PeerDiscoverTask{
			Announce: announcer,
			InfoHash: td.Info.Hash,
			Length:   uint64(length),
		})
	}

	fileDownload.AssemblyFile()

	return nil
}

func decodeTorrentData(torrentFile *bufio.Reader) (*torrent.TorrentData, error) {
	data, err := bencoding.DecodeTo[map[string]interface{}](torrentFile)
	if err != nil {
		return nil, err
	}

	return torrent.TorrentDataFrom(data)
}
