package db

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MainDBClient mongo.Collection
var RelDBClient mongo.Collection
var KeyDBClient mongo.Collection

// Transaction structure
type MainDBItem struct {
	Index                   big.Int
	Hash                    []byte
	PrevHash                []byte
	Date                    time.Time
	ObjectType              byte
	SubType                 byte
	Data                    []byte
	Targets                 []string
	Signatures              [][]byte
	PublicKeys              [][]byte
	DerivationPaths         [][]byte
	OrganizationTransaction []byte // Genesis transaction of the organization referred by this transaction
}

// Key structure
type KeyDBItem struct {
	Name           string
	DerivationPath string
	Address        string
	PublicKey      []byte
	EncPrivateKey  []byte
}

// Connect to db on init
func init() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// store db access
	MainDBClient = *client.Database("themachine").Collection("transactions")        // main chain of all verified transactions
	RelDBClient = *client.Database("themachine").Collection("related_transactions") // transactions related to our keys, verified or not
	KeyDBClient = *client.Database("themachine").Collection("keys")                 // the keys consisting the account of this node

	fmt.Println("Connected to The Machine db ")
}
