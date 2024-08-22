package download

import (
	"fmt"
	"sync/atomic"
)

type Piece struct {
	Idx      uint32
	Hash     []byte
	SavePath string
}

type PieceManager struct {
	downloadedPieces atomic.Uint32
	totalPieces      uint32
	pieceCh          chan Piece
	pieces           []Piece
	done             chan struct{}
}

func NewPieceManager(pieces []string) *PieceManager {
	pieceCh := make(chan Piece, len(pieces))
	for i, p := range pieces {
		pieceCh <- Piece{
			Idx:      uint32(i),
			Hash:     []byte(p),
			SavePath: fmt.Sprintf("piece_%d", i),
		}
	}

	return &PieceManager{
		downloadedPieces: atomic.Uint32{},
		totalPieces:      uint32(len(pieces)),
		pieceCh:          pieceCh,
		pieces:           make([]Piece, len(pieces)),
		done:             make(chan struct{}),
	}
}

func (pm *PieceManager) Notify(p Piece) {
	downloadedPieces := pm.downloadedPieces.Add(1)

	pm.presentProgress(downloadedPieces)

	if downloadedPieces == pm.totalPieces {
		pm.done <- struct{}{}
		close(pm.pieceCh)
	}

	pm.pieces[p.Idx] = p
}

func (pm *PieceManager) presentProgress(progress uint32) {
	fmt.Printf("\r%.2f%%", (float32(progress)/float32(pm.totalPieces))*100)
}
