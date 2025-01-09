package node

import (
	"testing"

	"github.com/ackhia/flash/crypto"
	ma "github.com/multiformats/go-multiaddr"
)

func TestGetTransaction(t *testing.T) {
	serverNode, clientNode := Node{}, Node{}

	clientPrivKey, _ := crypto.CreateKeyPair()
	serverPrivKey, _ := crypto.CreateKeyPair()

	serverNode.Init(clientPrivKey, []string{})
	serverAddr := serverNode.host.Addrs()[0].String()

	maddr, err := ma.NewMultiaddr(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create Multiaddr: %v", err)
	}

	clientNode.Init(serverPrivKey, []string{maddr.String() + "/p2p/" + serverNode.host.ID().String()})
}
