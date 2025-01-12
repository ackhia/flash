package node

import (
	"log"
	"testing"

	"github.com/ackhia/flash/crypto"
	"github.com/ackhia/flash/models"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	ma "github.com/multiformats/go-multiaddr"
)

func createNetwork(t *testing.T) (Node, Node) {
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

	privKey := serverHost.Peerstore().PrivKey(serverHost.ID())
	serverNode.Init(privKey, []string{}, serverHost)
	serverAddr := serverNode.host.Addrs()[0].String()

	maddr, err := ma.NewMultiaddr(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create Multiaddr: %v", err)
	}

	serverMultiAddr := maddr.String() + "/p2p/" + serverNode.host.ID().String()

	privKey = clientHost.Peerstore().PrivKey(clientHost.ID())
	clientNode.Init(privKey, []string{serverMultiAddr}, clientHost)

	log.Printf("Client peer ID: %s", clientNode.host.ID())
	log.Printf("Server peer ID: %s", serverNode.host.ID())

	return serverNode, clientNode
}

func TestVerifyTx(t *testing.T) {
	server, client := createNetwork(t)

	tx := models.Tx{
		To:     server.host.ID().String(),
		From:   client.host.ID().String(),
		Amount: 20,
	}

	err := crypto.SignTx(&tx, client.privKey)
	if err != nil {
		t.Fatalf("Could not sign tx: %v", err)
	}

	err = client.SendTx(&tx)
	if err != nil {
		t.Fatalf("Could not send tx: %v", err)
	}

	if len(tx.Verifiers) != 1 {
		t.Fatal("Verification not found")
	}

	r, err := crypto.VerifyVerifier(&tx.Verifiers[0], &tx, server.privKey.GetPublic(), server.host.ID())

	if err != nil {
		t.Fatalf("VerifyVerifier failed with error %v", err)
	}

	if !r {
		t.Fatal("VerifyVerifier failed")
	}
}
