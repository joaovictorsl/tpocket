package discover

import (
	"context"
	"net"

	"github.com/joaovictorsl/gorkpool"
	"github.com/joaovictorsl/tpocket/torrent"
)

type PeerDiscovery struct {
	pool *gorkpool.GorkPool[string, struct{}, net.Addr]
	td   torrent.ITorrentData
}

func NewPeerDiscovery(ctx context.Context, td torrent.ITorrentData) *PeerDiscovery {
	pool := gorkpool.NewGorkPool(
		ctx,
		make(chan struct{}),
		make(chan net.Addr),
		func(id string, ic chan struct{}, oc chan net.Addr) (gorkpool.GorkWorker[string, struct{}, net.Addr], error) {
			return NewPeerDiscoverWorker(id, td.Info().Hash(), td.Info().TotalLength(), ic, oc), nil
		},
	)

	return &PeerDiscovery{
		pool: pool,
		td:   td,
	}
}

func (pd *PeerDiscovery) Start() {
	for _, a := range pd.td.Announcers() {
		pd.pool.AddWorker(a)
	}
}

func (pd PeerDiscovery) AddrCh() chan net.Addr {
	return pd.pool.OutputCh()
}

func (pd *PeerDiscovery) GetMorePeers() {
	pd.pool.AddTask(struct{}{})
}
