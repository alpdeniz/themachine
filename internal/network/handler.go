package network

// Main connection handler
// After a connection is made by/to the network, it is passed to the handler below
// It listens for and returns responses to each category of requests from other nodes

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/alpdeniz/themachine/internal/compute"
	"github.com/alpdeniz/themachine/internal/transaction"
)

// keep listening to all connections, parse and reply
func (c *Connection) handle(ch chan<- ConnectionChannel) {

	// Read loop
	for {
		// until
		if stopServer {
			return
		}
		// get message, output
		message, err := c.read()
		if err != nil {
			fmt.Println("Error reading incoming connection: ", err)
			// send remove signal
			ch <- ConnectionChannel{
				true,
				c,
			}
			return
		}

		// message = strings.TrimSuffix(message, "\n")
		fmt.Println("Message Received:", message)

		actionType := MessageType(message[0])
		switch actionType {

		case Connect:
			fmt.Println("Serving peers", len(connections)-1) // -1 for this connection as we won't be sending it back
			// put connected peers into a slice
			var peers []string
			for _, v := range connections {
				if v == c {
					continue
				}
				host, _, _ := net.SplitHostPort(v.Conn.RemoteAddr().String())
				peers = append(peers, host)
			}

			// Currently unused connect response format as it is handled before passing to connection handler thread
			// c.write(prependCode(ConnectResponse, []byte(strings.Join(peers, ","))))

			// Return response
			c.write([]byte(strings.Join(peers, ",")))

		case Head:

			fmt.Println("Serving HEAD to", c.Conn.RemoteAddr().String())
			lastTx, err := transaction.GetLast()
			if err != nil {
				fmt.Println("Error getting the last transaction as head", err)
			}
			// sends back the last index it has
			var tmp = make([]byte, 8)
			// cast uint64 to bytes
			binary.BigEndian.PutUint64(tmp, lastTx.Index)
			respnse := prependCode(HeadResponse, tmp)
			fmt.Println("Response to write", respnse)
			c.write(respnse)

		case HeadResponse:

			index := binary.BigEndian.Uint64(message[1:])
			fmt.Println("Got head response", index)

		case Fetch:

			if len(message) < 33 {
				fmt.Println("Error in fetch request. Short message length", len(message))
				c.write([]byte("Short message"))
				continue
			}

			txhash := message[1:33]
			tx := transaction.Retrieve(txhash)
			if tx == nil {
				c.write([]byte("No such transaction"))
				continue
			}

			c.write(prependCode(FetchResponse, tx.ToBytes()))

		case FetchResponse:

			// process and save this message (if valid)
			_, err := transaction.Process(message[1:])
			if err != nil {
				fmt.Println("Could not verify tx (fetch): ", err)
				c.write([]byte("Transaction rejected"))
				continue
			}

		case Relay:

			fmt.Println("Got relay request", string(message[1:]))
			// process and transmit message (if valid)
			tx, err := transaction.Process(message[1:])
			if err != nil {
				fmt.Println("Could not verify tx (relay):", err)
				c.write([]byte("Transaction rejected"))
				continue
			}

			relayed := RelayTransaction(c, message[1:])
			fmt.Println("RELAYED to", relayed, tx.Hash)

			c.write(prependCode(RelayResponse, tx.Hash))

		case RelayResponse:

			hash := message[1:]
			fmt.Println("Got response hash:", hex.EncodeToString(hash), len(hash))

		// Compute request handler
		case Compute:

			if len(message) < 33 {
				fmt.Println("Error in compute request. Short message length", len(message))
				c.write(prependCode(ComputeResponse, []byte("Short message")))
				continue
			}

			// to match the process on requester's side
			pid := message[1:5]
			requestCode := message[5:9]
			// tx to be executed
			txhash := message[5:37]
			// fetch
			tx := transaction.Retrieve(txhash)
			// check if executable
			var result []byte
			if tx.ObjectType == transaction.Executable {
				result = compute.Execute(string(tx.Data))
			} else {
				fmt.Println("Not an executable transaction:", tx.Hash)
				continue
			}

			// prepend message type, pid and request code to result
			response := append(prependCode(ComputeResponse, []byte(pid)), append(requestCode, result...)...)
			c.write(response)

		// Compute response handler. This is to stay here just in case.
		// Currently python module handles all communications, well, it could use
		// the node's existing connection, thus requiring the handler below
		case ComputeResponse:

			// parse computation response
			pid := message[1:5]
			requestCode := message[5:9]
			result := message[9:]
			fmt.Println("Got compute response", binary.BigEndian.Uint32(pid), binary.BigEndian.Uint32(requestCode), result)
			// feed into parent process
			// compute.HandleResponse(binary.BigEndian.Uint32(pid), binary.BigEndian.Uint32(requestCode), result)

		}
	}
}
