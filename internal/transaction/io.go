package transaction

// Un/Serialization functions of type Transaction

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alpdeniz/themachine/internal/crypto"
	"github.com/alpdeniz/themachine/internal/db"
)

// Initialize a transaction for the first time
func Build(objectType ObjectType, subType string, organizationTx []byte, message []byte, targets []string) (*Transaction, error) {

	// Set data
	tx := Transaction{}
	tx.ObjectType = objectType
	tx.Meta = [4]byte{byte(objectType), byte(0x00), byte(0x00), byte(0x00)}
	tx.SubType = ObjectSubType(subType)
	tx.Data = message
	tx.DataLength = uint32(len(message))
	tx.Targets = targets

	// build organization
	if objectType != Genesis {
		// check hash length
		if len(organizationTx) < 32 {
			return nil, errors.New("Invalid organization transaction hash")
		}
		// fetch if exists
		organizationTransaction := Retrieve(organizationTx)
		if organizationTransaction == nil {
			return nil, errors.New("Cannot find organization transaction")
		}
		// parse
		organization, err := ParseOrganizationData(organizationTransaction.Data)
		if err != nil {
			return nil, errors.New("Cannot parse organization data")
		}
		// finally set in Transaction
		tx.Organization = *organization
		tx.OrganizationTx = organizationTx
	}

	tx.CalculateHash()
	return &tx, nil
}

// Export transaction as a db object
func (tx *Transaction) ToDBItem() db.MainDBItem {
	item := db.MainDBItem{
		Hash:                    tx.Hash,
		PrevHash:                nil,
		Date:                    time.Now(),
		ObjectType:              byte(tx.ObjectType),
		SubType:                 0x00,
		Data:                    tx.Data,
		Targets:                 tx.Targets,
		Signatures:              tx.Signatures,
		PublicKeys:              tx.PublicKeys,
		DerivationPaths:         tx.DerivationPaths,
		OrganizationTransaction: tx.OrganizationTx,
	}
	return item
}

// Construct a transaction from a db object
func FromDBItem(item db.MainDBItem) (*Transaction, error) {
	tx := Transaction{}
	tx.Hash = item.Hash
	tx.Data = item.Data
	tx.DataLength = uint32(len(item.Data))
	tx.ObjectType = ObjectType(item.ObjectType)
	tx.SubType = ObjectSubType(item.SubType)
	tx.Date = item.Date
	tx.OrganizationTx = item.OrganizationTransaction
	tx.Targets = item.Targets
	tx.PublicKeys = item.PublicKeys
	tx.Signatures = item.Signatures
	tx.DerivationPaths = item.DerivationPaths
	tx.Meta = [4]byte{byte(item.ObjectType), byte(0x00), byte(0x00), byte(0x00)}

	// build organization
	var orgData []byte
	if tx.ObjectType == Genesis {

		// json data related to the organization
		orgData = tx.Data

	} else {
		// check hash length
		if len(tx.OrganizationTx) < 32 {
			return nil, errors.New("Invalid organization transaction hash")
		}
		// fetch if exists
		organizationTransaction := Retrieve(tx.OrganizationTx)
		if organizationTransaction == nil {
			return nil, errors.New("Cannot find organization transaction")
		}
		// json data related to the organization
		orgData = organizationTransaction.Data
	}

	// parse
	organization, err := ParseOrganizationData(orgData)
	if err != nil {
		return nil, errors.New("Cannot parse organization data")
	}
	// finally set in Transaction
	tx.Organization = *organization

	return &tx, nil
}

// Export transaction as bytes, ready to relay into the network
func (tx *Transaction) ToBytes() []byte {

	// total length of byte array
	// Genesis: 4 + 4 + mLength + 2 + targetLength + signatures
	// Organization Transaction:  4 + 32 + 4 + mLength + 2 + targetLength + signatures (33 pk + 64 sig + 16 derivation path (4 * uint32) )
	// Message to Network: 4 + mLength ?
	metaLength, txHashLength, messageLenBytes, targetLenBytes := 4, 0, 4, 2
	var transactionBytes []byte
	transactionBytes = append(transactionBytes, tx.Meta[:]...)

	// set organization tx (32 byte) if not genesis
	if tx.ObjectType != Genesis {
		fmt.Println("It is not a genesis transaction")
		txHashLength = 32 // TRANSACTION_HASH_LENGTH
		transactionBytes = append(transactionBytes, tx.OrganizationTx...)
	}

	// set message length bytes as uint32 and append the message (4 + mLength)
	transactionBytes = append(transactionBytes, []byte{0, 0, 0, 0}...)
	binary.LittleEndian.PutUint32(transactionBytes[metaLength+txHashLength:metaLength+txHashLength+messageLenBytes], uint32(len(tx.Data)))
	transactionBytes = append(transactionBytes, tx.Data[:]...)

	// set target byte length as uint16 and append targets (2 + tLength)
	targetString := strings.Join(tx.Targets, ",")
	transactionBytes = append(transactionBytes, []byte{0, 0}...)
	binary.LittleEndian.PutUint16(transactionBytes[metaLength+txHashLength+messageLenBytes+len(tx.Data):metaLength+txHashLength+messageLenBytes+len(tx.Data)+targetLenBytes], uint16(len([]byte(targetString))))
	transactionBytes = append(transactionBytes, []byte(targetString)[:]...)

	// append signatures
	for i := range tx.Signatures {
		transactionBytes = append(transactionBytes, tx.PublicKeys[i][:]...)
		transactionBytes = append(transactionBytes, tx.Signatures[i][:]...)
		transactionBytes = append(transactionBytes, tx.DerivationPaths[i][:]...)
	}
	return transactionBytes
}

// Get bytes and construct the transaction
func ParseBytes(txBytes []byte) (*Transaction, error) {

	// total length of byte array
	// Genesis: 4 + 4 + mLength + 2 + targetLength + signatures
	// Normal:  4 + 32 + 4 + mLength + 2 + targetLength + signatures
	metaLength, txHashLength, messageLenBytes, targetLenBytes := 4, 0, 4, 2

	// This is not a tx
	if len(txBytes) < 8 {
		return nil, errors.New("Short message length")
	}

	tx := Transaction{}

	// Get meta
	meta := txBytes[0:metaLength]
	fmt.Println("OBJECT TYPE:", ObjectType(meta[0]).String())
	if ObjectType(meta[0]) != Genesis {
		fmt.Println("It is not a genesis transaction")
		txHashLength = 32
		tx.OrganizationTx = txBytes[metaLength : metaLength+txHashLength]
	}

	// Get message length
	messageLengthBytes := txBytes[metaLength+txHashLength : metaLength+txHashLength+messageLenBytes]
	messageLength := binary.LittleEndian.Uint32(messageLengthBytes)
	fmt.Println("Message length: ", messageLength)
	// Get message
	if messageLength < 1 || len(txBytes) < metaLength+txHashLength+messageLenBytes {
		return nil, errors.New("Invalid transaction")
	}
	tx.Data = txBytes[metaLength+txHashLength+messageLenBytes : metaLength+txHashLength+messageLenBytes+int(messageLength)]
	fmt.Println("Message: ", string(tx.Data))

	// Get targets length
	targetLengthBytes := txBytes[metaLength+txHashLength+messageLenBytes+int(messageLength) : metaLength+txHashLength+messageLenBytes+int(messageLength)+targetLenBytes]
	targetLength := binary.LittleEndian.Uint16(targetLengthBytes)
	fmt.Println("Targets length: ", targetLength, int(targetLength))
	// Get targets
	if len(txBytes) < metaLength+txHashLength+messageLenBytes+int(messageLength)+targetLenBytes+int(targetLength) {
		return nil, errors.New("Invalid transaction")
	}
	tx.Targets = strings.Split(string(txBytes[metaLength+txHashLength+messageLenBytes+int(messageLength)+targetLenBytes:metaLength+txHashLength+messageLenBytes+int(messageLength)+targetLenBytes+int(targetLength)]), ",")
	fmt.Println("Targets: ", tx.Targets)

	// Get signature(s) = [](public key + signature)
	multipleSignatureBytes := txBytes[metaLength+txHashLength+messageLenBytes+int(messageLength)+targetLenBytes+int(targetLength):]
	fmt.Println("Signature bytes length: ", len(multipleSignatureBytes))

	// extract public keys and signatures
	for i := 0; (i+1)*113 <= len(multipleSignatureBytes); i++ {
		fmt.Println("Parsing ", i, "th signature")
		tx.PublicKeys = append(tx.PublicKeys, make([]byte, 33))
		copy(tx.PublicKeys[i][:], multipleSignatureBytes[i*101:i*101+33])
		tx.Signatures = append(tx.Signatures, make([]byte, 64))
		copy(tx.Signatures[i][:], multipleSignatureBytes[i*101+33:i*101+97])
		tx.DerivationPaths = append(tx.DerivationPaths, make([]byte, 16))
		copy(tx.DerivationPaths[i][:], multipleSignatureBytes[i*101+97:i*101+97+16])
		// convert derivation steps to uint32
		tx.DerivationSteps = append(tx.DerivationSteps, crypto.ParseDerivationPathBytes(tx.DerivationPaths[i]))
	}

	// Build transaction by casting the parameters into its arrays
	copy(tx.Meta[:], meta)

	// Make sure tx hash is set
	tx.CalculateHash()

	return &tx, nil
}

// Parse organization transaction's data
func ParseOrganizationData(data []byte) (*Organization, error) {
	var organization Organization
	err := json.Unmarshal(data, &organization)
	if err != nil {
		return nil, err
	}
	return &organization, nil
}

// DB Methods
// ---------------------------
// Append
func (tx *Transaction) Save() {
	fmt.Println("Saving transaction", hex.EncodeToString(tx.Hash))
	db.Insert(tx.ToDBItem())
}

// Save transactions related to this node (referred by and referring to)
func (tx *Transaction) SaveRelated() {
	db.InsertRelated(tx.ToDBItem())
}

// Get
func Retrieve(txid []byte) *Transaction {
	item := db.Get(txid[:])
	if len(item.Data) == 0 {
		fmt.Println("No such transaction with id", txid)
		return nil
	}

	tx, err := FromDBItem(item)
	if err != nil {
		fmt.Println("Error while building transaction by db entry", err)
		return nil
	}

	return tx
}

// Get by object type (e.g. File, Genesis, Object, Certificat)
func RetrieveByObjectType(objectType ObjectType) []*Transaction {
	items := db.GetByObjectType(byte(objectType))
	if len(items) == 0 {
		fmt.Println("No such transaction with object type", objectType)
	}

	var transactions []*Transaction
	for _, v := range items {
		tx, err := FromDBItem(v)
		if err != nil {
			fmt.Println("Error building transaction", err)
		}
		transactions = append(transactions, tx)
	}
	return transactions
}

func GetLast() (*Transaction, error) {
	return FromDBItem(db.GetLastTransaction())
}
