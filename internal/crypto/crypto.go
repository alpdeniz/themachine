package crypto

// Provides fundamental cryptograpghic functions
// - Hash    sha256(x)
// - DHash   sha256(x)^2
// - Hash160 ripemd160(sha256(x))
// - Pbkdf2  HMAC-SHA256(x)^n
// - Encrypt AES-GCM
// - Decrypt AES-GCM
// - Sign    ECC - secp256k1
// - Verify  ECC - secp256k1

// For encrypt/decrypt: AES-GCM or AES-CBC or ...?
// For sign/verify: ECC or PBC (Pairing Based Cryptography e.g. BLS)
// - ECC requires n signatures for n signees
// - BLS requires 1 signature for n signees
// - Study & Discuss Schnorr
// Note: Signature method affects HD key functions (BLS does not allow child derivation from master public key. Needs fact check)

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ripemd160"
)

func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func DHash(data []byte) []byte {
	return Hash(Hash(data))
}

func Hash160(data []byte) []byte {
	rp := ripemd160.New()
	rp.Write(Hash(data))
	return rp.Sum(nil)
}

func Pbkdf2(key []byte, salt []byte, kdfIterations int, kdfKeyLength int) []byte {
	return pbkdf2.Key(key, salt, kdfIterations, kdfKeyLength, sha256.New)
}

func Encrypt(message []byte, encryptionKey []byte) []byte {

	// setup GCM
	aesGCM, err := setupGCM(encryptionKey)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// set random nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println(err)
		return nil
	}

	return aesGCM.Seal(nonce, nonce, message, nil)
}

func Decrypt(data []byte, encryptionKey []byte) ([]byte, error) {

	// setup GCM
	aesGCM, err := setupGCM(encryptionKey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// separate nonce
	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

func setupGCM(key []byte) (cipher.AEAD, error) {
	// get block size
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// setup GCM cipher
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return aesGCM, nil
}

func Sign(message []byte, key []byte) ([]byte, error) {
	// in case it has a leading zero byte
	if len(key) == 33 {
		key = key[1:]
	}

	signature, err := secp256k1.Sign(message, key)
	if err != nil {
		return nil, err
	}
	// remove the last byte (used for public key recovery)
	if len(signature) == 65 {
		signature = signature[:64]
	}
	return signature, err
}

func Verify(signature []byte, message []byte, publicKey []byte) bool {
	return secp256k1.VerifySignature(publicKey, message, signature)
}
