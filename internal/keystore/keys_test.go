package keystore

import (
	"fmt"
	"testing"

	"github.com/alpdeniz/themachine/internal/crypto"
)

func TestGenerateKey(t *testing.T) {

	Open()
	kp := NewKeyPair("test")
	if kp == nil {
		t.Error("Cannot generate new key pair")
	}

}

func TestKeyPair(t *testing.T) {

	kp := GetKeyPairByName("test")
	if kp == nil {
		t.Error("Cannot get previously generated key pair")
	}

	fmt.Println("Got freshly generated key pair: ", kp)

	ok := IsOwnedAddress([]byte(kp.Address))
	if !ok {
		t.Error("Owned key pair is not found by address")
	}

}

func TestSignAndVerify(t *testing.T) {
	kp := GetKeyPairByName("test")

	message := []byte("HELLO")
	hash := crypto.DHash(message)
	sig, err := crypto.Sign(hash, kp.PrivateKey)
	if err != nil {
		t.Error("Cannot sign message")
	}

	ok := crypto.Verify(sig, hash, kp.PublicKey)
	if !ok {
		t.Error("Signature is not valid")
	}
}
