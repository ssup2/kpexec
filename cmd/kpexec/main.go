package main

import (
	"fmt"
	"os"

	"github.com/ssup2/kpexec/pkg/cmd/kpexec"
)

func main() {
	// Run command
	cmd := kpexec.New()
	if err := cmd.Execute(); err != nil {
		fmt.Printf("failed to execute kpexec error : %+v\n", err)
		os.Exit(1)
	}
}
