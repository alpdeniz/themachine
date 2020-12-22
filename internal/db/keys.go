package db

// DB methods related to Keys
// - GetKeyByName 			  Directly by name
// - GetKeyByDerivationPath   Filters by derivation path
// - GetKeyPairs              Returns all keys
// - AddKey                   For keys provided by an organization
// - CountNumberOfKeys        Dummy

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

// Gets a single keypair by its given name
func GetKeyByName(name string) KeyDBItem {
	var key KeyDBItem
	filter := bson.M{"name": name}
	KeyDBClient.FindOne(context.TODO(), filter).Decode(&key)
	return key
}

// GetsKeyByDerivationPath fetches a list of keys with given a derivation path
func GetsKeyByDerivationPath(derivationPath string) []KeyDBItem {
	var keys = []KeyDBItem{}

	filter := bson.M{"derivationpath": derivationPath}
	cur, err := KeyDBClient.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("ERROR")
		log.Fatal(err)
	}

	for i := 0; cur.Next(context.TODO()); {
		keys = append(keys, KeyDBItem{})
		err := cur.Decode(&keys[i])
		if err != nil {
			log.Fatal(err)
		}
		i++
	}

	return keys
}

// Get all keypairs in this account as a list
func GetKeyPairs() []KeyDBItem {
	var accountKeys = []KeyDBItem{}

	filter := bson.D{}
	cur, err := KeyDBClient.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("ERROR")
		log.Fatal(err)
	}

	for i := 0; cur.Next(context.TODO()); {
		accountKeys = append(accountKeys, KeyDBItem{})
		err := cur.Decode(&accountKeys[i])
		if err != nil {
			log.Fatal(err)
		}
		i++
	}

	return accountKeys
}

// Add new key to the account
func AddKey(name string, derivationPath string, address string, publicKey []byte, encPrivateKey []byte) {
	keyDBItem := KeyDBItem{
		name,
		derivationPath,
		address,
		publicKey,
		encPrivateKey,
	}

	_, err := KeyDBClient.InsertOne(context.TODO(), keyDBItem)
	if err != nil {
		log.Fatal(err)
	}

}

func CountNumberOfKeys() int64 {
	count, err := KeyDBClient.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println("Cannot count keys")
	}
	return count
}
