package network

import (
	"fmt"
	mrand "math/rand"
	"net"
)

type MessageType byte

const (
	Connect         MessageType = iota // 0 Get other node addresses
	Head                               // 1 Get last transaction info
	Relay                              // 2 Relay transactions forward - first verify, sign if it is asked by the transaction, then forward
	Compute                            // 3 Compute transaction code by given id - if authorized
	Fetch                              // 4 Fetch transaction by given id - public
	ConnectResponse                    // Currently unused since connection is passed to handler after
	HeadResponse
	RelayResponse
	ComputeResponse
	FetchResponse
)

type ConnectionChannel struct {
	IsClosed   bool
	Connection *Connection
}

var SOCKET_PORT = 8443

const MAXIMUM_CONNECTIONS = 20
const MESSAGE_TERMINATOR = byte(0xFF) // socket reads until

var connections []*Connection         // connection pool
var peers []string                    // keeps active peer list as string
var knownPeers []string               // keeps all peer list as string
var ch = make(chan ConnectionChannel) // connection save/remove channel
var stopServer = false
var seeds = []string{"164.90.197.171"}

// StartSocket fires up the socket listener and signal channel handler
// param port int (optional)
func StartNetwork(port int) {
	// set socket listener port (optional)
	SOCKET_PORT = port
	fmt.Println("Starting socket server on ", SOCKET_PORT)
	// start socket listener
	go listen(ch)
	// listen for all new and removed connections via signal channel
	go startChannelListener()
	// connect to other nodes supplied in seeds
	go ConnectToNodes(seeds, ch)
}

// Stop channel and socket listeners, close connections
func StopNetwork() {
	stopServer = true
	// Close and remove all connections
	for _, c := range connections {
		removeConnection(c)
	}
	fmt.Println("Stopped network...")
}

// startChannelListener listens for all channel messages to add or remove connections
func startChannelListener() {
	// loop forever
	for {
		// until
		if stopServer {
			return
		}
		result := <-ch
		if result.IsClosed {
			removeConnection(result.Connection)
		} else {
			saveConnection(result.Connection)
		}
	}
}

// listen starts Socket listener, signals back on connection and passes connections to its handler
func listen(ch chan<- ConnectionChannel) {

	// set listener
	ln, _ := net.Listen("tcp", fmt.Sprintf("%s:%d", "", SOCKET_PORT))
	defer ln.Close()

	// run loop forever
	for {
		// until
		if stopServer {
			return
		}
		// accept connection
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error while accepting a connection:", err)
			return
		}

		// create connection interface
		c := initConnection(conn)
		// signal back in order to save
		ch <- ConnectionChannel{
			false,
			&c,
		}

		// handle in different thread
		go c.handle(ch)
	}
}

// ConnectToNode
func ConnectToNode(host string) (*Connection, error) {

	// connect to socket
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, SOCKET_PORT))
	if err != nil {
		return nil, err
	}

	// wrap connection
	c := initConnection(conn)
	peers, err = c.connect()
	if err != nil {
		return nil, err
	}

	fmt.Println(peers)
	// // continue to look for new connections
	go ConnectToNodes(peers, ch)

	return &c, nil
}

func ConnectToNodes(list []string, ch chan<- ConnectionChannel) {
	for _, v := range list {
		// keep under limit
		if len(connections) >= MAXIMUM_CONNECTIONS {
			fmt.Println("Reached maximum number of connections", len(connections))
			return
		}

		if isInSlice(peers, v) {
			return
		}

		// find peers
		conn, err := ConnectToNode(v)
		if err != nil {
			fmt.Println("Cannot connect to peer", v, err)
			continue
		}
		fmt.Println("Got new connection via seeds")
		ch <- ConnectionChannel{
			false,
			conn,
		}

		// handle in different threads
		go conn.handle(ch)
	}

	// Syncronize with peers
	go StartToSyncronize()
}

// Relays a transaction to all connected nodes
func RelayTransaction(origin *Connection, tx []byte) int {

	counter := 0
	for _, c := range connections {
		// do not send it back to the connection you got it from
		if c == origin {
			fmt.Println("Skipping relaying to origin", origin.Conn.RemoteAddr().String())
			continue
		}

		err := c.Relay(tx)
		if err != nil {
			fmt.Println("Could not relay transaction to", origin.Conn.RemoteAddr().String())
			continue
		}

		counter++
	}
	return counter
}

// Syncronize transactions after startup
func StartToSyncronize() {
	// Node is not connected to the network, stop
	if len(connections) == 0 {
		return
	}
	// Get head from up to 10 random nodes
	for i := 0; i < 10 && i < len(connections); i++ {
		conn := connections[mrand.Intn(len(connections))]
		conn.GetHead()
	}
}

func saveConnection(c *Connection) {
	// save connection
	connections = append(connections, c)
	// keep string peer list ready
	host, _, _ := net.SplitHostPort(c.Conn.RemoteAddr().String())
	knownPeers = appendIfMissing(knownPeers, host)
	peers = appendIfMissing(peers, host)
	fmt.Println("Saved connection: ", len(connections))
}

func removeConnection(c *Connection) {
	fmt.Println("Connection to remove ", c.Conn.RemoteAddr().String())
	c.Conn.Close()
	c.Conn = nil
	for i, v := range connections {
		if v.Conn == nil {
			fmt.Println("Removed one connection ")
			connections = append(connections[:i], connections[i+1:]...)
			peers = append(peers[:i], peers[i+1:]...)
		}
	}
	fmt.Println("Removed connection: ", len(connections))
}

func isInSlice(slice []string, needle string) bool {
	for _, v := range slice {
		if v == needle {
			return true
		}
	}
	return false
}
