package db

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gobuffalo/packr/v2/file/resolver/encoding/hex"
)

func TestKeys(t *testing.T) {

	count := CountNumberOfKeys()
	if count <= 0 {
		t.Error("Cannot count keys")
	}

	kps := GetKeyPairs()
	if int64(len(kps)) != count {
		t.Error("Key counts are not equal")
	}
	fmt.Println("Number of keys: ", count)

	rootKeys := GetsKeyByDerivationPath("0")
	if len(rootKeys) == 0 {
		t.Error("No root keys found")
	}

	key := GetKeyByName("Node")
	if key.DerivationPath != "0" {
		t.Error("Node key derivation path is not 0", key)
	}
}

func TestTransactions(t *testing.T) {

	count := CountNumberOfTransactions()
	if count <= 0 {
		t.Error("Cannot count transactions")
	}
	fmt.Println("Number of transactions: ", count)

	txHashBytes, _ := hex.DecodeString("2b3fb3187a38410fa07b7d564d8e8dabc0d0494d748d4b38a94e82ebfe83d923")
	tx := Get(txHashBytes)
	if !bytes.Equal(tx.Hash, txHashBytes) {
		t.Error("Tx hash does not match")
	}

	txs := GetByObjectType(0x00)
	if len(txs) == 0 {
		t.Error("Could not find any genesis transactions")
	}
}
