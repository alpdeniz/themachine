package transaction

import (
	"encoding/hex"
	"fmt"

	"github.com/alpdeniz/themachine/internal/crypto"
	"github.com/alpdeniz/themachine/internal/keystore"
)

// Process applies the procedure to all incoming transactions
func Process(transactionBytes []byte) (*Transaction, error) {

	// extract first
	tx, err := ParseBytes(transactionBytes)
	if err != nil {
		return nil, err
	}

	// Validate
	ok, err := tx.Validate()
	if !ok || err != nil {
		fmt.Println("Invalid transaction", hex.EncodeToString(tx.Hash), err)
		return nil, err
	}

	// // Check to see if it is related to this node, if yes, save into related
	processRelated(tx)

	// // Validate
	// ok, err = Verify(tx)
	// if !ok || err != nil {
	// 	fmt.Println("Valid but unverified transaction", hex.EncodeToString(tx.Hash), err)
	// // Not saving but relaying
	// 	return tx, err
	// }

	// Validated
	tx.Save()

	return tx, nil
}

// check if transaction is valid
func (tx *Transaction) Validate() (bool, error) {

	// ok, err := tx.CheckInitialSignature()
	// if !ok {
	// 	return false, err
	// }

	// check hash, check rules etc.
	return true, nil
}

// check if transaction is verified
func (tx *Transaction) Verify() (bool, error) {

	// check if signed by required number of signers
	numberOfSignatures := tx.CheckSignatures()
	if tx.IsVerified(numberOfSignatures) {
		return true, nil
	}

	return false, nil
}

func processRelated(tx *Transaction) {
	// Find out if asks for our signature
	for _, target := range tx.Targets {
		for _, keypair := range keystore.CurrentKeyMap {
			// check if our key is eligible to sign
			if crypto.IsPathUnderPath(crypto.ParseDerivationPathString(keypair.DerivationPath), crypto.ParseDerivationPathString(target)) {
				// then it is of importance to this node
				// the app should inform the user of this node to read and to sign or not
				tx.SaveRelated()
			}
		}
	}
}
