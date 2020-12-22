package network

import "testing"

func TestNetwork(t *testing.T) {
	StartNetwork(8443)

	c, err := ConnectToNode("localhost")
	if err != nil {
		t.Error("Cannot connect to node", err)
	}

	err = c.Relay([]byte("HELLO"))
	if err != nil {
		t.Error("Cannot relay", err)
	}

}
