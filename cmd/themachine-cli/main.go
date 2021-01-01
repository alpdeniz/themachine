package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alpdeniz/themachine/internal/keystore"
	"github.com/alpdeniz/themachine/internal/network"
	"github.com/alpdeniz/themachine/internal/transaction"
)

func parseArguments() (actionType int, objectType int, command string, message string) {
	flag.StringVar(&command, "c", "GetHead", "Command to execute") // GetHead, Broadcast,
	flag.IntVar(&actionType, "a", 0, "Transaction type to broadcast")
	flag.IntVar(&objectType, "o", 0, "Transaction type to broadcast")
	flag.StringVar(&message, "f", "Hi", "Transaction message to broadcast")
	flag.Parse()

	return actionType, objectType, command, message
}

var testOrg = transaction.Organization{
	Name:                            "Test Org",
	Description:                     "Test Desc",
	MasterPublicKey:                 []byte{},
	MinimumRequiredSignaturePaths:   []string{"1/*+"},
	RequiredSignaturePathsPerObject: map[string][]string{},
	Rules:                           []string{"Rule1", "Rule2"},
}

func main() {

	keystore.Open()
	conn, err := network.ConnectToNode("127.0.0.1")
	if err != nil {
		fmt.Println("Could not connect to node ", err)
	}

	actionTypeInt, objectTypeInt, command, message := parseArguments()

	switch command {
	case "GetHead":
		conn.GetHead()
	case "Broadcast":
		actionType := network.MessageType(actionTypeInt)
		objectType := transaction.ObjectType(objectTypeInt)

		fmt.Println("Type: ", actionType, objectType, message)

		targets := []string{"m/1'/5'/2"}

		orgDefinition, err := json.Marshal(testOrg)
		if err != nil {
			fmt.Println("Error building organization definition bytes")
		}
		// testOrganizationTxHash := []byte{151, 2, 198, 59, 111, 224, 90, 47, 1, 250, 57, 117, 143, 99, 238, 144, 137, 184, 208, 170, 201, 15, 131, 164, 56, 35, 73, 142, 134, 31, 142, 252}

		tx, err := transaction.Build(transaction.Genesis, "json", []byte{}, orgDefinition, targets)
		if err != nil {
			fmt.Printf("Cannot build transaction %s\n", err)
			return
		}

		fmt.Println("Tx contents: ", string(tx.ToBytes()))
		fmt.Println("Tx hash: ", hex.EncodeToString(tx.Hash))

		hash := conn.Relay(tx.ToBytes())
		fmt.Printf("Sent to peers: %s to %s\n", hash, conn.Conn.RemoteAddr().Network())
	}

	// Handle ctrl+c signal as shutdown
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	network.StopNetwork()
	os.Exit(1)
}
