package crypto

import (
	"os"
	"testing"
)

func TestReadWriteKeyFile(t *testing.T) {
	priv, pub := CreateKeyPair()
	keyFilename := "test_key.yaml"
	WriteKeyfile(keyFilename, priv, pub)
	privRead, pubRead, err := ReadKeyfile(keyFilename)

	if err != nil {
		t.Fatalf("Could not read %s %s", keyFilename, err)
	}

	if !priv.Equals(privRead) {
		t.Fatal("Private keys don't match")
	}

	if !pub.Equals(pubRead) {
		t.Fatal("Public keys don't match")
	}

	os.Remove("test_key.yaml")
}
