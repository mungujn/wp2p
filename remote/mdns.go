package remote

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	log "github.com/sirupsen/logrus"
)

// HandlePeerFound to be called when new  peer is found
func (rfs *RemoteFilesystem) HandlePeerFound(pi peer.AddrInfo) {
	foundPeerId := pi.ID.Pretty()
	if foundPeerId == rfs.hostId {
		return
	}
	log.Info("found peer: ", pi.ID.Pretty(), ", connecting")
	rfs.handshakePeer(pi)
}

// initMDNS starts the mDNS service for notification about peer discovery
func (rfs *RemoteFilesystem) initMDNS() {
	ser := mdns.NewMdnsService(rfs.host, rfs.networkName, rfs)
	if err := ser.Start(); err != nil {
		log.Error("failed to start mDNS", err)
	}
}
