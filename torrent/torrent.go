package torrent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
)

var (
	ErrFieldMissing = errors.New("field missing")
)

type TrackerResponse struct {
	Interval int
	Peers    []net.Addr
}

func trackerResponseFrom(source map[string]interface{}) (*TrackerResponse, error) {
	tr := &TrackerResponse{}

	interval, err := getField[int]("interval", source)
	if err != nil {
		return tr, err
	}

	strPeers, err := getField[string]("peers", source)
	if err != nil {
		return tr, err
	}

	peersBytes := []byte(strPeers)
	peers := make([]net.Addr, 0)

	for i := 0; i < len(peersBytes); i += 6 {
		currPeer := peersBytes[i : i+6]
		ip := currPeer[:4]
		portBytes := currPeer[4:]
		port := binary.BigEndian.Uint16(portBytes)

		peers = append(peers, &net.TCPAddr{
			IP:   ip,
			Port: int(port),
		})
	}

	tr.Interval = interval
	tr.Peers = peers

	return tr, nil
}

type TorrentData struct {
	// URL of trackers
	Announcers []string
	// Information about the file(s)
	Info *TorrentInfo
}

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
	// Hash of info
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

	ti.Hash = calcHash([]byte(ti.encode()))

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

func torrentDataFrom(source map[string]interface{}) (*TorrentData, error) {
	td := &TorrentData{}
	ti := &TorrentInfo{}

	// Get announcers
	announcers := make([]string, 0)
	announceList, err := getField[[]interface{}]("announce-list", source)
	if err != nil && err != ErrFieldMissing {
		return td, err
	}
	for _, v := range announceList {
		a, ok := v.([]interface{})
		if !ok {
			panic("Announce list is not a list of list of string")
		}
		announcers = append(announcers, a[0].(string))
	}

	if len(announcers) != 0 {
		announce, err := getField[string]("announce", source)
		if err != nil {
			return td, err
		}
		announcers = append(announcers, announce)
	}

	// Get info
	mapInfo, err := getField[map[string]interface{}]("info", source)
	if err != nil {
		return td, err
	}

	// Get name
	name, err := getField[string]("name", mapInfo)
	if err != nil {
		return td, err
	}

	// Get piece length
	pieceLength, err := getField[int]("piece length", mapInfo)
	if err != nil {
		return td, err
	}

	// Get pieces
	piecesStr, err := getField[string]("pieces", mapInfo)
	if err != nil {
		return td, err
	}
	pieces := make([]string, 0)
	for i := 0; i < len(piecesStr); i += 20 {
		pieces = append(pieces, piecesStr[i:i+20])
	}

	// Get length or files
	_, okL := mapInfo["length"]
	_, okF := mapInfo["files"]
	if okL == okF {
		return td, fmt.Errorf("there can only be a key length or a key files, not both or neither")
	}

	length := 0
	var files []*TorrentFileInfo

	if okL {
		l, err := getField[int]("length", mapInfo)
		if err != nil {
			return td, err
		}

		length = l
	} else {
		fs, err := filesFrom(mapInfo)
		if err != nil {
			return td, err
		}

		files = fs
	}

	ti.Name = name
	ti.PieceLength = pieceLength
	ti.Pieces = pieces
	ti.Length = length
	ti.Files = files

	td.Announcers = announcers
	td.Info = ti

	return td, nil
}

func filesFrom(source map[string]interface{}) ([]*TorrentFileInfo, error) {
	iFiles, err := getField[[]interface{}]("files", source)
	if err != nil {
		return nil, err
	}

	files := make([]*TorrentFileInfo, 0)
	for _, m := range iFiles {
		m2, ok := m.(map[string]interface{})
		if !ok {
			fmt.Println("Deu ruim1")
		}
		tfi := &TorrentFileInfo{}
		length, err := getField[int]("length", m2)
		if err != nil {
			return nil, err
		}

		path, err := getField[[]interface{}]("path", m2)
		if err != nil {
			return nil, err
		}

		path2 := make([]string, 0)
		for _, v := range path {
			vStr, ok := v.(string)
			if !ok {
				fmt.Println("Deu ruim2")
			}

			path2 = append(path2, vStr)
		}

		tfi.Length = length
		tfi.Path = path2

		files = append(files, tfi)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("files cannot be empty")
	}

	return files, nil
}

func getField[T any](field string, source map[string]interface{}) (T, error) {
	var zero T
	iField, ok := source[field]
	if !ok {
		return zero, ErrFieldMissing
	}

	fieldValue, ok := iField.(T)
	if !ok {
		return zero, fmt.Errorf("%s is not a %v, it is a %v", field, reflect.TypeOf(zero), reflect.TypeOf(iField))
	}

	return fieldValue, nil
}
