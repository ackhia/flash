package node

import (
	"testing"

	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	ma "github.com/multiformats/go-multiaddr"
)

func TestGetTransaction(t *testing.T) {
	mn := mocknet.New()

	serverNode, clientNode := Node{}, Node{}

	clientHost, err := mn.GenPeer()
	if err != nil {
		panic(err)
	}

	serverHost, err := mn.GenPeer()
	if err != nil {
		panic(err)
	}

	if err := mn.LinkAll(); err != nil {
		panic(err)
	}

	serverNode.Init(nil, []string{}, clientHost)
	serverAddr := serverNode.host.Addrs()[0].String()

	maddr, err := ma.NewMultiaddr(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create Multiaddr: %v", err)
	}

	serverMultiAddr := maddr.String() + "/p2p/" + serverNode.host.ID().String()
	clientNode.Init(nil, []string{serverMultiAddr}, serverHost)
}
