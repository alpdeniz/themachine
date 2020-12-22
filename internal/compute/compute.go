package compute

// Compute module allows distributed computation of executable transactions
// It works in a tree structure where any script can execute a remote call inside
// This module is just a dummy as there is much research and work to do
// Some TODOs:
// - Study efficient sandboxed environments for untrusted code execution. Docker, codejail ...
// - Study distributed computation, including related academic papers and Ethereum's EVM
// - Use existing connections from inside execution environment for remote calls

// Should require payment for computation

import (
	"fmt"
	"io"
	"os/exec"
)

type ComputeChannel []byte

var cmds = make(map[int]exec.Cmd)
var stdins = make(map[int]io.WriteCloser)

func Execute(code string) []byte {

	// import the machine computation module to include remote calls
	// cmd := exec.Command("/usr/bin/python3", "-m", "themachine", "-c", code)
	cmd := exec.Command("/usr/bin/python3", "-c", code)
	result, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error while computing", err)
	}

	return result
}

// Placeholder method for code validation
// To be used when validating an executable transaction
func ValidateCode(code string) (bool, error) {
	return true, nil
}
