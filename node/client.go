package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	fcrypto "github.com/ackhia/flash/crypto"
	"github.com/ackhia/flash/models"
	"github.com/ackhia/flash/transport"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func (n Node) getTransactions(addrInfo string) {
	serverAddr, err := peer.AddrInfoFromString(addrInfo)

	if err != nil {
		log.Fatalf("Invalid address %s %v", addrInfo, err)
	}

	n.host.Connect(context.Background(), *serverAddr)

	stream, err := n.host.NewStream(context.Background(), serverAddr.ID, protocol.ID("/flash/transactions/1.0.0"))

	if err != nil {
		log.Printf("Could not create stream %v", err)
		return
	}

	data, _ := io.ReadAll(stream)

	txs := make(map[string][]models.Tx)
	if err = json.Unmarshal(data, &txs); err != nil {
		log.Print("Could not unmarshal transactions")
		return
	}

	stream.Close()
}

func (n Node) fetchVerifications(tx *models.Tx) error {
	peers := n.host.Peerstore().Peers()

	for _, p := range peers {

		//Don't connect to myself
		if p == n.host.ID() {
			continue
		}

		err := n.getNodeVerification(tx, p)
		if err != nil {
			log.Printf("Error sending tx to peer %s: %v", p, err)
		} else {
			fmt.Printf("Message sent to peer %s\n", p)
		}
	}
	return nil
}

func (n Node) getNodeVerification(tx *models.Tx, p peer.ID) error {
	fmt.Printf("Connecting to %s", p)

	protocolID := protocol.ID(verifyTxProtocol)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := n.host.NewStream(ctx, p, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}

	defer stream.Close()

	msg, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("error marshalling struct to JSON: %v", err)
	}

	log.Print("Sending verification request")

	err = transport.SendBytes(msg, stream)
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	log.Print("Verification request sent")

	data, err := transport.ReceiveBytes(stream)
	if err != nil {
		return fmt.Errorf("failed to receive response: %v", err)
	}

	var verifier models.Verifier
	verifier.Sig = data
	verifier.ID = p.String()

	pubKey := n.host.Peerstore().PubKey(p)
	r, err := fcrypto.VerifyVerifier(&verifier, tx, pubKey, p)

	if err != nil {
		return fmt.Errorf("failed verify verifier: %v", err)
	}

	if !r {
		return fmt.Errorf("invalid sig")
	}

	tx.Verifiers = append(tx.Verifiers, verifier)

	return nil
}

func (n Node) VerifyTx(tx *models.Tx) error {

	err := n.fetchVerifications(tx)
	if err != nil {
		return err
	}

	n.Txs[n.host.ID().String()] = append(n.Txs[n.host.ID().String()], *tx)

	return nil
}

func (n Node) BuildTx(from string, to string, amount float64) (*models.Tx, error) {

	if amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}

	_, err := peer.Decode(from)
	if err != nil {
		return nil, fmt.Errorf("invalid From peer ID: %v", err)
	}

	_, err = peer.Decode(to)
	if err != nil {
		return nil, fmt.Errorf("invalid To peer ID: %v", err)
	}

	tx := models.Tx{
		SequenceNum: n.nextSequenceNum,
		From:        from,
		To:          to,
		Amount:      amount,
	}
	n.nextSequenceNum++

	return &tx, nil
}

func CommitTx(tx *models.Tx) {
	//Check verifiers and set Commited to true
}
