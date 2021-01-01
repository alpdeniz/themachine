package transaction

// Types for transaction package

import (
	"time"
)

type Organization struct {
	Name                            string
	Description                     string
	MasterPublicKey                 []byte              // so that everyone can recognize its children
	MinimumRequiredSignaturePaths   []string            // Minimum signature needed to validate a transaction
	RequiredSignaturePathsPerObject map[string][]string // An organization may allow some time and disallow another
	Rules                           []string            // no idea - dummy
}

type Transaction struct {
	Index                     uint64        // Index to work out longest chain
	Hash                      []byte        // hash of transaction contents
	Meta                      [4]byte       // To store transaction type + some flags
	ObjectType                ObjectType    // see TransactionObjectType
	SubType                   ObjectSubType // file type, cert type etc
	Targets                   []string      // Key derivation paths which are needed sign for this to be part of the chain
	OrganizationTx            []byte        // 32 bytes Genesis tx hash
	Organization              Organization  // points to organization Genesis tx
	DataLength                uint32
	Data                      []byte     // Data of ObjectTypes x
	Signatures                [][]byte   // Signatures 64 bytes each
	PublicKeys                [][]byte   // Respective public keys 33 bytes each
	DerivationPaths           [][]byte   // Respective derivation paths of public keys in relation to organizational master public key as 4 uint32
	DerivationSteps           [][]uint32 // Ready to use form of above
	MinimumRequiredSignatures []int      // for each derivation path
	Date                      time.Time
}

type ObjectSubType string

const (
	FileTypePDF    = "pdf"
	FileTypeJPG    = "jpg"
	FileTypePNG    = "png"
	FileTypeTXT    = "txt"
	FileTypeJSON   = "json"
	CodeTypePython = "python"
	CodeTypeNodeJS = "nodejs"
)

type ObjectType int

const (
	Genesis     ObjectType = iota // 0 to start a new organization       JSON
	File                          // 1 any significant file to store.    Binary
	Object                        // 2 to store records, information ... JSON
	Certificate                   // 3 e.g. university degree, awards    JSON/PDF
	Executable                    // 4 for compute
	Asset                         // 5 e.g. cars for notaries, land for land registries
	Token                         // 6 to pay for compute
	Decision                      // 7 e.g. verdict in court, a decision managers make
	Law                           // 8 Cannot be encrypted
	Proposal                      // 9 to vote for
	EncryptedFile
	EncryptedCertificate
	EncryptedDecision
	EncryptedIdentity // cannot be plaintext
	EncryptedProposal
	EncryptedExecutable
	EncryptedAsset
	EncryptedObject
)

// Returns string representation of a ObjectType
func (o ObjectType) String() string {
	return [...]string{"Genesis", "File", "Object", "Certificate", "Executable", "Asset", "Token", "Decision", "Law", "Proposal"}[o]
}
