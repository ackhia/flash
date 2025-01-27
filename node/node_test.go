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

func createNetworkTwoPeers(t *testing.T, clientBalance float64, serverBalance float64) (*Node, *Node) {
	mn := mocknet.New()

	clientHost, err := mn.GenPeer()
	assert.NoError(t, err)

	serverHost, err := mn.GenPeer()
	assert.NoError(t, err)

	err = mn.LinkAll()
	assert.NoError(t, err)

	genesis := make(map[string]float64)
	genesis[clientHost.ID().String()] = clientBalance
	genesis[serverHost.ID().String()] = serverBalance

	privKey := serverHost.Peerstore().PrivKey(serverHost.ID())
	serverNode := New(privKey, &serverHost, genesis, []string{})
	serverNode.Start()
	serverMultiAddr := createMultiaddress(t, serverNode)

	privKey = clientHost.Peerstore().PrivKey(clientHost.ID())
	clientNode := New(privKey, &clientHost, genesis, []string{serverMultiAddr})
	clientNode.Start()

	log.Printf("Client peer ID: %s", clientNode.Host.ID())
	log.Printf("Server peer ID: %s", serverNode.Host.ID())

	return serverNode, clientNode
}

func createNetworkThreePeers(t *testing.T, node1Balance float64, node2Balance float64, node3Balance float64) (*Node, *Node, *Node) {
	mn := mocknet.New()

	node1Host, err := mn.GenPeer()
	assert.NoError(t, err)

	node2Host, err := mn.GenPeer()
	assert.NoError(t, err)

	node3Host, err := mn.GenPeer()
	assert.NoError(t, err)

	err = mn.LinkAll()
	assert.NoError(t, err)

	genesis := make(map[string]float64)
	genesis[node1Host.ID().String()] = node1Balance
	genesis[node2Host.ID().String()] = node2Balance
	genesis[node3Host.ID().String()] = node3Balance

	privKey := node1Host.Peerstore().PrivKey(node1Host.ID())
	node1 := New(privKey, &node1Host, genesis, []string{})
	node1.Start()
	node1MultiAddr := createMultiaddress(t, node1)

	privKey = node2Host.Peerstore().PrivKey(node2Host.ID())
	node2 := New(privKey, &node2Host, genesis, []string{node1MultiAddr})
	node2.Start()
	node2MultiAddr := createMultiaddress(t, node2)

	privKey = node3Host.Peerstore().PrivKey(node3Host.ID())
	node3 := New(privKey, &node3Host, genesis, []string{node1MultiAddr, node2MultiAddr})
	node3.Start()

	return node1, node2, node3
}

func createMultiaddress(t *testing.T, node *Node) string {
	addr := node.Host.Addrs()[0].String()

	maddr, err := ma.NewMultiaddr(addr)
	assert.NoError(t, err)

	serverMultiAddr := maddr.String() + "/p2p/" + node.Host.ID().String()
	return serverMultiAddr
}

func TestVerifyTx_NormalTx(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 1000, 3000)

	pubKeyBytes, err := crypto.MarshalPublicKey(client.Host.Peerstore().PubKey(client.Host.ID()))
	assert.NoError(t, err)

	tx, err := client.BuildTx(client.Host.ID().String(),
		server.Host.ID().String(),
		20,
		pubKeyBytes,
	)

	assert.NoError(t, err)

	err = fcrypto.SignTx(tx, client.privKey)
	assert.NoError(t, err)

	err = client.VerifyTx(tx)
	assert.NoError(t, err)

	if len(tx.Verifiers) != 1 {
		t.Fatal("Verification not found")
	}

	r, err := fcrypto.VerifyVerifier(&tx.Verifiers[0], tx, server.privKey.GetPublic(), server.Host.ID())

	assert.NoError(t, err)

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

	clientTx := &client.Txs[client.Host.ID().String()][0]
	serverTx := &server.Txs[client.Host.ID().String()][0]

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

	pubKeyBytes, err := crypto.MarshalPublicKey(client.Host.Peerstore().PubKey(client.Host.ID()))
	assert.NoError(t, err)

	tx, err := client.BuildTx(client.Host.ID().String(),
		server.Host.ID().String(),
		2000,
		pubKeyBytes)

	assert.NoError(t, err)

	err = fcrypto.SignTx(tx, client.privKey)
	assert.NoError(t, err)

	err = client.VerifyTx(tx)

	assert.Error(t, err)
}

func TestVerifyTx_VerifierBalanceTooLow(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 1000, 500)

	pubKeyBytes, err := crypto.MarshalPublicKey(client.Host.Peerstore().PubKey(client.Host.ID()))
	assert.NoError(t, err)

	tx, err := client.BuildTx(client.Host.ID().String(),
		server.Host.ID().String(),
		100,
		pubKeyBytes)

	assert.NoError(t, err)

	err = fcrypto.SignTx(tx, client.privKey)
	assert.NoError(t, err)

	err = client.VerifyTx(tx)

	assert.Error(t, err)
}

func TestCalcTotalCoins(t *testing.T) {
	server, _ := createNetworkTwoPeers(t, 1000, 3000)
	assert.Equal(t, server.TotalCoins, float64(4000))
}

func TestTransfer_Normal(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 500, 1000)

	toAddr := server.Host.ID().String()
	err := client.Transfer(toAddr, 25)

	assert.NoError(t, err)

	assert.Equal(t, float64(1025), server.Balances[toAddr])
	assert.Equal(t, float64(475), server.Balances[client.Host.ID().String()])
	assert.Equal(t, float64(1025), client.Balances[toAddr])
	assert.Equal(t, float64(475), client.Balances[client.Host.ID().String()])

}

func TestTransfer_TwoTransfers(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 500, 1000)

	toAddr := server.Host.ID().String()
	err := client.Transfer(toAddr, 25)

	assert.NoError(t, err)

	assert.Equal(t, float64(1025), server.Balances[toAddr])
	assert.Equal(t, float64(475), server.Balances[client.Host.ID().String()])
	assert.Equal(t, float64(1025), client.Balances[toAddr])
	assert.Equal(t, float64(475), client.Balances[client.Host.ID().String()])

	err = client.Transfer(toAddr, 30)

	assert.NoError(t, err)

	assert.Equal(t, float64(1055), server.Balances[toAddr])
	assert.Equal(t, float64(445), server.Balances[client.Host.ID().String()])
	assert.Equal(t, float64(1055), client.Balances[toAddr])
	assert.Equal(t, float64(445), client.Balances[client.Host.ID().String()])

	assert.Equal(t, client.nextSequenceNum, 2)
	assert.Equal(t, server.nextSequenceNum, 0)
}

func TestTransfer_BalanceInsufficient(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 500, 1000)

	toAddr := server.Host.ID().String()
	err := client.Transfer(toAddr, 600)

	assert.Error(t, err)

	assert.Equal(t, float64(1000), server.Balances[toAddr])
	assert.Equal(t, float64(500), server.Balances[client.Host.ID().String()])
	assert.Equal(t, float64(1000), client.Balances[toAddr])
	assert.Equal(t, float64(500), client.Balances[client.Host.ID().String()])

}

func TestTransfer_BalanceNoConsensus(t *testing.T) {
	server, client := createNetworkTwoPeers(t, 1500, 1000)

	toAddr := server.Host.ID().String()
	err := client.Transfer(toAddr, 600)

	assert.Error(t, err)

	assert.Equal(t, float64(1000), server.Balances[toAddr])
	assert.Equal(t, float64(1500), server.Balances[client.Host.ID().String()])
	assert.Equal(t, float64(1000), client.Balances[toAddr])
	assert.Equal(t, float64(1500), client.Balances[client.Host.ID().String()])

}

func TestTransfer_NormalThreePeers(t *testing.T) {
	node1, node2, node3 := createNetworkThreePeers(t, 1000, 1000, 1000)

	toAddr := node2.Host.ID().String()
	fromAddr := node1.Host.ID().String()
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

func TestNodeSync(t *testing.T) {
	mn := mocknet.New()

	clientHost, err := mn.GenPeer()
	assert.NoError(t, err)

	serverHost, err := mn.GenPeer()
	assert.NoError(t, err)

	err = mn.LinkAll()
	assert.NoError(t, err)

	genesis := make(map[string]float64)
	genesis[clientHost.ID().String()] = 500
	genesis[serverHost.ID().String()] = 1000

	privKey := serverHost.Peerstore().PrivKey(serverHost.ID())
	serverNode := New(privKey, &serverHost, genesis, []string{})
	serverNode.Start()
	serverMultiAddr := createMultiaddress(t, serverNode)

	privKey = clientHost.Peerstore().PrivKey(clientHost.ID())
	clientNode := New(privKey, &clientHost, genesis, []string{serverMultiAddr})
	clientNode.Start()

	err = clientNode.Transfer(serverHost.ID().String(), 30)
	assert.NoError(t, err)

	//Create and join a new node
	newHost, err := mn.GenPeer()
	assert.NoError(t, err)

	mn.LinkAll()
	assert.NoError(t, err)

	privKey = serverHost.Peerstore().PrivKey(serverHost.ID())
	clientMultiAddr := createMultiaddress(t, clientNode)
	newNode := New(privKey, &newHost, genesis, []string{serverMultiAddr, clientMultiAddr})
	newNode.Start()

	assert.Equal(t, 1, len(newNode.Txs))

	assert.Equal(t, float64(470), newNode.Balances[clientNode.Host.ID().String()])
	assert.Equal(t, float64(1030), newNode.Balances[serverNode.Host.ID().String()])

	clientNode.Transfer(newNode.Host.ID().String(), 20)

	assert.Equal(t, float64(450), newNode.Balances[clientNode.Host.ID().String()])
	assert.Equal(t, float64(1030), newNode.Balances[serverNode.Host.ID().String()])
	assert.Equal(t, float64(20), newNode.Balances[newNode.Host.ID().String()])

	assert.Equal(t, float64(450), clientNode.Balances[clientNode.Host.ID().String()])
	assert.Equal(t, float64(1030), clientNode.Balances[serverNode.Host.ID().String()])
	assert.Equal(t, float64(20), clientNode.Balances[newNode.Host.ID().String()])

	assert.Equal(t, float64(450), serverNode.Balances[clientNode.Host.ID().String()])
	assert.Equal(t, float64(1030), serverNode.Balances[serverNode.Host.ID().String()])
	assert.Equal(t, float64(20), serverNode.Balances[newNode.Host.ID().String()])
}
