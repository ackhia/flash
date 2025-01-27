package node

import (
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"

	fcrypto "github.com/ackhia/flash/crypto"
	"github.com/ackhia/flash/models"
)

const verifyTxProtocol = "/flash/verify-transaction/1.0.0"
const commitTxProtocol = "/flash/commit-transaction/1.0.0"

type Node struct {
	Host            host.Host
	nextSequenceNum int
	privKey         crypto.PrivKey
	Txs             map[string][]models.Tx
	genesis         map[string]float64
	Balances        map[string]float64
	TotalCoins      float64
	bootstraoPeers  []string
}

func New(privKey crypto.PrivKey, host *host.Host, genesis map[string]float64, bootstraoPeers []string) *Node {
	n := Node{
		privKey:        privKey,
		Txs:            make(map[string][]models.Tx),
		genesis:        genesis,
		Balances:       make(map[string]float64),
		bootstraoPeers: bootstraoPeers,
	}

	if host == nil {
		n.Host, _ = libp2p.New(libp2p.Identity(privKey))
	} else {
		n.Host = *host
	}

	return &n
}

func (n *Node) Start() {
	log.Print("Node starting")
	go n.startTransactionServer()
	go n.startVerificationServer()
	go n.startCommitTxServer()

	for _, peer := range n.bootstraoPeers {
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

func (n *Node) Transfer(to string, amount float64) error {
	pubKeyBytes, err := crypto.MarshalPublicKey(n.privKey.GetPublic())
	if err != nil {
		return err
	}

	tx, err := n.BuildTx(
		n.Host.ID().String(),
		to,
		amount,
		pubKeyBytes,
	)

	if err != nil {
		return fmt.Errorf("could not build tx: %v", err)
	}

	err = fcrypto.SignTx(tx, n.privKey)
	if err != nil {
		return fmt.Errorf("could not sign tx: %v", err)
	}

	err = n.VerifyTx(tx)
	if err != nil {
		return fmt.Errorf("could not send tx: %v", err)
	}

	n.CommitTx(tx)

	return nil
}
