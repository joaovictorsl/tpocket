package torrent

import (
	"fmt"
	"strings"

	"github.com/joaovictorsl/tpocket/util"
)

type TorrentInfo struct {
	// A UTF-8 encoded string which is the suggested name to save the
	// file (or directory) as.
	//
	// It is purely advisory.
	Name string
	// Number of bytes in each piece the file is split into. For the
	// purposes of transfer, files are split into fixed-size pieces
	// which are all the same length except for possibly the last one
	// which may be truncated.
	//
	// This is almost always a power of two, most commonly
	// 2 18 = 256 K (BitTorrent prior to version 3.2 uses 2 20 = 1 M as default)
	PieceLength int
	// String whose length is a multiple of 20. It is to be subdivided into
	// strings of length 20, each of which is the SHA1 hash of the piece at
	// the corresponding index.
	Pieces []string
	// Length of the file in bytes.
	//
	// If this field is != 0 then the download
	// represents a single file.
	Length int
	// List of all the files in the download.
	//
	// If this field is != nil then the download
	// represents multiple files.
	Files []*TorrentFileInfo
	// Info hash
	Hash []byte
}

func NewTorrentInfo(name string, pieceLen int, pieces []string, length int, files []*TorrentFileInfo) *TorrentInfo {
	ti := &TorrentInfo{
		Name:        name,
		PieceLength: pieceLen,
		Pieces:      pieces,
		Length:      length,
		Files:       files,
	}

	ti.Hash = util.CalcHash([]byte(ti.encode()))

	return ti
}

func (ti TorrentInfo) encode() string {
	var encoded string
	piecesStr := strings.Join(ti.Pieces, "")
	if ti.Length != 0 {
		encoded = fmt.Sprintf(
			"d6:lengthi%de4:name%d:%s12:piece lengthi%de6:pieces%d:%se",
			ti.Length,
			len(ti.Name),
			ti.Name,
			ti.PieceLength,
			len(piecesStr),
			piecesStr,
		)
	} else {
		files := "l"
		for _, f := range ti.Files {
			files += f.encode()
		}
		files += "e"

		encoded = fmt.Sprintf(
			"d5:files%s4:name%d:%s12:piece lengthi%de6:pieces%d:%se",
			files,
			len(ti.Name),
			ti.Name,
			ti.PieceLength,
			len(piecesStr),
			piecesStr,
		)
	}

	return encoded
}

type TorrentFileInfo struct {
	// Length of the file in bytes.
	Length int
	// A list of UTF-8 encoded strings corresponding to subdirectory names,
	// the last of which is the actual file name.
	Path []string
}

func (tfi TorrentFileInfo) encode() string {
	path := ""
	for _, s := range tfi.Path {
		path += fmt.Sprintf("%d:%s", len(s), s)
	}

	return fmt.Sprintf(
		"d6:lengthi%de4:pathl%see",
		tfi.Length,
		path,
	)
}
