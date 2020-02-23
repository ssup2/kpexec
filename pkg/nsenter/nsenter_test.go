package nsenter

import (
	"bytes"
	"testing"
)

func TestNsenterBuilder(t *testing.T) {
	// Set builder
	builder := NsenterBuilder{}
	builder.SetProgram([]string{"echo", "test"})

	// Get nsenter command, Set stdout/stderr
	cmd := builder.GetCmd()
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
