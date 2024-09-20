package download

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joaovictorsl/gorkpool"
	"github.com/joaovictorsl/tpocket/torrent"
)

type Downloader struct {
	downloadedPieces atomic.Uint32

	ctx    context.Context
	cancel context.CancelFunc

	pool       *gorkpool.GorkPool[net.Addr, Piece, DownloadedPiece]
	td         torrent.ITorrentInfo
	peerAddrCh chan net.Addr
}

func NewDownloader(ctx context.Context, peerAddrCh chan net.Addr, ti torrent.ITorrentInfo) *Downloader {
	pieces := ti.Pieces()
	outputCh := make(chan DownloadedPiece, len(pieces))
	inputCh := make(chan Piece, len(pieces))
	for i, p := range pieces {
		inputCh <- Piece{
			Idx:      uint32(i),
			Hash:     []byte(p),
			SavePath: fmt.Sprintf("result/piece_%d", i), // TODO: Make this configurable
		}
	}

	downloaderCtx, cancel := context.WithCancel(ctx)
	pool := gorkpool.NewGorkPool(
		downloaderCtx,
		inputCh,
		outputCh,
		func(id net.Addr, ic chan Piece, oc chan DownloadedPiece) (gorkpool.GorkWorker[net.Addr, Piece, DownloadedPiece], error) {
			return NewDownloadWorker(id, ic, oc, ti.Hash(), uint32(ti.PieceLength())), nil
		},
	)

	dm := &Downloader{
		downloadedPieces: atomic.Uint32{},
		ctx:              downloaderCtx,
		cancel:           cancel,
		pool:             pool,
		td:               ti,
		peerAddrCh:       peerAddrCh,
	}

	return dm
}

func (dm *Downloader) Start() {
	go dm.monitorPeers()
	go dm.listenForNewPeerAddr()
}

func (dm *Downloader) monitorPeers() {
	peerPerformance := make(map[net.Addr]uint32)
	for res := range dm.pool.OutputCh() {
		if _, ok := peerPerformance[res.Worker]; !ok {
			peerPerformance[res.Worker] = 0
		}

		peerPerformance[res.Worker] += 1

		fmt.Println(res.Worker.String(), peerPerformance[res.Worker])
	}
}

func (dm *Downloader) listenForNewPeerAddr() {
	for addr := range dm.peerAddrCh {
		if !dm.pool.Contains(addr) {
			dm.pool.AddWorker(addr)
		}
	}
}

func (dm *Downloader) AssemblyFile() {
	<-dm.ctx.Done()
	totalPieces := uint32(len(dm.td.Pieces()))

	// Check if all the pieces were downloaded
	// If not, it means we were cancelled
	if dm.downloadedPieces.Load() != totalPieces {
		panic("downloadedPieces != len(dm.td.Pieces())")
	}

	lastPiece := 0
	at := int64(0)

	for _, f := range dm.td.Files() {
		path := f.Path()
		if len(path) > 1 {
			if err := os.MkdirAll(strings.Join(path[:len(path)-1], "/"), 0777); err != nil {
				panic(err)
			}
		}

		resFile, err := os.OpenFile(strings.Join(path, "/"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}

		fmt.Println("File", f.Name())
		remaining := f.Length()
		buff := make([]byte, dm.td.PieceLength())
		for remaining != 0 {
			bytesToRead := dm.td.PieceLength()
			if remaining < dm.td.PieceLength() {
				bytesToRead = remaining
			}
			fmt.Println("lastPiece", lastPiece)
			fmt.Println("at", at)
			fmt.Println("remaining", remaining)
			fmt.Println("bytesToRead", bytesToRead)

			piece, err := os.OpenFile(fmt.Sprintf("piece_%d", lastPiece), os.O_RDONLY, 0666)
			if err != nil {
				panic(err)
			}

			n, err := piece.ReadAt(buff[:bytesToRead], at)
			if err != nil {
				panic(err)
			}

			remaining -= uint64(n)
			resFile.Write(buff[:n])

			if uint64(int(at)+n) == dm.td.PieceLength() {
				lastPiece++
				at = 0
			} else {
				at = at + int64(n)
			}

			fmt.Println("bytesRead", n)

			piece.Close()
		}

		resFile.Close()
	}
}
