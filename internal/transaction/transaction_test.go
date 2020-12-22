package transaction

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/alpdeniz/themachine/internal/keystore"
)

var (
	tx      *Transaction
	txBytes []byte
	testOrg = Organization{
		Name:                            "Test Org",
		Description:                     "Test Desc",
		MasterPublicKey:                 []byte{},
		MinimumRequiredSignaturePaths:   []string{"1/*+"},
		RequiredSignaturePathsPerObject: map[string][]string{},
		Rules:                           []string{"Rule1", "Rule2"},
	}
)

func TestBuild(t *testing.T) {

	organizationTx := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	exampleHash, _ := hex.DecodeString("337349113a5c51e8bc96c9c7182fdb16fd3aec7f292ce5e1847ef0f70530bb2d")
	exampleBytes := []byte{0, 0, 0, 0, 170, 0, 0, 0, 123, 34, 78, 97, 109, 101, 34, 58, 34, 84, 101, 115, 116, 32, 79, 114, 103, 34, 44, 34, 68, 101, 115, 99, 114, 105, 112, 116, 105, 111, 110, 34, 58, 34, 84, 101, 115, 116, 32, 68, 101, 115, 99, 34, 44, 34, 77, 97, 115, 116, 101, 114, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 34, 44, 34, 77, 105, 110, 105, 109, 117, 109, 82, 101, 113, 117, 105, 114, 101, 100, 83, 105, 103, 110, 97, 116, 117, 114, 101, 80, 97, 116, 104, 115, 34, 58, 91, 34, 49, 47, 42, 43, 34, 93, 44, 34, 82, 101, 113, 117, 105, 114, 101, 100, 83, 105, 103, 110, 97, 116, 117, 114, 101, 80, 97, 116, 104, 115, 80, 101, 114, 79, 98, 106, 101, 99, 116, 34, 58, 123, 125, 44, 34, 82, 117, 108, 101, 115, 34, 58, 91, 34, 82, 117, 108, 101, 49, 34, 44, 34, 82, 117, 108, 101, 50, 34, 93, 125, 8, 0, 109, 47, 49, 39, 47, 49, 58, 53}
	targets := []string{"m/1'/1:5"} // e.g 5 signatures are enough for approval
	var err error = nil

	foundation, err := json.Marshal(testOrg)
	if err != nil {
		t.Error("Cannot marshal organization object")
	}
	tx, err = Build(Genesis, "json", organizationTx, foundation, targets)
	if err != nil {
		t.Error("Cannot build genesis transaction")
	}

	hash := tx.CalculateHash()
	txBytes = tx.ToBytes()

	if !bytes.Equal(txBytes, exampleBytes) {
		t.Error("Transaction bytes are not correct")
	}

	if !bytes.Equal(hash, exampleHash) {
		t.Error("Transaction hash is not correct")
	}

}

func TestParseBytes(t *testing.T) {

	tx2, err := ParseBytes(txBytes)
	if err != nil {
		t.Error("Error parsing tx byte hex", err)
	}

	if len(tx2.Targets) != len(tx.Targets) {
		t.Error("Parsed tx targets are not correct: ", tx.Targets, tx2.Targets)
	}

	if !bytes.Equal(tx2.Hash, tx.Hash) {
		t.Error("Parsed tx message is not correct, message", string(tx.Data), string(tx2.Data))
	}

}

func TestSign(t *testing.T) {
	keystore.Open()
	keypair := keystore.GetKeyPairByName("Node")
	if keypair == nil {
		t.Error("Could not find Keypair")
		return
	}

	tx.Sign(*keypair)

	fmt.Println(tx.Signatures, tx.DerivationPaths)

	if len(tx.Signatures) == 0 {
		t.Error("Could not sign the transaction")
	}

	ok := tx.VerifySignatureByIndex(0)
	if !ok {
		t.Error("Signature is not valid")
	}
}

func TestSave(t *testing.T) {

	// Currently
	tx2, err := Process(tx.ToBytes())
	if err != nil {
		t.Error("Error while processing transaction")
	}

	tx3 := Retrieve(tx.Hash)
	if tx3 == nil {
		t.Error("Transaction does not exist in db")
	}

	if !bytes.Equal(tx2.Hash, tx3.Hash) {
		t.Error("Transaction hashes are not equal", hex.EncodeToString(tx.Hash), hex.EncodeToString(tx2.Hash))
	}

}
