package crypto

import (
	"fmt"
	"os"
	"testing"

	"github.com/ackhia/flash/models"
	"github.com/libp2p/go-libp2p/core/peer"
)

func TestReadWriteKeyFile(t *testing.T) {
	priv, pub := CreateKeyPair()
	keyFilename := "test_key.yaml"
	WriteKeyfile(keyFilename, priv, pub)
	privRead, pubRead, err := ReadKeyfile(keyFilename)

	if err != nil {
		t.Fatalf("Could not read %s %v", keyFilename, err)
	}

	if !priv.Equals(privRead) {
		t.Fatal("Private keys don't match")
	}

	if !pub.Equals(pubRead) {
		t.Fatal("Public keys don't match")
	}

	os.Remove("test_key.yaml")
}

func TestSignVerify(t *testing.T) {
	privSender, _ := CreateKeyPair()
	privVerifier, pubVerifier := CreateKeyPair()

	tx := models.Tx{
		From:   "Me",
		To:     "You",
		Amount: 25,
	}

	if SignTx(&tx, privSender) != nil {
		t.Fatal("Could not sign tx")
	}

	if len(tx.Sig) == 0 {
		t.Fatal("No sig present")
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
