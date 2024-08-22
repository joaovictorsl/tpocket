package torrent

import (
	"bufio"
	"context"
	"fmt"
	"net"

	"github.com/joaovictorsl/bencoding"
	"github.com/joaovictorsl/gorkpool"
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

	peerAddrCh := make(chan []net.Addr)
	discoverPool := gorkpool.NewBoundedGorkPool(ctx, 10, func(c chan PeerDiscoverTask) gorkpool.BoundedGorkWorker[PeerDiscoverTask] {
		return NewPeerDiscoverWorker(c, peerAddrCh)
	})
	dm := NewDownloadManager(context.Background(), peerAddrCh, td)

	for _, announcer := range td.Announcers {
		discoverPool.AddTask(PeerDiscoverTask{
			Announce: announcer,
			InfoHash: td.Info.Hash,
			Length:   uint64(length),
		})
	}

	fmt.Println("tr", len(tr.Peers))
	panic("wehat going on")

	pm := NewPieceManager(td.Info.Pieces)
	for _, peer := range tr.Peers {
		go NewDownloadWorker(peer, td.Info.Hash, uint32(td.Info.PieceLength), pm).Process()
	}

	<-pm.done

	return nil
}

func decodeTorrentData(torrentFile *bufio.Reader) (*TorrentData, error) {
	data, err := bencoding.DecodeTo[map[string]interface{}](torrentFile)
	if err != nil {
		return nil, err
	}

	return torrentDataFrom(data)
}
