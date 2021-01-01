package network

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Connection struct {
	Conn net.Conn
}

func initConnection(conn net.Conn) Connection {
	return Connection{
		conn,
	}
}

func (c *Connection) connect() ([]string, error) {
	// send initial code
	_, err := c.write([]byte{byte(Connect)})
	if err != nil {
		fmt.Println("Error sending bytes to peer", err)
		c.Conn.Close()
		return nil, err
	}

	// read reply to connect, which is a slice of peer strings
	message, err := c.read()
	if err != nil {
		fmt.Println("Error reading bytes from peer", err)
		c.Conn.Close()
		return nil, err
	}

	// get peer slice
	peers := strings.Split(string(message), ",")
	return peers, nil
}

// Relays the transaction to the connected node, response is handled in connection handler thread
func (c *Connection) Relay(tx []byte) error {
	fmt.Println("Relaying to ", c.Conn.RemoteAddr().String())
	_, err := c.write(prependCode(Relay, tx))
	if err != nil {
		fmt.Println("Could not relay transaction to", c.Conn.RemoteAddr().String(), err)
		return err
	}

	return nil
}

// Asks for the last transaction recorded to the node
func (c *Connection) GetHead() {
	fmt.Println("Getting head from ", c.Conn.RemoteAddr().String())
	c.write(prependCode(Head, endMessage([]byte{})))
}

func appendIfMissing(slice []string, elem string) []string {
	for _, v := range slice {
		if v == elem {
			return slice
		}
	}
	return append(slice, elem)
}

func endMessage(message []byte) []byte {
	return append(message, MESSAGE_TERMINATOR)
}

func prependCode(ct MessageType, message []byte) []byte {
	return append([]byte{byte(ct)}, message...)
}

// read & write (append message type and terminator bytes)
func (c *Connection) read() ([]byte, error) {
	message, err := bufio.NewReader(c.Conn).ReadBytes(MESSAGE_TERMINATOR)
	if err != nil {
		return nil, err
	}

	return message[:len(message)-1], nil
}

// write function writes the given message + EOF byte 0xFF
func (c *Connection) write(message []byte) (int, error) {
	return c.Conn.Write(endMessage(message))
}
