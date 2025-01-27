package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/ackhia/flash/models"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func CreateKeyPair() (crypto.PrivKey, crypto.PubKey) {
	privKey, publicKey, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		log.Fatalf("Error generating keys: %v", err)
	}

	return privKey, publicKey
}

func WritePrivateKey(filename string, privKey crypto.PrivKey) error {
	// Serialize the private key to bytes
	keyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Encode the key bytes to base64
	base64Key := base64.StdEncoding.EncodeToString(keyBytes)

	// Write the base64-encoded key to the file
	err = os.WriteFile(filename, []byte(base64Key), 0600) // 0600 ensures file security
	if err != nil {
		return fmt.Errorf("failed to write private key to file: %w", err)
	}

	return nil
}

// readPrivateKey reads a libp2p crypto.PrivKey from a human-readable text file.
func ReadPrivateKey(filename string) (crypto.PrivKey, error) {
	// Read the file contents
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Decode the base64-encoded key
	keyBytes, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 private key: %w", err)
	}

	// Unmarshal the private key from bytes
	privKey, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}

	return privKey, nil
}

func hashTx(tx *models.Tx) []byte {
	data := fmt.Sprintf("%d%s%s%f%X", tx.SequenceNum, tx.From, tx.To, tx.Amount, tx.Pubkey)
	hash := sha256.Sum256([]byte(data))

	return hash[:]
}

func hashTxWithSig(tx *models.Tx) []byte {
	data := fmt.Sprintf("%d%s%s%f%x%x", tx.SequenceNum, tx.From, tx.To, tx.Amount, tx.Sig, tx.Pubkey)
	hash := sha256.Sum256([]byte(data))

	return hash[:]
}

func SignTx(tx *models.Tx, privKey crypto.PrivKey) error {
	hash := hashTx(tx)

	if sig, err := privKey.Sign(hash); err != nil {
		return err
	} else {
		tx.Sig = sig
	}

	return nil
}

func VerifyTxSig(tx models.Tx) (bool, error) {
	//TODO: Check the from is derived from the pubkey

	hash := hashTx(&tx)

	pubKey, err := crypto.UnmarshalPublicKey(tx.Pubkey)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal public key: %v", err)
	}

	result, err := pubKey.Verify(hash, tx.Sig)

	if err != nil {
		return false, err
	}

	return result, nil
}

func VerifyVerifier(verifier *models.Verifier, tx *models.Tx, pubKey crypto.PubKey, p peer.ID) (bool, error) {
	if verifier.ID != p.String() {
		return false, nil
	}

	hash := hashTxWithSig(tx)
	return pubKey.Verify(hash, verifier.Sig)
}

func CreateVerifyerSig(tx *models.Tx, privKey crypto.PrivKey) ([]byte, error) {
	hash := hashTxWithSig(tx)
	sig, err := privKey.Sign(hash)
	if err != nil {
		return nil, err
	}

	return sig, nil
}
