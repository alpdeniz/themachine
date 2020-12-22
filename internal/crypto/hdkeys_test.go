package crypto

import (
	"bytes"
	"testing"
)

func TestGenerate(t *testing.T) {
	_, err := NewWallet()
	if err != nil {
		t.Error("Error generating master key", err)
	}
}

func TestCKD(t *testing.T) {
	master, err := WalletFromExtendedKey("xprv9s21ZrQH143K3n3bx3AGikGfb8ctEHUvQxs7p1Xc9k57aFcaHEASDUCMyWG66aFfavRQKkRotD3b69RVND6kgu5TEaPAy4gydrFsiHPtpnH") // NewWallet()
	if err != nil {
		t.Error("Error creating new wallet", err)
	}
	// get other key from neutered parent
	neuteredMaster := master.Pub()

	// derive from private
	child, err := master.Child(124)
	if err != nil {
		t.Error("Error deriving child", err)
	}
	// // derive from public
	neuteredChild, err := neuteredMaster.Child(124)
	if err != nil {
		t.Error("Error deriving child from neutered key", err)
	}

	// Check if children's public keys match
	if !bytes.Equal(child.Pub().Key, neuteredChild.Pub().Key) {
		t.Error("Error matching children of private and public master")
	}
}

func TestChildSignatures(t *testing.T) {

	master, err := NewWallet()
	child, err := master.Child(1)

	message := []byte("Message")
	hash := DHash(message)

	privKey := child.Key
	pubKey := child.Pub().Key

	signature, err := Sign(hash, privKey)
	if err != nil {
		t.Error("Error signing message", err)
	}

	ok := Verify(signature, hash, pubKey)
	if !ok {
		t.Error("Error verifing message", err)
	}
}

func TestPaths(t *testing.T) {

	keyPath := ParseDerivationPathString("1/5")
	targetPath := ParseDerivationPathString("1/*")
	ok := IsPathUnderPath(keyPath, targetPath)
	if !ok {
		t.Error("Given path should be covered by target path")
	}

	master, err := NewWallet()
	if err != nil {
		t.Error("Could not generate master key", err)
	}

	publicMaster := master.Pub()
	childPublicMaster := DeriveFromMPK(keyPath, publicMaster.String())

	ok = CheckPublicKeyPath(keyPath, childPublicMaster.Key, publicMaster.String())
	if !ok {
		t.Error("Derived public key is not in given path relative to the master")
	}
}
