package main

import (
	"bufio"
	"context"
	"os"
)

func main() {
	files := []string{
		"debian-12.6.0-amd64-netinst.iso.torrent",
		// "slackware-14.2-source-dvd.torrent",
		// "sintel.torrent",
		// "Gunner.2024.720p.WEBRip.800MB.x264-GalaxyRG.torrent",
		// "The.Union.2024.720p.NF.WEBRip.800MB.x264-GalaxyRG.torrent",
	}
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if err := Download(context.Background(), bufio.NewReader(f)); err != nil {
			panic(err)
		}
	}
}
