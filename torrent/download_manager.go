package torrent

import (
	"context"
	"net"

	"github.com/joaovictorsl/gorkpool"
)

type DownloadManager struct {
	downloadPool *gorkpool.UnboundedGorkPool[DownloadTask]
	pm           *PieceManager
}

func NewDownloadManager(ctx context.Context, peerAddrCh chan net.Addr, td *TorrentInfo) *DownloadManager {
	dm := &DownloadManager{
		downloadPool: gorkpool.NewUnboundedGorkPool[DownloadTask](ctx, len(td.Pieces)),
		pm:           NewPieceManager(td.Pieces),
	}

	go func() {
		for addr := range peerAddrCh {
			dm.downloadPool.AddWorker(
				NewDownloadWorker(addr, td.Hash, uint32(td.PieceLength), dm.pm),
			)
		}
	}()

	return dm
}
