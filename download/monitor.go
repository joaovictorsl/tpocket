package download

import (
	"fmt"
	"net"
	"time"

	"github.com/joaovictorsl/tpocket/discover"
)

type DownloadMonitor struct {
	discovery *discover.PeerDiscovery
	downloader *Downloader
	peerPerformance map[net.Addr]uint32
}

func NewDownloadMonitor(
	discovery *discover.PeerDiscovery,
	downloader *Downloader,
) *DownloadMonitor {
	return &DownloadMonitor{
		discovery: discovery,
		downloader: downloader,
		peerPerformance: make(map[net.Addr]uint32),
	}
}

func (dm *DownloadMonitor) Start() {
	go dm.monitor()
}

func (dm *DownloadMonitor) monitor() {
	t := time.NewTicker(5 * time.Second)

	select {
	case <-t.C:
		// Judge performance and decide wether or not we should ask for more peers
		dm.discovery.GetMorePeers()
	case r := <-dm.downloader.pool.OutputCh():
		dm.recordPerformance(r)
	}
}

func (dm *DownloadMonitor) recordPerformance(res DownloadedPiece) {
	if _, ok := dm.peerPerformance[res.Worker]; !ok {
		dm.peerPerformance[res.Worker] = 0
	}

	dm.peerPerformance[res.Worker] += 1

	fmt.Println(res.Worker.String(), dm.peerPerformance[res.Worker])
}
