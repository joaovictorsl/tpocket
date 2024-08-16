package torrent

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
	}

	pm.pieces[p.Idx] = p
}

func (pm *PieceManager) presentProgress(progress uint32) {
	fmt.Printf("\r%.2f%%", (float32(progress)/float32(pm.totalPieces))*100)
}

// func (pm *PieceManager) assemblyFile() {
// 	f, err := os.OpenFile("complete", os.O_CREATE|os.O_WRONLY, 0666)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()

// 	begin := int64(0)
// 	var b []byte
// 	for _, p := range pm.pieces {
// 		b, err = os.ReadFile(p.SavePath)
// 		if err != nil {
// 			panic(err)
// 		}

// 		f.WriteAt(b, begin)

// 		os.Remove(p.SavePath)
// 	}

// 	pm.done <- struct{}{}
// }
