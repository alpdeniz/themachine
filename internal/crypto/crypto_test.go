package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestHashes(t *testing.T) {
	input := []byte("covid-19")
	sha2 := Hash(input)
	doubleSha2 := DHash(input)

	if hex.EncodeToString(sha2) != "88529c3ac8ebd2dcb21a432c4ea0190c8370850b73ef95a527d150d4d424bc62" {
		t.Error("Error sha2")
	}

	if hex.EncodeToString(doubleSha2) != "0babbdf4d1c0a701d12aa79fe1564f81b2d29b6ab60d5a571e9c916f0796a009" {
		t.Error("Error double sha2")
	}

}

func TestEncryption(t *testing.T) {
	message := []byte("This should not be compromised")

	password := []byte("password")
	salt := []byte("salt")
	key := Pbkdf2(password, salt, 1000, 32)

	cyphertext := Encrypt(message, key)
	plaintext, err := Decrypt(cyphertext, key)
	if err != nil {
		t.Error("Error decrypting cyphertext", err)
	}

	if !bytes.Equal(message, plaintext) {
		t.Error("Resulting bytes after decryption does not match")
	}
}
