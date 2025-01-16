package node

import (
	"bytes"
	"log"
	"testing"

	"github.com/ackhia/flash/crypto"
	"github.com/ackhia/flash/models"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
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

	var genesis []models.GenesisPeer

	genesis = append(genesis, models.GenesisPeer{
		PeerID:  clientHost.ID(),
		Balance: 1000,
	})

	genesis = append(genesis, models.GenesisPeer{
		PeerID:  serverHost.ID(),
		Balance: 0,
	})

	privKey := serverHost.Peerstore().PrivKey(serverHost.ID())
	serverNode.Init(privKey, []string{}, genesis, serverHost)
	serverMultiAddr := createMultiaddress(t, serverNode)

	privKey = clientHost.Peerstore().PrivKey(clientHost.ID())
	clientNode.Init(privKey, []string{serverMultiAddr}, genesis, clientHost)

	log.Printf("Client peer ID: %s", clientNode.host.ID())
	log.Printf("Server peer ID: %s", serverNode.host.ID())

	return serverNode, clientNode
}

func createMultiaddress(t *testing.T, serverNode Node) string {
	serverAddr := serverNode.host.Addrs()[0].String()

	maddr, err := ma.NewMultiaddr(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create Multiaddr: %v", err)
	}

	serverMultiAddr := maddr.String() + "/p2p/" + serverNode.host.ID().String()
	return serverMultiAddr
}

func TestVerifyTx(t *testing.T) {
	server, client := createNetwork(t)

	tx, err := client.BuildTx(client.host.ID().String(),
		server.host.ID().String(),
		20)

	if err != nil {
		t.Fatalf("Could not build tx: %v", err)
	}

	err = crypto.SignTx(tx, client.privKey)
	if err != nil {
		t.Fatalf("Could not sign tx: %v", err)
	}

	err = client.VerifyTx(tx)
	if err != nil {
		t.Fatalf("Could not send tx: %v", err)
	}

	if len(tx.Verifiers) != 1 {
		t.Fatal("Verification not found")
	}

	r, err := crypto.VerifyVerifier(&tx.Verifiers[0], tx, server.privKey.GetPublic(), server.host.ID())

	if err != nil {
		t.Fatalf("VerifyVerifier failed with error %v", err)
	}

	if !r {
		t.Fatal("VerifyVerifier failed")
	}

	//Check the uncommited tx is in the client and server txs
	if len(server.Txs) != 1 {
		t.Fatal("Tx not in server")
	}

	if len(client.Txs) != 1 {
		t.Fatal("Tx not in client")
	}

	clientTx := client.Txs[client.host.ID().String()][0]
	serverTx := server.Txs[client.host.ID().String()][0]

	if !bytes.Equal(clientTx.Sig, tx.Sig) {
		t.Fatal("Invlid client tx")
	}

	if !bytes.Equal(serverTx.Sig, tx.Sig) {
		t.Fatal("Invlid server tx")
	}

	assert.False(t, clientTx.Comitted)
	assert.False(t, serverTx.Comitted)
}
