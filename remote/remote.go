package remote

import (
	"context"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/host"

	log "github.com/sirupsen/logrus"

	"bufio"
	"crypto/rand"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"

	"github.com/multiformats/go-multiaddr"
)

const (
	path      = "p"
	handshake = "h"
)

type RemoteFilesystem struct {
	iam         string
	rootFolder  string
	listenHost  string
	listenPort  int
	networkName string
	protocolId  string
	host        host.Host
	hostId      string
	peers       map[string]peer.AddrInfo
}

func New(iam, rootfloder, listenHost string, listenPort int, networkName, protocolId string) *RemoteFilesystem {
	log.Debugf("setting up the remote node system using host %s and port %d", listenHost, listenPort)
	return &RemoteFilesystem{
		iam:         iam,
		rootFolder:  rootfloder,
		listenHost:  listenHost,
		listenPort:  listenPort,
		networkName: networkName,
		protocolId:  protocolId,
		peers:       make(map[string]peer.AddrInfo),
	}
}

func (rfs *RemoteFilesystem) StartHost(ctx context.Context) error {
	log.Infof("will listen on: %s with port: %d\n", rfs.listenHost, rfs.listenPort)

	r := rand.Reader

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		log.Error("error generating host identity: ", err)
		return err
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", rfs.listenHost, rfs.listenPort))
	if err != nil {
		log.Error("error generating listening addresses: ", err)
		return err
	}

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		log.Error("error creating host:", err)
	}
	rfs.host = host
	rfs.setUpGracefulHostStop(ctx)

	// Set a function as stream handler.
	// This function is called when a peer initiates a connection and starts a stream with this peer.
	host.SetStreamHandler(protocol.ID(rfs.protocolId), handleStream)
	rfs.hostId = host.ID().Pretty()

	log.Info("\nthis hosts Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", rfs.listenHost, rfs.listenPort, rfs.hostId)

	rfs.initMDNS()

	return nil
}

func (rfs *RemoteFilesystem) GetFile(ctx context.Context, username, path string) ([]byte, error) {
	var fullPath string
	if username == "" {
		fullPath = rfs.rootFolder + "/" + path
	} else {
		fullPath = rfs.rootFolder + "/" + username + "/" + path
	}
	log.Debug("reading file: ", fullPath)
	return ioutil.ReadFile(fullPath)
}

func (lfs *RemoteFilesystem) GetOnlineNodes(ctx context.Context) ([]string, error) {
	return []string{"user_2", "user_3"}, nil
}

func (rfs *RemoteFilesystem) setUpGracefulHostStop(ctx context.Context) error {
	go func(host host.Host) {
		<-ctx.Done()
		log.Error("Got Interrupt signal, stopping host")
		host.Close()
	}(rfs.host)
	return nil
}

func (rfs *RemoteFilesystem) handshakePeer(peer peer.AddrInfo) {
	log.Info("handshake peer: ", peer.ID)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := rfs.host.Connect(ctx, peer); err != nil {
		log.Error("connection to peer failed: ", err)
	}

	// open a stream, this stream will be handled by handleStream other end
	stream, err := rfs.host.NewStream(ctx, peer.ID, protocol.ID(rfs.protocolId))

	if err != nil {
		log.Error("stream open failed: ", err)
	} else {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		err = rfs.writeData(rw, []byte(handshake), handshake)
		if err != nil {
			log.Error("error writing handshake message: ", err)
		} else {
			_, peerUsername, getting, err := readData(rw)
			if getting != handshake {
				err = errors.New("not getting handshake")
			}
			if err != nil {
				log.Error("error reading handshake response: ", err)
			} else {
				rfs.peers[peerUsername] = peer
				fmt.Println("Connected to: ", peer)
			}
		}
	}

	log.Info("peer handshake failed")
}

func (rfs *RemoteFilesystem) writeData(rw *bufio.ReadWriter, data []byte, sending string) error {
	_, err := rw.WriteString(fmt.Sprintf("%s:%s\n", rfs.iam))
	if err != nil {
		log.Error("error writing identity to buffer: ", err)
		return err
	}
	err = rw.Flush()
	if err != nil {
		log.Error("error flushing identity to buffer: ", err)
		return err
	}
	_, err = rw.Write(data)
	if err != nil {
		log.Error("error writing data to buffer: ", err)
		return err
	}
	err = rw.Flush()
	if err != nil {
		log.Error("error flushing data buffer: ", err)
		return err
	}
	return nil
}

func handleStream(stream network.Stream) {
	log.Info("got a new connection stream")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	data, sender, getting, err := readData(rw)
	if err != nil {
		log.Error("error reading data: ", err)
	}
	if getting == handshake {
		log.Infof("got handshake from: %s", sender)
	} else {
		path := string(data)
		log.Infof("user: %s is requesting %s", sender, path)
	}
	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) ([]byte, string, string, error) {
	theyAre, err := rw.ReadString('\n')
	if err != nil {
		log.Error("error reading from buffer: ", err)
		return nil, "", "", err
	}

	log.Info("received from client: ", theyAre)

	if theyAre != "" {
		parts := strings.Split(theyAre, ":")
		if len(parts) != 2 {
			err = errors.New("invalid handshake message")
		}
		clientUsername := parts[0]
		sending := parts[1]

		data, err := rw.ReadBytes('\n')
		if err == nil {
			return data, clientUsername, sending, nil
		}
	} else {
		err = errors.New("client did not send custom identity")
	}

	return nil, "", "", err
}
