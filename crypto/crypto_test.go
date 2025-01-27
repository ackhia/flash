package crypto

import (
	"fmt"
	"os"
	"testing"

	"github.com/ackhia/flash/models"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func TestReadWriteKeyFile(t *testing.T) {
	priv, _ := CreateKeyPair()
	keyFilename := "test_key.yaml"
	WritePrivateKey(keyFilename, priv)
	privRead, err := ReadPrivateKey(keyFilename)

	if err != nil {
		t.Fatalf("Could not read %s %v", keyFilename, err)
	}

	if !priv.Equals(privRead) {
		t.Fatal("Private keys don't match")
	}

	os.Remove("test_key.yaml")
}

func TestSignVerify(t *testing.T) {
	privSender, pubSender := CreateKeyPair()
	privVerifier, pubVerifier := CreateKeyPair()

	pubKeyBytes, err := crypto.MarshalPublicKey(pubSender)
	if err != nil {
		t.Fatal("Could not get public key")
	}

	tx := models.Tx{
		From:   "Me",
		To:     "You",
		Amount: 25,
		Pubkey: pubKeyBytes,
	}

	if SignTx(&tx, privSender) != nil {
		t.Fatal("Could not sign tx")
	}

	if len(tx.Sig) == 0 {
		t.Fatal("No sig present")
	}

	result, err := VerifyTxSig(tx)

	if err != nil {
		t.Fatal("Could not verify sig")
	}

	if !result {
		t.Fatal("Sig verify failed")
	}

	sig, err := CreateVerifyerSig(&tx, privVerifier)
	if err != nil {
		t.Fatal("Could not verify tx")
	}

	peerID, err := peer.IDFromPrivateKey(privVerifier)
	if err != nil {
		fmt.Println("Error deriving peer ID:", err)
		return
	}

	ver := models.Verifier{
		ID:  peerID.String(),
		Sig: sig,
	}

	ok, err := VerifyVerifier(&ver, &tx, pubVerifier, peerID)
	if err != nil {
		t.Fatal("Verify failed to calculate")
	}

	if !ok {
		t.Fatal("Verify failed")
	}
}
