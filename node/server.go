package node

import (
	"bytes"
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

func (n *Node) startVerificationServer() {
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

		bal, ok := n.Balances[tx.From]
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

		result, err := fcrypto.VerifyTxSig(tx)
		if err != nil {
			log.Printf("Could not verify tx sig %v", err)
			return
		}

		if !result {
			log.Print("Tx has invalid sig")
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

func (n *Node) startCommitTxServer() {
	n.host.SetStreamHandler(commitTxProtocol, func(s network.Stream) {

		defer s.Close()

		log.Print("Client connected to commit server")

		data, err := transport.ReceiveBytes(s)
		if err != nil {
			log.Printf("Could not read tx %v", err)
			return
		}
		log.Print("Received commit request")

		var tx models.Tx
		err = json.Unmarshal(data, &tx)
		if err != nil {
			log.Printf("Could not unmarshall tx %v", err)
			return
		}

		var localTx *models.Tx
		for i := range n.Txs[tx.From] {
			t := &n.Txs[tx.From][i]
			if bytes.Compare(t.Sig, tx.Sig) == 0 {
				localTx = t
				break
			}
		}

		if localTx == nil {
			log.Print("Could not find local tx")
			return
		}

		if localTx.Amount != tx.Amount ||
			localTx.From != tx.From ||
			bytes.Compare(localTx.Pubkey, tx.Pubkey) != 0 ||
			localTx.To != tx.To ||
			bytes.Compare(localTx.Sig, tx.Sig) != 0 ||
			localTx.Comitted ||
			tx.Comitted ||
			localTx.SequenceNum != tx.SequenceNum {
			log.Print("tx dose not match database")
			return
		}

		for _, v := range tx.Verifiers {
			peerID, err := peer.Decode(v.ID)
			if err != nil {
				log.Printf("Error decoding Peer ID %v", err)
				return
			}

			result, err := fcrypto.VerifyVerifier(&v, &tx, n.host.Network().Peerstore().PubKey(peerID), peerID)

			if err != nil {
				log.Printf("Verifier not valid %v", err)
				return
			}

			if !result {
				log.Printf("Verifier not valid")
				return
			}
		}

		_, err = n.isVerifierConsensus(&tx)
		if err != nil {
			log.Print("Consensus could not be reached")
			return
		}

		localTx.Verifiers = tx.Verifiers
		localTx.Comitted = true
		n.calcBalances()

		transport.SendBytes([]byte("ok"), s)
	})

	select {}
}
