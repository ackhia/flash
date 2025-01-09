package node

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type Node struct {
	host host.Host
}

func (n *Node) Init(privKey crypto.PrivKey, bootstraoPeers []string) {
	log.Print("Node starting")

	n.host, _ = libp2p.New(libp2p.Identity(privKey))

	go n.startTransactionServer()

	for _, item := range bootstraoPeers {
		n.getTransactions(item)
	}
}

func (n Node) getTransactions(addrInfo string) {
	serverAddr, err := peer.AddrInfoFromString(addrInfo)

	if err != nil {
		log.Fatalf("Invalid address %s %s", addrInfo, err)
	}

	// Connect to the server
	n.host.Connect(context.Background(), *serverAddr)

	// Open a stream to the server using the transactions protocol
	stream, _ := n.host.NewStream(context.Background(), serverAddr.ID, protocol.ID("/flash/transactions/1.0.0"))

	// Read the data from the stream
	data, _ := io.ReadAll(stream)
	fmt.Println("Transactions:", string(data))

	stream.Close()
}

func (n Node) startTransactionServer() {
	n.host.SetStreamHandler("/flash/transactions/1.0.0", func(s network.Stream) {
		defer s.Close()
		s.Write([]byte("[{\"id\":1,\"amount\":100},{\"id\":2,\"amount\":200}]"))
	})

	select {}
}
