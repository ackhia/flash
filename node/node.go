package node

import (
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/ackhia/flash/models"
)

const verifyTxProtocol = "/flash/verify-transaction/1.0.0"

type Node struct {
	host            host.Host
	txDb            []models.Tx
	nextSequenceNum int
	privKey         crypto.PrivKey
}

func (n *Node) Init(privKey crypto.PrivKey, bootstraoPeers []string, host host.Host) {
	log.Print("Node starting")

	n.privKey = privKey

	if host == nil {
		n.host, _ = libp2p.New(libp2p.Identity(privKey))
	} else {
		n.host = host
	}

	go n.startTransactionServer()
	go n.startVerificationServer()

	for _, item := range bootstraoPeers {
		n.getTransactions(item)
	}
}
