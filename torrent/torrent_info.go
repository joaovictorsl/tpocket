package torrent

import (
	"fmt"
	"strings"

	"github.com/joaovictorsl/tpocket/util"
)

type torrentInfo struct {
	// A UTF-8 encoded string which is the suggested name to save the
	// file (or directory) as.
	//
	// It is purely advisory.
	name string
	// Number of bytes in each piece the file is split into. For the
	// purposes of transfer, files are split into fixed-size pieces
	// which are all the same length except for possibly the last one
	// which may be truncated.
	//
	// This is almost always a power of two, most commonly
	// 2 18 = 256 K (BitTorrent prior to version 3.2 uses 2 20 = 1 M as default)
	pieceLength uint64
	// String whose length is a multiple of 20. It is to be subdivided into
	// strings of length 20, each of which is the SHA1 hash of the piece at
	// the corresponding index.
	pieces []string
	// Length of the file in bytes.
	//
	// If this field is != 0 then the download
	// represents a single file.
	length uint64
	// List of all the files in the download.
	//
	// If this field is != nil then the download
	// represents multiple files.
	files []*torrentFileInfo
	// Info hash
	hash []byte
}

func (ti torrentInfo) Name() string {
	return ti.name
}

func (ti torrentInfo) PieceLength() uint64 {
	return ti.pieceLength
}

func (ti torrentInfo) Pieces() []string {
	return ti.pieces
}

func (ti torrentInfo) TotalLength() uint64 {
	length := uint64(0)

	if ti.length != 0 {
		length = ti.length
	} else {
		for _, f := range ti.files {
			length += f.length
		}
	}

	return length
}

func (ti torrentInfo) Files() []ITorrentFileInfo {
	if ti.length != 0 {
		return []ITorrentFileInfo{
			torrentFileInfo{length: ti.length, path: []string{ti.name}},
		}
	}

	files := make([]ITorrentFileInfo, len(ti.files))
	for i, f := range ti.files {
		files[i] = f
	}
	return files
}

func (ti torrentInfo) Hash() []byte {
	return ti.hash
}

func newTorrentInfo(name string, pieceLen uint64, pieces []string, length uint64, files []*torrentFileInfo) *torrentInfo {
	ti := &torrentInfo{
		name:        name,
		pieceLength: pieceLen,
		pieces:      pieces,
		length:      length,
		files:       files,
	}

	ti.hash = util.CalcHash([]byte(ti.encode()))

	return ti
}

func (ti torrentInfo) encode() string {
	var encoded string
	piecesStr := strings.Join(ti.pieces, "")
	if ti.length != 0 {
		encoded = fmt.Sprintf(
			"d6:lengthi%de4:name%d:%s12:piece lengthi%de6:pieces%d:%se",
			ti.length,
			len(ti.name),
			ti.name,
			ti.pieceLength,
			len(piecesStr),
			piecesStr,
		)
	} else {
		files := "l"
		for _, f := range ti.files {
			files += f.encode()
		}
		files += "e"

		encoded = fmt.Sprintf(
			"d5:files%s4:name%d:%s12:piece lengthi%de6:pieces%d:%se",
			files,
			len(ti.name),
			ti.name,
			ti.pieceLength,
			len(piecesStr),
			piecesStr,
		)
	}

	return encoded
}
