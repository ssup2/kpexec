package nsenter

import (
	"bytes"
	"testing"
)

func TestNsenterBuilder(t *testing.T) {
	// Set nsenter
	nse, _ := New()
	nse.SetProgram([]string{"echo", "test"})

	// Get nsenter command, Set stdout/stderr
	cmd := nse.GetExecCmd()
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	// Run nsenter
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Get stdout
	if outb.String() != "test\n" {
		t.Fatalf("stdout %v is not expected", outb.String())
	}
}
