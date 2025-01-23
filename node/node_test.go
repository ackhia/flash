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

func createNetworkTwoPeers(t *testing.T, clientBalance float64, serverBalance float64) (Node, Node) {
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

func createNetworkThreePeers(t *testing.T, node1Balance float64, node2Balance float64, node3Balance float64) (Node, Node, Node) {
	mn := mocknet.New()

	node1, node2, node3 := Node{}, Node{}, Node{}

	node1tHost, err := mn.GenPeer()
	if err != nil {
		panic(err)
	}

	node2Host, err := mn.GenPeer()
	if err != nil {
		panic(err)
	}

	node3Host, err := mn.GenPeer()
	if err != nil {
		panic(err)
	}

	if err := mn.LinkAll(); err != nil {
		panic(err)
	}

	genesis := make(map[string]float64)
	genesis[node1tHost.ID().String()] = node1Balance
	genesis[node2Host.ID().String()] = node2Balance
	genesis[node3Host.ID().String()] = node3Balance

	privKey := node1tHost.Peerstore().PrivKey(node1tHost.ID())
	node1.Init(privKey, []string{}, genesis, node1tHost)
	node1MultiAddr := createMultiaddress(t, node1)

	privKey = node2Host.Peerstore().PrivKey(node2Host.ID())
	node2.Init(privKey, []string{node1MultiAddr}, genesis, node2Host)
	node2MultiAddr := createMultiaddress(t, node2)

	privKey = node3Host.Peerstore().PrivKey(node3Host.ID())
	node3.Init(privKey, []string{node1MultiAddr, node2MultiAddr}, genesis, node3Host)

	return node1, node2, node3
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
	server, client := createNetworkTwoPeers(t, 1000, 3000)

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

	clientTx := &client.Txs[client.host.ID().String()][0]
	serverTx := &server.Txs[client.host.ID().String()][0]

	if !bytes.Equal(clientTx.Sig, tx.Sig) {
		t.Fatal("Invlid client tx")
	}

	if !bytes.Equal(serverTx.Sig, tx.Sig) {
		t.Fatal("Invlid server tx")
	}

	assert.False(t, clientTx.Comitted)
	assert.False(t, serverTx.Comitted)

	assert.Equal(t, float64(1000), client.Balances[clientTx.From])
	assert.Equal(t, float64(3000), client.Balances[clientTx.To])

	client.CommitTx(tx)

	assert.True(t, clientTx.Comitted)
	assert.True(t, serverTx.Comitted)

	assert.Equal(t, float64(1000-20), client.Balances[clientTx.From])
	assert.Equal(t, float64(3000+20), client.Balances[clientTx.To])

	assert.Equal(t, len(client.Txs[tx.From]), 1)
	assert.Equal(t, len(client.Txs[tx.To]), 0)
	assert.Equal(t, len(server.Txs[tx.From]), 1)
	assert.Equal(t, len(server.Txs[tx.To]), 0)
}

func TestVerifyTx_ClientBalanceTooLow(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 1000, 3000)

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
	server, client := createNetworkTwoPeers(t, 1000, 500)

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
	server, _ := createNetworkTwoPeers(t, 1000, 3000)
	assert.Equal(t, server.TotalCoins, float64(4000))
}

func TestTransfer_Normal(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 500, 1000)

	toAddr := server.host.ID().String()
	err := client.Transfer(toAddr, 25)

	assert.NoError(t, err)

	assert.Equal(t, float64(1025), server.Balances[toAddr])
	assert.Equal(t, float64(475), server.Balances[client.host.ID().String()])
	assert.Equal(t, float64(1025), client.Balances[toAddr])
	assert.Equal(t, float64(475), client.Balances[client.host.ID().String()])

}

func TestTransfer_TwoTransfers(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 500, 1000)

	toAddr := server.host.ID().String()
	err := client.Transfer(toAddr, 25)

	assert.NoError(t, err)

	assert.Equal(t, float64(1025), server.Balances[toAddr])
	assert.Equal(t, float64(475), server.Balances[client.host.ID().String()])
	assert.Equal(t, float64(1025), client.Balances[toAddr])
	assert.Equal(t, float64(475), client.Balances[client.host.ID().String()])

	err = client.Transfer(toAddr, 30)

	assert.NoError(t, err)

	assert.Equal(t, float64(1055), server.Balances[toAddr])
	assert.Equal(t, float64(445), server.Balances[client.host.ID().String()])
	assert.Equal(t, float64(1055), client.Balances[toAddr])
	assert.Equal(t, float64(445), client.Balances[client.host.ID().String()])

	assert.Equal(t, client.nextSequenceNum, 2)
	assert.Equal(t, server.nextSequenceNum, 0)
}

func TestTransfer_BalanceInsufficient(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 500, 1000)

	toAddr := server.host.ID().String()
	err := client.Transfer(toAddr, 600)

	assert.Error(t, err)

	assert.Equal(t, float64(1000), server.Balances[toAddr])
	assert.Equal(t, float64(500), server.Balances[client.host.ID().String()])
	assert.Equal(t, float64(1000), client.Balances[toAddr])
	assert.Equal(t, float64(500), client.Balances[client.host.ID().String()])

}

func TestTransfer_BalanceNoConsensus(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 1500, 1000)

	toAddr := server.host.ID().String()
	err := client.Transfer(toAddr, 600)

	assert.Error(t, err)

	assert.Equal(t, float64(1000), server.Balances[toAddr])
	assert.Equal(t, float64(1500), server.Balances[client.host.ID().String()])
	assert.Equal(t, float64(1000), client.Balances[toAddr])
	assert.Equal(t, float64(1500), client.Balances[client.host.ID().String()])

}

func TestTransfer_NormalThreePeers(t *testing.T) {
	node1, node2, node3 := createNetworkThreePeers(t, 1000, 1000, 1000)

	toAddr := node2.host.ID().String()
	fromAddr := node1.host.ID().String()
	err := node1.Transfer(toAddr, 25)

	assert.NoError(t, err)

	assert.Equal(t, float64(1025), node1.Balances[toAddr])
	assert.Equal(t, float64(975), node1.Balances[fromAddr])
	assert.Equal(t, float64(1025), node2.Balances[toAddr])
	assert.Equal(t, float64(975), node2.Balances[fromAddr])
	assert.Equal(t, float64(1025), node3.Balances[toAddr])
	assert.Equal(t, float64(975), node3.Balances[fromAddr])

	assert.Equal(t, float64(3000), node1.TotalCoins)
	assert.Equal(t, float64(3000), node2.TotalCoins)
	assert.Equal(t, float64(3000), node3.TotalCoins)
}
