package torrent

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrFieldMissing = errors.New("field missing")
)

type TorrentData struct {
	// URL of trackers
	Announcers []string
	// Information about the file(s)
	Info *TorrentInfo
}

func TorrentDataFrom(source map[string]interface{}) (*TorrentData, error) {
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

	length := 0
	var files []*TorrentFileInfo

	if okL {
		l, err := getField[int]("length", mapInfo)
		if err != nil {
			return nil, err
		}

		length = l
	} else {
		fs, err := filesFrom(mapInfo)
		if err != nil {
			return nil, err
		}

		files = fs
	}

	ti := NewTorrentInfo(name, pieceLength, pieces, length, files)

	return &TorrentData{
		Announcers: announcers,
		Info:       ti,
	}, nil
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
