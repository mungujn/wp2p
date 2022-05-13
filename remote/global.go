package remote

import (
	"context"

	"github.com/libp2p/go-libp2p/p2p/net/connmgr"

	"github.com/libp2p/go-libp2p-core/peer"

	libp2phost "github.com/libp2p/go-libp2p-core/host"

	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
)

func connectToPublicPeers(ctx context.Context, host libp2phost.Host, customBootstrapPeer string) {
	// This connects to public bootstrappers
	bootstrapPeers := dht.DefaultBootstrapPeers

	// This uses a custom bootstrapper peer
	// addr, err := multiaddr.NewMultiaddr(customBootstrapPeer)
	// if err != nil {
	// 	log.Error("failed to parse custom bootstrap peer multiaddr: ", err)
	// } else {
	// 	log.Infof("parsed global bootstrap peer address %s successfully", customBootstrapPeer)
	// }

	// bootstrapPeers := []multiaddr.Multiaddr{addr}

	for _, addr := range bootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Error("failed to parse bootstrap peer multiaddr: ", err)
		}

		// We ignore errors as a bootstrap peers may be down
		// and that is fine.
		log.Info("bootstrapping with: ", pi.ID)
		err = host.Connect(ctx, *pi)
		if err != nil {
			log.Error("failed to connect to bootstrap peer: ", err)
		}
	}
}

func getGlobalHost(
	ctx context.Context,
	prvKey crypto.PrivKey,
	sourceMultiAddr multiaddr.Multiaddr,
	idht *dht.IpfsDHT,
) (libp2phost.Host, error) {
	log.Debug("setting up global host")
	cmgi, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater
	)
	if err != nil {
		log.Error("failed to create connection manager: ", err)
		return nil, err
	}
	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
		libp2p.DefaultTransports,
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(cmgi),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h libp2phost.Host) (routing.PeerRouting, error) {
			idht, err := dht.New(ctx, h)
			return idht, err
		}),
		// Let this host use relays and advertise itself on relays if
		// it finds it is behind NAT. Use libp2p.Relay(options...) to
		// enable active relays and more.
		libp2p.EnableAutoRelay(),
		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		//
		// This service is highly rate-limited and should not cause any
		// performance issues.
		libp2p.EnableNATService(),
	)
}
