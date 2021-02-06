package main

import (
	"os"

	"github.com/ssup2/kpexec/pkg/cmd/cnsenter"
)

func main() {
	// Run command
	cmd := cnsenter.New()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
