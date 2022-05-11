package remote

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	log "github.com/sirupsen/logrus"
)

//interface to be called when new  peer is found
func (rfs *RemoteFilesystem) HandlePeerFound(pi peer.AddrInfo) {
	foundPeerId := pi.ID.Pretty()
	if foundPeerId == rfs.hostId {
		return
	}
	log.Info("Found peer: ", pi.ID.Pretty(), ", connecting")
	rfs.handshakePeer(pi)
}

//Initialize the MDNS service
func (rfs *RemoteFilesystem) initMDNS() {
	// register with service so that we get notified about peer discovery
	// An hour might be a long long period in practical applications. But this is fine for us
	ser := mdns.NewMdnsService(rfs.host, rfs.networkName, rfs)
	if err := ser.Start(); err != nil {
		log.Error("failed to start mDNS", err)
	}
}
