package node

import (
	"encoding/json"
	"log"

	fcrypto "github.com/ackhia/flash/crypto"
	"github.com/ackhia/flash/models"
	"github.com/ackhia/flash/transport"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n Node) startTransactionServer() {
	n.host.SetStreamHandler("/flash/transactions/1.0.0", func(s network.Stream) {
		defer s.Close()
		data, err := json.Marshal(n.Txs)
		if err != nil {
			log.Printf("could not marshal transactions")
			return
		}
		s.Write(data)
	})

	select {}
}

func (n Node) startVerificationServer() {
	n.host.SetStreamHandler(verifyTxProtocol, func(s network.Stream) {
		defer s.Close()

		log.Print("Client connected to verification server")

		data, err := transport.ReceiveBytes(s)
		if err != nil {
			log.Printf("Could not read tx %v", err)
			return
		}
		log.Print("Received verification request")

		var tx models.Tx
		err = json.Unmarshal(data, &tx)
		if err != nil {
			log.Printf("Could not unmarshall tx %v", err)
			return
		}

		bal, ok := n.balances[tx.From]
		if !ok || bal < tx.Amount {
			log.Printf("Balance too low for %s", tx.From)
			return
		}

		if len(n.Txs[tx.From]) != tx.SequenceNum {
			log.Printf("Invalid sequence number %d", tx.SequenceNum)
			return
		}

		if tx.Amount <= 0 {
			return
		}

		_, err = peer.Decode(tx.From)
		if err != nil {
			log.Printf("Invalid From peer ID: %v", err)

			return
		}

		_, err = peer.Decode(tx.To)
		if err != nil {
			log.Printf("Invalid To peer ID: %v", err)
			return
		}

		sig, err := fcrypto.CreateVerifyerSig(&tx, n.privKey)
		if err != nil {
			log.Printf("Could not sign tx %v", err)
			return
		}

		tx.Comitted = false
		n.Txs[tx.From] = append(n.Txs[tx.From], tx)

		err = transport.SendBytes(sig, s)
		if err != nil {
			log.Printf("Could not send bytes %v", err)
		}
	})

	select {}
}
