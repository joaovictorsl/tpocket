package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joaovictorsl/mytorrent/torrent"
)

func main() {
	files := []string{
		// "debian-12.6.0-amd64-netinst.iso.torrent",
		// "slackware-14.2-source-dvd.torrent",
		"itsworking.gif.torrent",
		"codercat.gif.torrent",
		"sample.torrent",
	}
	for _, file := range files {
		client := &torrent.Client{}
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if err := client.Download(bufio.NewReader(f)); err != nil {
			panic(err)
		}

		assemblyFile(strings.Split(file, ".torrent")[0])
	}
}

func assemblyFile(name string) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	begin := int64(0)
	var b []byte
	for i := 0; i < 2524; i++ {
		savePath := fmt.Sprintf("piece_%d", i)
		b, err = os.ReadFile(savePath)
		if err != nil {
			panic(err)
		}

		n, err := f.WriteAt(b, begin)
		if err != nil {
			panic(err)
		}

		begin += int64(n)

		os.Remove(savePath)
	}
}
