package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"

	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"

	"github.com/ackhia/flash/pb/github.com/ackhia/flash/pb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	proto "google.golang.org/protobuf/proto"

	ma "github.com/multiformats/go-multiaddr"
)

const protocolId protocol.ID = "/echo/1.0.0"

func StartP2p(listenF int, targetF string, insecureF bool, seedF int64) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// Make a host that listens on the given multiaddress
	ha, err := makeBasicHost(listenF, insecureF, seedF)
	if err != nil {
		log.Fatal(err)
	}

	if targetF == "" {
		startListener(ha, listenF, insecureF)
		// Run until canceled.
		<-ctx.Done()
	} else {
		runSender(ha, targetF)

		reader := bufio.NewReader(os.Stdin)

		// Wait for an Enter key press
		_, _ = reader.ReadString('\n')
	}
}

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It won't encrypt the connection if insecure is true.
func makeBasicHost(listenPort int, insecure bool, randseed int64) (host.Host, error) {
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it at least
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
	}

	if insecure {
		opts = append(opts, libp2p.NoSecurity)
	}

	return libp2p.New(opts...)
}

func getHostAddress(ha host.Host) string {
	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", ha.ID()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := ha.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}

func startListener(ha host.Host, listenPort int, insecure bool) {
	fullAddr := getHostAddress(ha)
	log.Printf("I am %s\n", fullAddr)

	ha.SetStreamHandler(protocolId, func(s network.Stream) {
		defer s.Close()
		log.Println("listener received new tx")

		data, err := receiveMessage(s)
		if err != nil {
			s.Reset()
			log.Println(err)
			return
		}

		tx := pb.Tx{}
		if err := proto.Unmarshal(data, &tx); err != nil {
			log.Print("failed to unmarshal message: %w", err)
			return
		}

		log.Printf("Received: %f from: %s to: %s", tx.Amount, tx.From, tx.To)
	})

	log.Println("listening for connections")

	if insecure {
		log.Printf("Now run \"./flash -l %d -d %s -insecure\" on a different terminal\n", listenPort+1, fullAddr)
	} else {
		log.Printf("Now run \"./flash -l %d -d %s\" on a different terminal\n", listenPort+1, fullAddr)
	}
}

func runSender(ha host.Host, targetPeer string) error {
	fullAddr := getHostAddress(ha)
	log.Printf("I am %s\n", fullAddr)

	// Turn the targetPeer into a multiaddr.
	maddr, err := ma.NewMultiaddr(targetPeer)
	if err != nil {
		log.Println(err)
		return err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return err
	}

	// We have a peer ID and a targetAddr, so we add it to the peerstore
	// so LibP2P knows how to contact it
	ha.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	log.Println("sender opening stream")
	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set above because
	// we use the same /echo/1.0.0 protocol
	s, err := ha.NewStream(context.Background(), info.ID, protocolId)
	if err != nil {
		log.Println(err)
		return err
	}
	defer s.Close()

	tx := &pb.Tx{To: "You", From: fullAddr, Amount: 10}

	sendMessage(s, &tx)

	return nil
}

func sendMessage[T proto.Message](w io.Writer, msg *T) error {
	// Serialize the Protobuf message to bytes
	data, err := proto.Marshal(*msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write the length of the message as a fixed-size 4-byte header
	length := uint32(len(data))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// Write the serialized message to the stream
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}

	return nil
}

type protoMessage[T any] interface {
	proto.Message
	*T
}

func receiveMessage(r io.Reader) ([]byte, error) {

	//func receiveMessage[T any/*proto.Message*/](r io.Reader) (T, error) {
	// Read the length of the message (4 bytes)
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}

	// Read the message data
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}

	return data, nil
}
