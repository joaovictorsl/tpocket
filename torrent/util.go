package torrent

import "crypto/sha1"

func calcHash(b []byte) []byte {
	h := sha1.New()
	h.Write(b)
	return h.Sum(nil)
}
