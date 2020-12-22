package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/alpdeniz/themachine/internal/crypto"
	"github.com/alpdeniz/themachine/internal/keystore"
)

func (tx *Transaction) GetHashedBytes() []byte {
	var totalLen int = len(tx.Meta) + len(tx.OrganizationTx) + len(tx.Data) + len(strings.Join(tx.Targets, ","))
	tmp := make([]byte, totalLen)
	slices := [4][]byte{tx.Meta[:], tx.OrganizationTx, tx.Data, []byte(strings.Join(tx.Targets, ","))}
	var i int
	for _, s := range slices {
		i += copy(tmp[i:], s)
	}

	return tmp
}

func (tx *Transaction) CalculateHash() []byte {

	authMessage := tx.GetHashedBytes()
	tx.Hash = crypto.DHash(authMessage)
	return tx.Hash
}

func (tx *Transaction) CheckInitialSignature() (bool, error) {
	if len(tx.PublicKeys) == 0 || len(tx.Signatures) == 0 || len(tx.DerivationPaths) == 0 {
		return false, errors.New("Transaction is not signed")
	}
	ok := tx.VerifySignatureByIndex(0)
	return ok, nil
}

func (tx *Transaction) CheckSignatures() int {

	signatureCounter := 0
	for i, v := range tx.PublicKeys {
		// skip first signature
		if i == 0 {
			continue
		}

		// check signature
		ok := tx.VerifySignatureByIndex(i)
		if !ok {
			fmt.Println("Invalid signature in transaction", hex.EncodeToString(tx.Hash))
			continue
		}
		// check if provided path is correct for organization master key
		if crypto.CheckPublicKeyPath(tx.DerivationSteps[i], v, hex.EncodeToString(tx.Organization.MasterPublicKey)) {
			// go over targets
			for _, v2 := range tx.Targets {
				// check if it belongs to one of the target paths
				if crypto.IsPathUnderPath(tx.DerivationSteps[i], crypto.ParseDerivationPathString(v2)) {
					signatureCounter++
				}
			}
		} // given derivation path to the public key does not correspond
	}
	return signatureCounter
}

func (tx *Transaction) IsVerified(signatureCounter int) bool {
	sum := 0
	for range tx.MinimumRequiredSignatures {
		sum++
	}
	return sum == signatureCounter
}

// Function to call when the key is ready and transaction is approved
func (tx *Transaction) Sign(keypair keystore.KeyPair) {
	// This signing should be done by people
	fmt.Println("Signing transaction with", keypair)
	sig, err := crypto.Sign(tx.Hash, keypair.PrivateKey)
	if err != nil {
		fmt.Println("Error signing transaction", err)
	}
	tx.Signatures = append(tx.Signatures, sig)
	tx.PublicKeys = append(tx.PublicKeys, keypair.PublicKey)
	var tmpBytes = make([]byte, 4)
	var derivationPathBytes = make([]byte, 16)
	for i, v := range crypto.ParseDerivationPathString(keypair.DerivationPath) {
		binary.BigEndian.PutUint32(tmpBytes, v)
		copy(derivationPathBytes[i*4:(i+1)*4], tmpBytes)
	}
	tx.DerivationPaths = append(tx.DerivationPaths, derivationPathBytes)
}

func (tx *Transaction) VerifySignatureByIndex(index int) bool {
	return crypto.Verify(tx.Signatures[index], tx.Hash, tx.PublicKeys[index])
}
