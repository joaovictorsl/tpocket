package download

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/joaovictorsl/tpocket/messages"
)

type Piece struct {
	Idx      uint32
	Hash     []byte
	SavePath string
}

type DownloadedPiece struct {
	Worker net.Addr
	Idx    uint32
}

type DownloadWorker struct {
	id       net.Addr
	inputCh  chan Piece
	outputCh chan DownloadedPiece
	infoHash []byte
	pieceLen uint32
	bitfield []byte

	pc     *PeerConn
	choked bool

	log     *log.Logger
	logFile *os.File
}

func NewDownloadWorker(peerAddr net.Addr, inputCh chan Piece, outputCh chan DownloadedPiece, infoHash []byte, pieceLen uint32) *DownloadWorker {
	return &DownloadWorker{
		id:       peerAddr,
		inputCh:  inputCh,
		outputCh: outputCh,
		infoHash: infoHash,
		pieceLen: pieceLen,
		pc:       NewPeerConn(peerAddr, pieceLen),
		choked:   true,
	}
}

func (w *DownloadWorker) ID() net.Addr {
	return w.id
}

func (w *DownloadWorker) SignalRemoval() {
	// This is here so we can implement the interface
	// We won't be needing this
}

func (w *DownloadWorker) Process() {
	w.startLogger()
	w.log.Println("Starting")
	w.log.Println("Piece length", w.pieceLen)

	defer w.logFile.Close()

	attempts := 0
	for attempts < 10 {
		err := w.handleDownload()
		if err != nil {
			w.log.Println("SAIU")
			attempts++
		}
		w.log.Println("Attempts", attempts)
		time.Sleep(5 * time.Second)
	}
}

func (w *DownloadWorker) hasPiece(piece int) bool {
	targetByte := piece / 8
	targetBit := byte(8 - piece%8)

	return (w.bitfield[targetByte]>>targetBit)&1 == 1
}

func (w *DownloadWorker) handleDownload() error {
	err := w.pc.Handshake(w.infoHash)
	if err != nil {
		w.log.Println("Handshake error", err)
		return err
	}
	defer w.pc.Close()

	data := make([]byte, w.pieceLen)

pieceLoop:
	for p := range w.inputCh {
		if w.bitfield != nil && !w.hasPiece(int(p.Idx)) {
			w.log.Printf("Peer does not have piece %d\n", p.Idx)
			w.inputCh <- p
			continue
		}
		w.log.Printf("Downloading piece %d\n", p.Idx)

		downloaded := uint32(0)
		requested := uint32(0)
		backlog := uint32(0)

		for downloaded < w.pieceLen {
			var msg messages.PeerMessage
			if !w.choked {
				for ; backlog < 10 && requested < w.pieceLen; backlog++ {
					err := w.pc.SendRequest(p.Idx, requested, w.pieceLen)
					if err != nil {
						w.log.Println("Error when requesting piece", err)
						w.inputCh <- p
						continue pieceLoop
					}
					w.log.Println("Requested", requested)
					requested += 16384
				}
			}

			w.log.Println("Waiting for message")
			msg, err := w.pc.ReadMessage()
			if err != nil {
				w.log.Println("Error when reading message", err)
				w.inputCh <- p
				return err
			}

			switch msg.Type() {
			case messages.CHOKE:
				// Do something
				w.log.Println("CHOKE")
				w.choked = true
				w.inputCh <- p
				continue pieceLoop
			case messages.UNCHOKE:
				w.log.Println("UNCHOKE")
				w.choked = false

				for ; backlog < 5 && requested < w.pieceLen; backlog++ {
					err := w.pc.SendRequest(p.Idx, requested, w.pieceLen)
					if err != nil {
						w.log.Println("Error when requesting piece", err)
						w.inputCh <- p
						continue pieceLoop
					}
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
					w.inputCh <- p
					continue pieceLoop
				}

				bitfieldMsg, _ := msg.(*messages.BitfieldMessage)
				w.bitfield = bitfieldMsg.Bitfield()
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

				// Request other piece
				if requested < w.pieceLen {
					err := w.pc.SendRequest(p.Idx, requested, w.pieceLen)
					if err != nil {
						w.log.Println("Error when requesting piece", err)
						w.inputCh <- p
						continue pieceLoop
					}
					requested += 16384
				}
			case messages.CANCEL:
				// Do something
				w.log.Println("CANCEL")
			default:
				w.log.Println("Invalid message type", msg.Type())
				w.inputCh <- p
				return err
			}
		}

		if !w.pc.HashMatches(data, p.Hash) {
			w.log.Println("Hash doesn't match")
			w.inputCh <- p
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
		w.outputCh <- DownloadedPiece{
			Worker: w.id,
			Idx:    p.Idx,
		}
	}

	return nil
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
	// TODO: Create log folder
	f, err := os.OpenFile("./log/"+w.pc.addr.String()+"_log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.log = log.New(f, "[DownloadWorker] ", log.Flags())
	w.logFile = f

	return nil
}
