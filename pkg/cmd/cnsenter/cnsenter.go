package cnsenter

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ssup2/kpexec/pkg/crictl"
	"github.com/ssup2/kpexec/pkg/nsenter"
)

const (
	OptRuntimeContainerd = "containerd"
	OptRuntimeCrio       = "cri-o"
	OptRuntimeDocker     = "docker"

	cnsenterExample = `
		# Run date command in containerd container's all namespaces.
		cnsenter -r containerd -c [CONTAINER ID] -a date

		# Run bash command in cri-o container's PID namespace and network namespace with two dash
		cnsenter -r cri-o -c [CONTAINER ID] -p -n -- bash -il

		# Run bash command with additional environment variables
		cnsenter -r containerd -c [CONTAINER ID] -a key1=value1 -e key2=value2 -- bash -il

		# Set CRI socket path / containerd socket path
		cnsenter -c [CONTAINER ID] --cri [CRI SOCKET PATH / CONTAINERD SOCKET PATH] -a date
		`
)

var (
	version = "latest"
)

// Cmd
func New() *cobra.Command {
	options := &Options{}

	cmd := &cobra.Command{
		Use:                   "cnsenter -c [CONTAINER ID] [flags] -- COMMAND [args...]",
		DisableFlagsInUseLine: true,
		Short:                 "Execute a command in a container through the CRI",
		Long:                  "Execute a command in a container through the CRI",
		Example:               cnsenterExample,
		Run: func(cmd *cobra.Command, args []string) {
			if options.version {
				fmt.Printf("version: %s\n", version)
			} else {
				if err := options.Run(args); err != nil {
					fmt.Printf("failed to run cnsenter : %+v\n", err)
					os.Exit(1)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&options.contRuntime, "runtime", "r", OptRuntimeContainerd, "container runtime")
	cmd.Flags().StringVarP(&options.contID, "container", "c", "", "container ID to enter")
	cmd.Flags().StringVarP(&options.criSocket, "cri", "", "", "CRI socket path")

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

	cmd.Flags().StringArrayVarP(&options.envs, "env", "e", nil, "set a additional environment")

	cmd.Flags().BoolVarP(&options.version, "version", "v", false, "Show version")

	return cmd
}

// CnsenterOptions
type Options struct {
	contRuntime string
	contID      string
	criSocket   string

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

	envs []string

	version bool
}

func (o *Options) Run(args []string) error {
	// Validate args and options
	if len(args) == 0 {
		return fmt.Errorf("you must specify at least one command for the container")
	}
	if len(o.contID) == 0 {
		return fmt.Errorf("container name must be specified")
	}

	// Get container infos via crictl
	cri, err := crictl.New(o.contRuntime)
	if err != nil {
		return err
	}
	if o.criSocket != "" {
		cri.SetSocketPath(o.criSocket)
	}

	contPID, err := cri.GetInitPid(o.contID)
	if err != nil {
		return err
	}
	contRoot, err := cri.GetRootPath(o.contID)
	if err != nil {
		return err
	}
	contWorkingDir, err := cri.GetCWDPath(o.contID)
	if err != nil {
		return err
	}
	contEnvs, err := cri.GetEnvs(o.contID)
	if err != nil {
		return err
	}

	// Allocate nsenter
	nse, err := nsenter.New()
	if err != nil {
		return err
	}

	// Set PID, command
	nse.SetOptTarget(contPID)
	nse.SetProgram(args)

	// Set namespace
	if o.nsAll {
		nse.SetOptAll()
	}
	if o.nsMount {
		nse.SetOptMount(nil)
	}
	if o.nsUTS {
		nse.SetOptUTS(nil)
	}
	if o.nsIPC {
		nse.SetOptIPC(nil)
	}
	if o.nsNet {
		nse.SetOptNetwork(nil)
	}
	if o.nsPID {
		nse.SetOptPID(nil)
	}
	if o.nsCgroup {
		nse.SetOptCgroup(nil)
	}
	if o.nsUser {
		nse.SetOptUser(nil)
	}

	// Set root and working directory
	if o.rootSymbolic != "" {
		if err := os.Symlink(contRoot, o.rootSymbolic); err != nil {
			return err
		}
	}
	if o.workingDir {
		if o.workingDirBase == "" {
			nse.SetOptWd(nil)
		} else {
			wd := o.workingDirBase + contWorkingDir
			contEnvs = append(contEnvs, "PWD="+wd) // For shells, set PWD env
			nse.SetOptWd(&wd)
		}
	}

	// Set UID, GID
	if o.uid != 0 {
		nse.SetOptUid(o.uid)
	}
	if o.gid != 0 {
		nse.SetOptGid(o.gid)
	}

	// Append envs
	for _, env := range o.envs {
		contEnvs = append(contEnvs, env)
	}

	// Run nsenter
	cmd := nse.GetExecCmd()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = contEnvs
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
