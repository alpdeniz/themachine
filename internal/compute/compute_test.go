package compute

import (
	"testing"
)

func TestExecute(t *testing.T) {

	// python code to test handling
	code := `
import sys, os
print("python output")`

	result := Execute(code)
	if string(result[:len(result)-1]) != "python output" {
		t.Error("Error with compute result")
	}

}
