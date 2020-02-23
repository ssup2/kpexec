package cnsenter

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ssup2/kpexec/pkg/dwrapper"
	"github.com/ssup2/kpexec/pkg/nsenter"

	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	OptDefaultSocketFile = "/run/containerd/containerd.sock"
	OptRuntimeDocker     = "docker"
)

var (
	cnsenterExample = templates.Examples(i18n.T(`
		# Run date in all mycontainer container namespaces.
		cnsenter -c mycontainer -a date

		# Run bash in PID namespace and network namespace of mycontainer container with two dash
		cnsenter -c mycontainer -p -n -- bash -il
		`))
)

// Cmd
func NewCmdCnsenter() *cobra.Command {
	options := &CnsenterOptions{}

	cmd := &cobra.Command{
		Use:                   "cnsenter CONTAINER [flags] -- COMMAND [args...]",
		DisableFlagsInUseLine: true,
		Short:                 "Execute a command in a container.",
		Long:                  "Execute a command in a container.",
		Example:               cnsenterExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(options.Complete(args))
			cmdutil.CheckErr(options.Validate())
			cmdutil.CheckErr(options.Run())
		},
	}

	cmd.Flags().StringVarP(&options.BaseDir, "basedir", "b", "", "set the base (host root) directory")

	cmd.Flags().StringVarP(&options.ContName, "container", "c", "", "container ID to enter")
	cmd.Flags().StringVarP(&options.ContRuntime, "runtime", "R", OptRuntimeDocker, "container runtime")
	cmd.Flags().StringVarP(&options.ContdSocket, "socket", "s", OptDefaultSocketFile, "containerd unix domain socket path")

	cmd.Flags().BoolVarP(&options.NsAll, "all", "a", false, "enter all container namespace")
	cmd.Flags().BoolVarP(&options.NsMount, "mount", "m", false, "enter container mount namespace")
	cmd.Flags().BoolVarP(&options.NsUTS, "uts", "u", false, "enter container UTS namespace")
	cmd.Flags().BoolVarP(&options.NsIPC, "ipc", "i", false, "enter container IPC namespace")
	cmd.Flags().BoolVarP(&options.NsNet, "net", "n", false, "enter container network namespace")
	cmd.Flags().BoolVarP(&options.NsPID, "pid", "p", false, "enter container PID namespace")
	cmd.Flags().BoolVarP(&options.NsCgroup, "cgroup", "C", false, "enter container cgroup namespace")
	cmd.Flags().BoolVarP(&options.NsUser, "user", "U", false, "enter container user namespace")

	cmd.Flags().IntVarP(&options.UID, "setuid", "S", 0, "set uid in entered namespace")
	cmd.Flags().IntVarP(&options.GID, "setgid", "G", 0, "set gid in entered namespace")

	return cmd
}

// CnsenterOptions
type CnsenterOptions struct {
	BaseDir string

	ContRuntime string
	ContdSocket string
	ContName    string
	ContCommand []string

	NsAll    bool
	NsMount  bool
	NsUTS    bool
	NsIPC    bool
	NsNet    bool
	NsPID    bool
	NsCgroup bool
	NsUser   bool

	UID int
	GID int
}

func (p *CnsenterOptions) Complete(args []string) error {
	p.ContCommand = args
	return nil
}

func (p *CnsenterOptions) Validate() error {
	if len(p.ContName) == 0 {
		return fmt.Errorf("container name must be specified")
	}
	if len(p.ContCommand) == 0 {
		return fmt.Errorf("you must specify at least one command for the container")
	}

	return nil
}

func (p *CnsenterOptions) Run() error {
	// Set containerd namespace
	var namespace string
	if p.ContRuntime == "docker" {
		namespace = "moby"
	} else {
		return fmt.Errorf("%s runtime not support", p.ContRuntime)
	}

	// Get container infos
	wrapper, err := dwrapper.New(p.BaseDir+p.ContdSocket, namespace)
	if err != nil {
		return err
	}
	defer wrapper.Close()

	pid, err := wrapper.GetInitPid(p.ContName)
	if err != nil {
		return err
	}
	envs, err := wrapper.GetEnv(p.ContName)
	if err != nil {
		return err
	}

	// Set nsenter
	builder, err := nsenter.New()
	if err != nil {
		return err
	}

	builder.SetOptTarget(pid)
	builder.SetProgram(p.ContCommand)

	if p.NsAll {
		builder.SetOptAll()
	}
	if p.NsMount {
		builder.SetOptMount(nil)
	}
	if p.NsUTS {
		builder.SetOptUTS(nil)
	}
	if p.NsIPC {
		builder.SetOptIPC(nil)
	}
	if p.NsNet {
		builder.SetOptNetwork(nil)
	}
	if p.NsPID {
		builder.SetOptPID(nil)
	}
	if p.NsCgroup {
		builder.SetOptCgroup(nil)
	}
	if p.NsUser {
		builder.SetOptUser(nil)
	}

	if p.UID != 0 {
		builder.SetOptUid(p.UID)
	}
	if p.GID != 0 {
		builder.SetOptGid(p.GID)
	}

	// run nsenter
	cmd := builder.GetCmd()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = envs
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
