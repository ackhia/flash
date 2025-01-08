package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"gopkg.in/yaml.v3"
)

type KeyPair struct {
	PrivKey string `yaml:"private_key"`
	PubKey  string `yaml:"public_key"`
}

func CreateKeyPair() (crypto.PrivKey, crypto.PubKey) {
	privKey, publicKey, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		log.Fatalf("Error generating keys: %v", err)
	}

	return privKey, publicKey
}

func writeKeyPairToYAML(filePath string, keyPair KeyPair) error {
	data, err := yaml.Marshal(&keyPair)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func WriteKeyfile(filename string, priv crypto.PrivKey, pub crypto.PubKey) {

	privKeyBytes, err := priv.Raw()
	if err != nil {
		log.Fatalf("Error converting private key to bytes: %v", err)
	}

	// Convert public key to a byte slice
	pubKeyBytes, err := pub.Raw()
	if err != nil {
		log.Fatalf("Error converting public key to bytes: %v", err)
	}

	privKeyStr := base64.StdEncoding.EncodeToString(privKeyBytes)
	pubKeyStr := base64.StdEncoding.EncodeToString(pubKeyBytes)

	// Write the key pair to a YAML file
	err = writeKeyPairToYAML(filename, KeyPair{privKeyStr, pubKeyStr})
	if err != nil {
		log.Fatalf("Error writing to YAML file: %s", err)
	}
}

func readKeyPairFromYAML(filePath string) (KeyPair, error) {
	var keyPair KeyPair
	data, err := os.ReadFile(filePath)
	if err != nil {
		return keyPair, err
	}

	err = yaml.Unmarshal(data, &keyPair)
	if err != nil {
		return keyPair, err
	}

	return keyPair, nil
}

func ReadKeyfile(filename string) (crypto.PrivKey, crypto.PubKey, error) {
	// Read the key pair from the YAML file
	keyPair, err := readKeyPairFromYAML(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading YAML file: %w", err)
	}

	fmt.Printf("%s - %s", keyPair.PrivKey, keyPair.PubKey)

	// Decode the private key from base64
	privKeyBytes, err := base64.StdEncoding.DecodeString(keyPair.PrivKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding private key: %w", err)
	}

	// Decode the public key from base64
	pubKeyBytes, err := base64.StdEncoding.DecodeString(keyPair.PubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding public key: %w", err)
	}

	// Convert the bytes back into a crypto.PrivKey and crypto.PubKey
	privKey, err := crypto.UnmarshalRsaPrivateKey(privKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling private key: %w", err)
	}

	pubKey, err := crypto.UnmarshalRsaPublicKey(pubKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling public key: %w", err)
	}

	return privKey, pubKey, nil
}
