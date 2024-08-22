package download

import (
	"context"
	"fmt"
	"net"

	"github.com/joaovictorsl/gorkpool"
	"github.com/joaovictorsl/tpocket/torrent"
)

type FileDownload struct {
	downloadPool *gorkpool.UnboundedGorkPool[Piece]
	td           *torrent.TorrentInfo
	pm           *PieceManager
	peerAddrCh   chan net.Addr
}

func NewFileDownload(ctx context.Context, peerAddrCh chan net.Addr, td *torrent.TorrentInfo) *FileDownload {
	pm := NewPieceManager(td.Pieces)
	dm := &FileDownload{
		downloadPool: gorkpool.NewUnboundedGorkPoolProvideChannel(ctx, pm.pieceCh),
		td:           td,
		pm:           pm,
		peerAddrCh:   peerAddrCh,
	}

	return dm
}

func (dm *FileDownload) Start() {
	go dm.listenForNewPeerAddr(dm.peerAddrCh)
}

func (dm *FileDownload) AssemblyFile() {
	dm.downloadPool.Wait()
	// TODO: Assembly file
	fmt.Println("Assembly File goes here")
}

func (dm *FileDownload) listenForNewPeerAddr(ch chan net.Addr) {
	for addr := range ch {
		fmt.Println("addr", addr)
		dm.downloadPool.AddWorker(
			NewDownloadWorker(addr, dm.td.Hash, uint32(dm.td.PieceLength), dm.pm),
		)
	}
}
