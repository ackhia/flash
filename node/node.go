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
	nextSequenceNum int
	privKey         crypto.PrivKey
	Txs             map[string][]models.Tx
	genesis         map[string]float64
	balances        map[string]float64
	TotalCoins      float64
}

func (n *Node) Init(privKey crypto.PrivKey, bootstraoPeers []string, genesis map[string]float64, host host.Host) {
	log.Print("Node starting")

	n.privKey = privKey
	n.genesis = genesis
	n.Txs = make(map[string][]models.Tx)
	n.balances = make(map[string]float64)

	if host == nil {
		n.host, _ = libp2p.New(libp2p.Identity(privKey))
	} else {
		n.host = host
	}

	go n.startTransactionServer()
	go n.startVerificationServer()

	for _, peer := range bootstraoPeers {
		txs, err := n.getTransactions(peer)

		if err != nil {
			log.Printf("Could not get transactions from %s %v", peer, err)
			continue
		}

		n.Txs = n.mergeTxs(n.Txs, txs)
	}

	n.calcBalances()
	n.TotalCoins = n.calcTotalCoins()
}

func (n Node) calcTotalCoins() float64 {
	var total float64
	for _, v := range n.genesis {
		total += v
	}

	return total
}
