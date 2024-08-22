package torrent

import (
	"log"
	"net"
	"os"

	"github.com/joaovictorsl/tpocket/torrent/messages"
)

type DownloadWorker struct {
	tskCh    chan DownloadTask
	pc       *PeerConn
	infoHash []byte
	pieceLen uint32
	pm       *PieceManager
	log      *log.Logger
	logFile  *os.File
	choked   bool
}

type DownloadTask struct {
}

func NewDownloadWorker(peerAddr net.Addr, infoHash []byte, pieceLen uint32, pm *PieceManager) *DownloadWorker {
	return &DownloadWorker{
		pc:       NewPeerConn(peerAddr, pieceLen),
		infoHash: infoHash,
		pieceLen: pieceLen,
		pm:       pm,
		choked:   true,
	}
}

func (w *DownloadWorker) ReceiveTaskChannel(ch chan DownloadTask) {

}

func (w *DownloadWorker) SignalRemoval() {

}

func (w *DownloadWorker) Process() {
	err := w.pc.Handshake(w.infoHash)
	if err != nil {
		return
	}

	w.startLogger()
	w.log.Println("Starting")
	w.log.Println("Piece length", w.pieceLen)

	defer w.logFile.Close()
	defer w.pc.Close()

	data := make([]byte, w.pieceLen)

pieceLoop:
	for p := range w.pm.pieceCh {
		w.log.Printf("Downloading piece %d\n", p.Idx)

		downloaded := uint32(0)
		requested := uint32(0)
		backlog := uint32(0)

		for downloaded < w.pieceLen {
			var msg messages.PeerMessage
			if !w.choked {
				for ; backlog < 5 && requested < w.pieceLen; backlog++ {
					err := w.pc.SendRequest(p.Idx, requested, w.pieceLen)
					if err != nil {
						w.log.Println("Error when requesting piece", err)
						w.pm.pieceCh <- p
						continue pieceLoop
					}
					w.log.Println("Requested", requested)
					requested += 16384
				}
			}

			w.log.Println("Waiting for message")
			msg, err = w.pc.ReadMessage()
			if err != nil {
				w.log.Println("Error when reading message", err)
				w.pm.pieceCh <- p
				return
			}

			switch msg.Type() {
			case messages.CHOKE:
				// Do something
				w.log.Println("CHOKE")
				w.choked = true
				w.pm.pieceCh <- p
				continue pieceLoop
			case messages.UNCHOKE:
				w.log.Println("UNCHOKE")
				w.choked = false

				for ; backlog < 5 && requested < w.pieceLen; backlog++ {
					err := w.pc.SendRequest(p.Idx, requested, w.pieceLen)
					if err != nil {
						w.log.Println("Error when requesting piece", err)
						w.pm.pieceCh <- p
						continue pieceLoop
					}
					w.log.Println("Requested", requested)
					requested += 16384
				}
			case messages.INTERESTED:
				// Do something
				w.log.Println("INTERESTED")
			case messages.NOT_INTERESTED:
				// Do something
				w.log.Println("NOT_INTERESTED")
			case messages.HAVE:
				// Do something
				w.log.Println("HAVE")
			case messages.BITFIELD:
				w.log.Println("BITFIELD")
				// Just say you're interested
				err := w.pc.SendInterest()
				if err != nil {
					w.log.Println("Error when sending interest", err)
					w.pm.pieceCh <- p
					continue pieceLoop
				}
			case messages.REQUEST:
				// Do something
				w.log.Println("REQUEST")
			case messages.PIECE:
				msgPiece, ok := msg.(*messages.PieceMessage)
				if !ok {
					w.log.Println("Failed to cast message to PieceMessage")
					continue pieceLoop
				}

				copy(data[downloaded:], msgPiece.Block)
				downloaded += uint32(len(msgPiece.Block))
				w.log.Println("Downloaded", downloaded)

				// Request other piece
				if requested < w.pieceLen {
					err := w.pc.SendRequest(p.Idx, requested, w.pieceLen)
					if err != nil {
						w.log.Println("Error when requesting piece", err)
						w.pm.pieceCh <- p
						continue pieceLoop
					}
					w.log.Println("Requested", requested)
					requested += 16384
				}
			case messages.CANCEL:
				// Do something
				w.log.Println("CANCEL")
			default:
				w.log.Println("Invalid message type", msg.Type())
				w.pm.pieceCh <- p
				return
			}
		}

		if !w.pc.HashMatches(data, p.Hash) {
			w.log.Println("Hash doesn't match")
			w.pm.pieceCh <- p
			continue
		}

		err := w.pc.SendHave(p.Idx)
		if err != nil {
			w.log.Println("Failed to send have message", err)
		}

		err = w.savePiece(p.SavePath, data)
		if err != nil {
			w.log.Println("Failed to save")
		}

		w.log.Printf("Successfully saved piece %d\n", p.Idx)
		w.pm.Notify(p)
	}

	w.log.Println("Done")
}

func (w *DownloadWorker) savePiece(savePath string, data []byte) error {
	f, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

func (w *DownloadWorker) startLogger() error {
	f, err := os.OpenFile("./log/"+w.pc.addr.String()+"_log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.log = log.New(f, "[DownloadWorker] ", log.Flags())
	w.logFile = f

	return nil
}
