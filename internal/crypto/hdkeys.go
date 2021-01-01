package crypto

// Hierarchical Deterministic Wallet utils (Key derivation functions)
// required for proof of authority via signatures
// Decision of signature algorithm affect the functions in this file

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/wemeetagain/go-hdwallet"
	// "go.dedis.ch/kyber/v3/pairing"
	// "go.dedis.ch/kyber/v3/sign/bdn"
	// "go.dedis.ch/kyber/v3/suites"
)

// Generate a new master key
func NewWallet() (*hdwallet.HDWallet, error) {
	randomSeed, err := hdwallet.GenSeed(256)
	if err != nil {
		return nil, err
	}

	return hdwallet.MasterKey(randomSeed), nil
}

// Build wallet from extended key
func WalletFromExtendedKey(extendedKey string) (*hdwallet.HDWallet, error) {
	return hdwallet.StringWallet(extendedKey)
}

// Corresponding address
func PublicKeyToAddress(key []byte) []byte {
	return Hash160(key)
}

// Parses derivation paths of form "i/j/k/l"
func ParseDerivationPathString(path string) []uint32 {
	var result []uint32
	var childIndex uint32
	elems := strings.Split(path, "/")
	for _, v := range elems {
		if len(v) == 0 {
			// decide what to do with "" or "0/1/" (ending with "/")
			continue
		}
		if v != "*" {
			childIndex = uint32(v[0])
			if len(v) == 2 {
				// hardened key
				childIndex += 0x80000000
			}
			result = append(result, childIndex)
		} else {
			result = append(result, 0)
		}
	}

	return result
}

func ParseDerivationPathBytes(path []byte) []uint32 {
	if len(path) != 16 {
		return nil
	}

	var result []uint32
	var childIndex uint32
	for i := 0; i < 4; i++ {
		childIndex = binary.BigEndian.Uint32(path[i*4 : (i+1)*4])
		result = append(result, childIndex)
	}
	return result
}

func DerivationPathToBytes(path []uint32) []byte {
	var result []byte
	var tmp = make([]byte, 4)
	for _, v := range path {
		binary.BigEndian.PutUint32(tmp, v)
		result = append(result, tmp...)
	}
	return result
}

// Determine if exact path is within a target path (see transaction.Transaction.Targets)
func IsPathUnderPath(exactPathSteps []uint32, targetPath []uint32) bool {
	// exactPathSteps := ParseDerivationPath(exactPath)
	for i, v := range targetPath {
		// it is a higer path
		if len(exactPathSteps) == i {
			return false
		}
		if v != 0 {
			if exactPathSteps[i] != v {
				return false
			}
		} else {
			return true
		}
	}
	return false
}

// Checks if given path is a correct derivation path for a master public key (of an organization)
func CheckPublicKeyPath(path []uint32, publicKey []byte, masterPublicKey string) bool {
	// extract derivation path from extended public key
	derivedKey := DeriveFromMPK(path, masterPublicKey)
	// return true if correct
	return bytes.Equal(derivedKey.Pub().Key, publicKey)
}

// Given path derive a child
func DeriveFromMPK(derivationSteps []uint32, masterPublicKey string) hdwallet.HDWallet {
	w, err := hdwallet.StringWallet(masterPublicKey)
	if err != nil {
		fmt.Println("Error reading master public key")
	}

	for i, v := range derivationSteps {
		// Stop if child index is 0
		if v == 0 {
			return *w
		}
		// Derive child
		w, err = w.Child(uint32(v))
		if err != nil {
			fmt.Printf("Error deriving child key %d at depth %d from master public key\n", v, i)
		}
	}
	return *w
}
