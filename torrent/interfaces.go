package torrent

import "net"

type ITorrentData interface {
	// URL of trackers
	Announcers() []string
	// Information about the file(s)
	Info() ITorrentInfo
}

type ITorrentInfo interface {
	// A UTF-8 encoded string which is the suggested name to save the
	// file (or directory) as.
	//
	// It is purely advisory.
	Name() string
	// Number of bytes in each piece the file is split into. For the
	// purposes of transfer, files are split into fixed-size pieces
	// which are all the same length except for possibly the last one
	// which may be truncated.
	//
	// This is almost always a power of two, most commonly
	// 2 18 = 256 K (BitTorrent prior to version 3.2 uses 2 20 = 1 M as default)
	PieceLength() uint64
	// String whose length is a multiple of 20. It is to be subdivided into
	// strings of length 20, each of which is the SHA1 hash of the piece at
	// the corresponding index.
	Pieces() []string
	// Length of the file(s) in bytes.
	TotalLength() uint64
	// List of all files which should be downloaded.
	Files() []ITorrentFileInfo
	// Info hash
	Hash() []byte
}

type ITorrentFileInfo interface {
	// Length of the file in bytes.
	Length() uint64
	// A list of UTF-8 encoded strings corresponding to subdirectory names,
	// the last of which is the actual file name.
	Path() []string
	// File name
	Name() string
}

type ITrackerResponse interface {
	// Announcer
	Announcer() string
	// Interval between announces
	Interval() int
	// Peers retrieved from the tracker
	Peers() []net.Addr
}
