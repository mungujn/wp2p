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
	remoteFilepath = "p"
	handshake      = "h"
	remoteData     = "d"
	remoteError    = "e"
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
	peerNames   []string
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
		peerNames:   make([]string, 0),
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
	host.SetStreamHandler(protocol.ID(rfs.protocolId), rfs.handleStream)
	rfs.hostId = host.ID().Pretty()

	log.Info("\nthis hosts Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", rfs.listenHost, rfs.listenPort, rfs.hostId)

	rfs.initMDNS()

	return nil
}

func (rfs *RemoteFilesystem) GetFile(ctx context.Context, username, path string) ([]byte, error) {
	var fullPath string
	if username == "" {
		fullPath = rfs.rootFolder + "/" + path
		log.Debug("reading local file: ", fullPath)
		return ioutil.ReadFile(fullPath)
	} else {
		log.Debugf("reading remote file %s from user %s ", path, username)

		peer, exists := rfs.peers[username]
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

func (rfs *RemoteFilesystem) GetOnlineNodes(ctx context.Context) ([]string, error) {
	return rfs.peerNames, nil
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
			if err != nil && err == errors.New("stream reset") {
				log.Error("stream reset: ", err)
			}
			if getting != handshake {
				err = errors.New("not getting handshake")
			}
			if err != nil {
				log.Error("error reading handshake response: ", err)
			} else {
				rfs.peers[peerUsername] = peer
				rfs.peerNames = append(rfs.peerNames, peerUsername)
				log.Infof("connected to peer: %s with username %s", peer, peerUsername)
			}
		}
	}

	log.Info("peer handshake failed")
}

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
	case remoteFilepath:
		path := string(data)
		log.Infof("user: %s is requesting %s", sender, path)
		responseData, err := ioutil.ReadFile(path)
		if err != nil {
			errStr := fmt.Sprintf("error reading file: %s", err)
			err = rfs.writeData(rw, []byte(errStr), remoteError)
		} else {
			err = rfs.writeData(rw, responseData, remoteData)
		}
		if err != nil {
			log.Error("error writing response data: ", err)
		}
	case remoteData:
		log.Error("getting remote data from: ", sender)
	default:
		log.Error("unknown data type: ", getting)
	}
	// 'stream' will stay open until you close it (or the other side closes it).
}

func (rfs *RemoteFilesystem) writeData(rw *bufio.ReadWriter, data []byte, sending string) error {
	_, err := rw.WriteString(fmt.Sprintf("%s:%s\n", rfs.iam, sending))
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
	_, err = rw.Write([]byte{'\n'})
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

func readData(rw *bufio.ReadWriter) ([]byte, string, string, error) {
	theyAre, err := rw.ReadString('\n')
	if err != nil {
		log.Error("error reading from buffer: ", err)
		return nil, "", "", err
	}

	log.Debug("string received from client: ", theyAre)

	if theyAre != "" {
		parts := strings.Split(theyAre, ":")
		if len(parts) != 2 {
			err = errors.New("invalid handshake message")
		}
		clientUsername := parts[0]
		sending := parts[1]
		sending = strings.Replace(sending, "\n", "", -1)

		data, err := rw.ReadBytes('\n')
		if err == nil {
			// log.Debugf("bytes received from client: %s %s %s", data, clientUsername, sending)
			return data, clientUsername, sending, nil
		}
	} else {
		err = errors.New("client did not send identity")
	}

	return nil, "", "", err
}

func (rfs *RemoteFilesystem) setUpGracefulHostStop(ctx context.Context) error {
	go func(host host.Host) {
		<-ctx.Done()
		log.Error("Got Interrupt signal, stopping host")
		host.Close()
	}(rfs.host)
	return nil
}
