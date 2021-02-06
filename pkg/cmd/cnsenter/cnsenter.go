package cnsenter

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ssup2/kpexec/pkg/dwrapper"
	"github.com/ssup2/kpexec/pkg/nsenter"
)

const (
	OptDefaultSocketFile = "/run/containerd/containerd.sock"

	OptRuntimeContainerd = "containerd"
	OptRuntimeDocker     = "docker"
)

var (
	cnsenterExample = `
		# Run date command in all docker container's namespaces.
		cnsenter -r docker -c [CONTAINER ID] -a date

		# Run bash command in containerd container's PID namespace and network namespace with two dash
		cnsenter -r containerd -c [CONTAINER ID] -p -n -- bash -il
		`
)

// Cmd
func New() *cobra.Command {
	options := &Options{}

	cmd := &cobra.Command{
		Use:                   "cnsenter -r [CONTAINER RUNTIME] -c [CONTAINER ID] [flags] -- COMMAND [args...]",
		DisableFlagsInUseLine: true,
		Short:                 "Execute a command in a container.",
		Long:                  "Execute a command in a container.",
		Example:               cnsenterExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Run(args); err != nil {
				fmt.Printf("failed to run cnsenter : %+v", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&options.contName, "container", "c", "", "container ID to enter")
	cmd.Flags().StringVarP(&options.contRuntime, "runtime", "r", OptRuntimeDocker, "container runtime")
	cmd.Flags().StringVarP(&options.contdSocket, "socket", "s", OptDefaultSocketFile, "containerd unix domain socket path")

	cmd.Flags().BoolVarP(&options.nsAll, "all", "a", false, "enter all container namespace")
	cmd.Flags().BoolVarP(&options.nsMount, "mount", "m", false, "enter container mount namespace")
	cmd.Flags().BoolVarP(&options.nsUTS, "uts", "u", false, "enter container UTS namespace")
	cmd.Flags().BoolVarP(&options.nsIPC, "ipc", "i", false, "enter container IPC namespace")
	cmd.Flags().BoolVarP(&options.nsNet, "net", "n", false, "enter container network namespace")
	cmd.Flags().BoolVarP(&options.nsPID, "pid", "p", false, "enter container PID namespace")
	cmd.Flags().BoolVarP(&options.nsCgroup, "cgroup", "C", false, "enter container cgroup namespace")
	cmd.Flags().BoolVarP(&options.nsUser, "user", "U", false, "enter container user namespace")

	cmd.Flags().StringVarP(&options.rootSymbolic, "root-symlink", "", "", "create the container's root symbolic link")
	cmd.Flags().BoolVarP(&options.workingDir, "wd", "w", false, "set the working directory")
	cmd.Flags().StringVarP(&options.workingDirBase, "wd-base", "", "", "set the working directory base path")

	cmd.Flags().IntVarP(&options.uid, "setuid", "S", 0, "set uid in entered namespace")
	cmd.Flags().IntVarP(&options.gid, "setgid", "G", 0, "set gid in entered namespace")

	return cmd
}

// CnsenterOptions
type Options struct {
	contRuntime string
	contdSocket string
	contName    string

	nsAll    bool
	nsMount  bool
	nsUTS    bool
	nsIPC    bool
	nsNet    bool
	nsPID    bool
	nsCgroup bool
	nsUser   bool

	rootSymbolic   string
	workingDir     bool
	workingDirBase string

	uid int
	gid int
}

func (o *Options) Run(args []string) error {
	// Validate args and options
	if len(args) == 0 {
		return fmt.Errorf("you must specify at least one command for the container")
	}
	if len(o.contName) == 0 {
		return fmt.Errorf("container name must be specified")
	}

	// Set containerd namespace
	var namespace string
	if o.contRuntime == OptRuntimeContainerd {
		namespace = "k8s.io"
	} else if o.contRuntime == OptRuntimeDocker {
		namespace = "moby"
	} else {
		return fmt.Errorf("%s runtime not support", o.contRuntime)
	}

	// Get container infos
	wrapper, err := dwrapper.New(o.contdSocket, namespace)
	if err != nil {
		return err
	}
	defer wrapper.Close()

	pid, err := wrapper.GetInitPid(o.contName)
	if err != nil {
		return err
	}
	root, err := wrapper.GetRootDir(o.contName)
	if err != nil {
		return err
	}
	workingDir, err := wrapper.GetWorkingDir(o.contName)
	if err != nil {
		return err
	}
	envs, err := wrapper.GetEnv(o.contName)
	if err != nil {
		return err
	}

	// Set nsenter
	builder, err := nsenter.New()
	if err != nil {
		return err
	}

	builder.SetOptTarget(pid)
	builder.SetProgram(args)

	if o.nsAll {
		builder.SetOptAll()
	}
	if o.nsMount {
		builder.SetOptMount(nil)
	}
	if o.nsUTS {
		builder.SetOptUTS(nil)
	}
	if o.nsIPC {
		builder.SetOptIPC(nil)
	}
	if o.nsNet {
		builder.SetOptNetwork(nil)
	}
	if o.nsPID {
		builder.SetOptPID(nil)
	}
	if o.nsCgroup {
		builder.SetOptCgroup(nil)
	}
	if o.nsUser {
		builder.SetOptUser(nil)
	}

	// Set root and working directory
	if o.rootSymbolic != "" {
		if err := os.Symlink(root, o.rootSymbolic); err != nil {
			return err
		}
	}
	if o.workingDir {
		if o.workingDirBase == "" {
			builder.SetOptWd(nil)
		} else {
			wd := o.workingDirBase + workingDir
			envs = append(envs, "PWD="+wd) // For shells, set PWD env
			builder.SetOptWd(&wd)
		}
	}

	// Set UID, GID
	if o.uid != 0 {
		builder.SetOptUid(o.uid)
	}
	if o.gid != 0 {
		builder.SetOptGid(o.gid)
	}

	// Run nsenter
	cmd := builder.GetExecCmd()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = envs
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
