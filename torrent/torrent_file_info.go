package torrent

import (
	"fmt"
)

type torrentFileInfo struct {
	// Length of the file in bytes.
	length uint64
	// A list of UTF-8 encoded strings corresponding to subdirectory names,
	// the last of which is the actual file name.
	path []string
}

func (tfi torrentFileInfo) Length() uint64 {
	return tfi.length
}

func (tfi torrentFileInfo) Path() []string {
	return tfi.path
}

func (tfi torrentFileInfo) Name() string {
	return tfi.path[len(tfi.path)-1]
}

func (tfi torrentFileInfo) encode() string {
	path := ""
	for _, s := range tfi.path {
		path += fmt.Sprintf("%d:%s", len(s), s)
	}

	return fmt.Sprintf(
		"d6:lengthi%de4:pathl%see",
		tfi.length,
		path,
	)
}

func filesFrom(source map[string]interface{}) ([]*torrentFileInfo, error) {
	iFiles, err := getField[[]interface{}]("files", source)
	if err != nil {
		return nil, err
	}

	files := make([]*torrentFileInfo, 0)
	for _, m := range iFiles {
		m2, ok := m.(map[string]interface{})
		if !ok {
			fmt.Println("Deu ruim1")
		}
		tfi := &torrentFileInfo{}
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

		tfi.length = uint64(length)
		tfi.path = path2

		files = append(files, tfi)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("files cannot be empty")
	}

	return files, nil
}
