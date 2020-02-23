package main

import (
	"os"

	"github.com/ssup2/kpexec/pkg/cmd/kpexec"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func main() {
	// Set stdin/stdout
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	// Set flags
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Run command
	cmd := kpexec.NewCmdKpexec(f, ioStreams)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
