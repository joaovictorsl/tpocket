package torrent

import (
	"errors"
	"fmt"
)

var (
	ErrFieldMissing = errors.New("field missing")
)

type torrentData struct {
	// URL of trackers
	announcers []string
	// Information about the file(s)
	info *torrentInfo
}

func (td torrentData) Announcers() []string {
	return td.announcers
}

func (td torrentData) Info() ITorrentInfo {
	return td.info
}

func TorrentDataFrom(source map[string]interface{}) (ITorrentData, error) {
	// Get announcers
	announcers := make([]string, 0)
	announceList, err := getField[[]interface{}]("announce-list", source)
	if err != nil && err != ErrFieldMissing {
		return nil, err
	}
	for _, v := range announceList {
		a, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("announce-list is not a list of list of string")
		}
		announcers = append(announcers, a[0].(string))
	}

	if len(announcers) == 0 {
		announce, err := getField[string]("announce", source)
		if err != nil {
			return nil, err
		}
		announcers = append(announcers, announce)
	}

	// Get info
	mapInfo, err := getField[map[string]interface{}]("info", source)
	if err != nil {
		return nil, err
	}

	// Get name
	name, err := getField[string]("name", mapInfo)
	if err != nil {
		return nil, err
	}

	// Get piece length
	pieceLength, err := getField[int]("piece length", mapInfo)
	if err != nil {
		return nil, err
	}

	// Get pieces
	piecesStr, err := getField[string]("pieces", mapInfo)
	if err != nil {
		return nil, err
	}
	pieces := make([]string, 0)
	for i := 0; i < len(piecesStr); i += 20 {
		pieces = append(pieces, piecesStr[i:i+20])
	}

	// Get length or files
	_, okL := mapInfo["length"]
	_, okF := mapInfo["files"]
	if okL == okF {
		return nil, fmt.Errorf("there can only be a key length or a key files, not both or neither")
	}

	length := uint64(0)
	var files []*torrentFileInfo

	if okL {
		l, err := getField[int]("length", mapInfo)
		if err != nil {
			return nil, err
		}

		length = uint64(l)
	} else {
		fs, err := filesFrom(mapInfo)
		if err != nil {
			return nil, err
		}

		files = fs
	}

	ti := newTorrentInfo(name, uint64(pieceLength), pieces, length, files)

	return &torrentData{
		announcers: announcers,
		info:       ti,
	}, nil
}
