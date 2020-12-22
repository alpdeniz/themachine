package keystore

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/alpdeniz/themachine/internal/crypto"
	"github.com/alpdeniz/themachine/internal/db"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/pbkdf2"
)

type KeyPair struct {
	Name           string
	DerivationPath string
	Address        string
	PublicKey      []byte
	PrivateKey     []byte
}

type KeyMap map[string]KeyPair

var CurrentKeyMap = KeyMap{}
var WALLET_FILE string
var addresses []string
var encryptionKey []byte

// kdf params
const keyLength = 32
const iterations = 1000000

// Derivation Paths
var MasterKeyDerivationPath = "0" // "0/1'/5'/10"
// 
var salt = []byte{2, 243, 118, 3, 1, 98, 46, 254, 251, 7, 1, 0, 33, 16, 100, 182, 207, 199, 255, 54, 13}

// Opens up and loads the keypairs saved in this node
func Open() bool {

	// Ask for password
	fmt.Println("Password:")
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Cannot read passwod, using default")
		password = "123"
	}
	// Trim new line
	password = strings.TrimSuffix(password, "\n")
	// derive key from password
	encryptionKey = pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha256.New)

	// Now set keys up
	keys := db.GetKeyPairs()
	fmt.Println("Keys: ", len(keys))
	if len(keys) == 0 {
		// generate new set of keys and save
		NewKeyPair("Node") // or request a master key derived one from an organization master public key

	} else {
		// Decrypt and load into memory
		for _, v := range keys {

			// Decrypt private key for use
			private, err := crypto.Decrypt(v.EncPrivateKey, encryptionKey)
			if err != nil {
				fmt.Println("Cannot open wallet with provided password:", err)
				return false
			}

			master, err := crypto.WalletFromExtendedKey(base58.Encode(private))
			if err != nil {
				fmt.Println("Cannot parse decrypted extended key", err)
				return false
			}

			// Load
			CurrentKeyMap[string(v.Address)] = KeyPair{
				v.Name,
				v.DerivationPath,
				v.Address,
				v.PublicKey,
				master.Key,
			}
		}
	}

	return true
}

// AddKeyPair if an organization assigns one
func AddKeyPair(name string, derivationPath string, extendedPrivateKey string) {

	master, err := crypto.WalletFromExtendedKey(extendedPrivateKey)
	if err != nil {
		fmt.Println("Error while adding keys", err)
		return
	}

	// serialize private key
	privkey := master.Serialize()
	// serialize pubkey
	pubkey := master.Pub().Key
	// calculate address
	address := master.Address()
	// add into account memory
	addresses = append(addresses, address)

	// encrypt private key
	encPrivate := crypto.Encrypt(privkey, encryptionKey)

	// persistent save
	db.AddKey(name, derivationPath, address, pubkey[:], encPrivate)
	// save into memory as well
	CurrentKeyMap[string(address[:])] = KeyPair{
		Name:           name,
		DerivationPath: derivationPath,
		Address:        address,
		PublicKey:      pubkey,
		PrivateKey:     master.Key,
	}
}

// NewKeyPair creates a new set of keys unrelated to an organization key hierarchy
func NewKeyPair(name string) *KeyPair {

	// Generate first keypair
	master, err := crypto.NewWallet()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// set private
	privkey := master.Serialize()
	// set compressed private key
	pubkey := master.Pub().Key
	// encrypt private key for storage
	encPrivate := crypto.Encrypt([]byte(privkey), encryptionKey)

	// calculate address
	address := master.Address()

	// save into db
	db.AddKey(name, MasterKeyDerivationPath, address, pubkey, encPrivate)
	// save into memory
	kp := KeyPair{
		name,
		MasterKeyDerivationPath,
		address,
		pubkey,
		master.Key,
	}
	CurrentKeyMap[string(address)] = kp
	fmt.Println("Generated a new set of keys", name, string(address))

	return &kp
}

// Sign by extended key
func SignWithExtendedKey(extendedKeyStr string, data []byte) ([]byte, error) {
	extendedKey, err := crypto.WalletFromExtendedKey(extendedKeyStr)
	if err != nil {
		return nil, err
	}

	return crypto.Sign([]byte(data), extendedKey.Key)
}

func IsOwnedAddress(address []byte) bool {
	_, ok := CurrentKeyMap[string(address[:])]
	return ok
}

// Gets a keypair by name from the memory
func GetKeyPairByName(name string) *KeyPair {
	for _, v := range CurrentKeyMap {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

// Gets a keypair by address from the memory
func GetKeyPairByAddress(address string) *KeyPair {
	for _, v := range CurrentKeyMap {
		if v.Address == address {
			return &v
		}
	}
	return nil
}
