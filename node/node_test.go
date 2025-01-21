package node

import (
	"bytes"
	"log"
	"testing"

	fcrypto "github.com/ackhia/flash/crypto"
	"github.com/libp2p/go-libp2p/core/crypto"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

func createNetwork(t *testing.T, clientBalance float64, serverBalance float64) (Node, Node) {
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

	genesis := make(map[string]float64)
	genesis[clientHost.ID().String()] = clientBalance
	genesis[serverHost.ID().String()] = serverBalance

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

func TestVerifyTx_NormalTx(t *testing.T) {
	server, client := createNetwork(t, 1000, 3000)

	pubKeyBytes, err := crypto.MarshalPublicKey(client.host.Peerstore().PubKey(client.host.ID()))
	if err != nil {
		t.Fatal("Could not get public key")
	}

	tx, err := client.BuildTx(client.host.ID().String(),
		server.host.ID().String(),
		20,
		pubKeyBytes,
	)

	if err != nil {
		t.Fatalf("Could not build tx: %v", err)
	}

	err = fcrypto.SignTx(tx, client.privKey)
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

	r, err := fcrypto.VerifyVerifier(&tx.Verifiers[0], tx, server.privKey.GetPublic(), server.host.ID())

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

func TestVerifyTx_ClientBalanceTooLow(t *testing.T) {
	server, client := createNetwork(t, 1000, 3000)

	pubKeyBytes, err := crypto.MarshalPublicKey(client.host.Peerstore().PubKey(client.host.ID()))
	if err != nil {
		t.Fatal("Could not get public key")
	}

	tx, err := client.BuildTx(client.host.ID().String(),
		server.host.ID().String(),
		2000,
		pubKeyBytes)

	if err != nil {
		t.Fatalf("Could not build tx: %v", err)
	}

	err = fcrypto.SignTx(tx, client.privKey)
	if err != nil {
		t.Fatalf("Could not sign tx: %v", err)
	}

	err = client.VerifyTx(tx)

	assert.Error(t, err)
}

func TestVerifyTx_VerifierBalanceTooLow(t *testing.T) {
	server, client := createNetwork(t, 1000, 500)

	pubKeyBytes, err := crypto.MarshalPublicKey(client.host.Peerstore().PubKey(client.host.ID()))
	if err != nil {
		t.Fatal("Could not get public key")
	}

	tx, err := client.BuildTx(client.host.ID().String(),
		server.host.ID().String(),
		100,
		pubKeyBytes)

	if err != nil {
		t.Fatalf("Could not build tx: %v", err)
	}

	err = fcrypto.SignTx(tx, client.privKey)
	if err != nil {
		t.Fatalf("Could not sign tx: %v", err)
	}

	err = client.VerifyTx(tx)

	assert.Error(t, err)
}

func TestCalcTotalCoins(t *testing.T) {
	server, _ := createNetwork(t, 1000, 3000)
	assert.Equal(t, server.TotalCoins, float64(4000))
}
