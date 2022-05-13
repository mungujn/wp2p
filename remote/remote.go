package remote

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	mrand "math/rand"

	libp2phost "github.com/libp2p/go-libp2p-core/host"

	log "github.com/sirupsen/logrus"

	"bufio"
	"crypto/rand"
	"fmt"

	"github.com/mungujn/web-exp/app"

	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"

	"github.com/multiformats/go-multiaddr"
)

const (
	remoteFilepath = "p"
	handshake      = "h"
	remoteData     = "d"
	remoteError    = "e"
)

// RemoteFilesystem is a remote filesystem host built around libp2p
type RemoteFilesystem struct {
	iam                 string
	rootFolder          string
	listenHost          string
	listenPort          int
	networkName         string
	protocolId          string
	host                libp2phost.Host
	hostId              string
	peerIds             map[string]peer.AddrInfo
	usernameToPeerId    map[string]string
	usernames           []string
	runGlobal           bool
	customBootstrapPeer string
	debug               bool
}

// New creates a new RemoteFilesystem host using libp2p
func New(dcfg app.Config) *RemoteFilesystem {
	log.Debugf("setting up the remote node system using host %s and port %d", dcfg.LocalNodeHost, dcfg.LocalNodePort)
	return &RemoteFilesystem{
		iam:                 dcfg.Username,
		rootFolder:          dcfg.LocalRootFolder,
		listenHost:          dcfg.LocalNodeHost,
		listenPort:          dcfg.LocalNodePort,
		networkName:         dcfg.NetworkName,
		protocolId:          fmt.Sprintf("/%s/%s", dcfg.ProtocolId, dcfg.ProtocolVersion),
		runGlobal:           dcfg.RunGlobal,
		customBootstrapPeer: dcfg.CustomBootstrapPeer,
		debug:               dcfg.Debug,
		peerIds:             make(map[string]peer.AddrInfo),
		usernameToPeerId:    make(map[string]string),
		usernames:           make([]string, 0),
	}
}

// StartHost starts up the host
func (rfs *RemoteFilesystem) StartHost(ctx context.Context) error {
	log.Infof("will listen on: %s with port: %d\n", rfs.listenHost, rfs.listenPort)

	var r io.Reader

	if rfs.debug {
		log.Info("debug mode is on, will use port number as random seed data")
		r = mrand.New(mrand.NewSource(int64(9223372036854775807 - rfs.listenPort)))
	} else {
		r = rand.Reader
	}

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

	var idht *dht.IpfsDHT
	var host libp2phost.Host

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	if rfs.runGlobal {
		log.Info("running global node")
		host, err = getGlobalHost(ctx, prvKey, sourceMultiAddr, idht)
	} else {
		log.Info("running local node")
		host, err = libp2p.New(
			libp2p.ListenAddrs(sourceMultiAddr),
			libp2p.Identity(prvKey),
		)
	}

	if err != nil {
		log.Error("error creating host:", err)
	}
	rfs.host = host
	rfs.setUpGracefulHostStop(ctx)

	// Set a function as stream handler.
	// This function is called when a peer initiates a connection and starts a stream with this peer.
	rfs.host.SetStreamHandler(protocol.ID(rfs.protocolId), rfs.handleStream)
	rfs.hostId = host.ID().Pretty()

	log.Debugf("\nthis hosts Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", rfs.listenHost, rfs.listenPort, rfs.hostId)
	log.Debug("this hosts peer ID Is: ", rfs.hostId)

	rfs.initMDNS()

	if rfs.runGlobal {
		connectToPublicPeers(ctx, rfs.host, rfs.customBootstrapPeer)
	}

	return nil
}

// GetFile returns a file from any peer, including the current one
func (rfs *RemoteFilesystem) GetFile(ctx context.Context, username, path string) ([]byte, error) {
	var fullPath string
	if username == "" {
		fullPath = rfs.rootFolder + "/" + path
		log.Debug("reading local file: ", fullPath)
		return ioutil.ReadFile(fullPath)
	} else {
		log.Debugf("reading remote file %s from user %s ", path, username)

		peerId, exists := rfs.usernameToPeerId[username]
		if !exists {
			return nil, fmt.Errorf("username %s is not assciated with a peer id", username)
		}
		peer, exists := rfs.peerIds[peerId]
		if !exists {
			return nil, fmt.Errorf("peer %s not known by current node", username)
		}

		stream, err := rfs.host.NewStream(ctx, peer.ID, protocol.ID(rfs.protocolId))
		if err != nil {
			return nil, fmt.Errorf("error opening stream to peer %s: %s", username, err)
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		err = rfs.writeData(rw, []byte(path), remoteFilepath)
		if err != nil {
			return nil, fmt.Errorf("error requesting file from peer %s: %s", username, err)
		}

		data, peerUsername, getting, err := readData(rw)
		if getting != remoteData {
			err = fmt.Errorf("not getting data bytes: %s", string(data))
			log.Error(err)
		}
		if peerUsername != username {
			err = fmt.Errorf("peer %s is not %s", peerUsername, username)
		}
		if err != nil {
			return nil, err
		}

		return data, nil
	}
}

// GetOnlineNodes returns usernames of all online nodes
func (rfs *RemoteFilesystem) GetOnlineNodes() []string {
	return rfs.usernames
}

// handshakePeer initiates a handshake with a peer for username exchange
func (rfs *RemoteFilesystem) handshakePeer(peer peer.AddrInfo) {
	peerId := peer.ID.Pretty()
	log.Info("handshake peer: ", peerId)
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
		rfs.peerIds[peerId] = peer
		log.Infof("peer %s now known by current node", peerId)

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		err = rfs.writeData(rw, []byte(rfs.hostId), handshake)
		if err != nil {
			log.Error("error writing handshake message: ", err)
		} else {
			log.Info("handshake message sent to peer: ", peerId)
		}
	}
}

// handleStream is called by libp2p when a stream with the protocol ID is opened
func (rfs *RemoteFilesystem) handleStream(stream network.Stream) {
	log.Info("got a new connection stream")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	data, sender, getting, err := readData(rw)
	if err != nil {
		log.Error("error reading data: ", err)
	}
	switch getting {
	case handshake:
		log.Infof("got handshake from: %s", sender)
		peerId := string(data)
		peerId = strings.Replace(peerId, "\n", "", -1)
		rfs.usernames = append(rfs.usernames, sender)
		rfs.usernameToPeerId[sender] = peerId
		log.Infof("connected to peer: %s with username %s", peerId, sender)
	case remoteFilepath:
		path := string(data)
		path = strings.Replace(path, "\n", "", -1)
		log.Infof("user: %s is requesting %s", sender, path)
		responseData, err := rfs.GetFile(context.TODO(), "", path)
		if err != nil {
			errStr := fmt.Sprintf("error reading file: %s", err)
			log.Error(errStr)
			err = rfs.writeData(rw, []byte(errStr), remoteError)
		} else {
			log.Infof("file %s read successfully", path)
			err = rfs.writeData(rw, responseData, remoteData)
		}
		if err != nil {
			log.Error("error writing response data: ", err)
		} else {
			log.Infof("sent response to user %s", sender)
		}
	case remoteData:
		log.Error("getting remote data from: ", sender)
	case remoteError:
		log.Error("getting remote error data from: ", sender)
	default:
		log.Error("unknown data type: ", getting)
	}
	// 'stream' will stay open until closed (or the other side closes it).
}

// writeData writes a message to a stream
func (rfs *RemoteFilesystem) writeData(rw *bufio.ReadWriter, data []byte, sending string) error {
	sendingCount := len(data)
	_, err := rw.WriteString(fmt.Sprintf("%s:%s:%d\n", rfs.iam, sending, sendingCount))
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

// readData reads data from a stream
func readData(rw *bufio.ReadWriter) ([]byte, string, string, error) {
	theyAre, err := rw.ReadString('\n')
	if err != nil {
		log.Error("error reading from buffer: ", err)
		return nil, "", "", err
	}

	log.Debug("string received from client: ", theyAre)

	if theyAre != "" {
		parts := strings.Split(theyAre, ":")
		if len(parts) != 3 {
			return nil, "", "", errors.New("invalid handshake message")
		}
		clientUsername := parts[0]
		sending := parts[1]
		countStr := parts[2]
		countStr = strings.Replace(countStr, "\n", "", -1)
		sendingCount, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, "", "", errors.New("invalid handshake message, invalid byte count")
		}

		data := make([]byte, sendingCount)
		count, err := rw.Read(data)
		if err == io.EOF || err == nil {
			if count != sendingCount {
				return nil, "", "", fmt.Errorf(
					"invalid read message, expected %d bytes, got %d", sendingCount, count,
				)
			}
			log.Debugf("read %d bytes received from client", count)
			return data, clientUsername, sending, nil
		} else {
			log.Error("error reading from buffer: ", err)
			return nil, "", "", err
		}
	} else {
		err = errors.New("client did not send identity")
	}

	return nil, "", "", err
}

// setUpGracefulHostStop sets up a graceful shutdown of the host
func (rfs *RemoteFilesystem) setUpGracefulHostStop(ctx context.Context) error {
	go func(host libp2phost.Host) {
		<-ctx.Done()
		log.Error("Got Interrupt signal, stopping host")
		host.Close()
	}(rfs.host)
	return nil
}
