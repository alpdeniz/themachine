package db

// DB methods related to transactions

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Gets the transaction by its hash
func Get(txhash []byte) MainDBItem {
	var dbItem MainDBItem
	filter := bson.M{"hash": txhash}
	MainDBClient.FindOne(context.TODO(), filter).Decode(&dbItem)
	return dbItem
}

// Gets the last transaction
func GetLastTransaction() MainDBItem {
	var lastDbItem MainDBItem
	filter := bson.D{}
	findOneOptions := options.FindOne()
	findOneOptions.SetSort(bson.D{primitive.E{Key: "date", Value: -1}})
	MainDBClient.FindOne(context.TODO(), filter, findOneOptions).Decode(&lastDbItem)
	return lastDbItem
}

// Get organization transactions
func GetByObjectType(objectType byte) []MainDBItem {
	var transactions []MainDBItem
	filter := bson.M{"objecttype": objectType}
	cur, err := MainDBClient.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("ERROR")
		log.Fatal(err)
	}

	for i := 0; cur.Next(context.TODO()); {
		transactions = append(transactions, MainDBItem{})
		err := cur.Decode(&transactions[i])
		if err != nil {
			log.Fatal(err)
		}
		i++
	}

	return transactions
}

// Saves a transaction
func Insert(dbItem MainDBItem) {
	// get the last transaction
	lastDbItem := GetLastTransaction()
	dbItem.Index = lastDbItem.Index + 1
	dbItem.PrevHash = lastDbItem.Hash
	// go
	_, err := MainDBClient.InsertOne(context.TODO(), dbItem)
	if err != nil {
		log.Fatal(err)
	}
}

// Saves a transaction related to this account
func InsertRelated(dbItem MainDBItem) {

	_, err := RelDBClient.InsertOne(context.TODO(), dbItem)
	if err != nil {
		log.Fatal(err)
	}
}

func CountNumberOfTransactions() int64 {
	count, err := MainDBClient.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println("Cannot count transactions")
	}
	return count
}
