package nsenter

import (
	"fmt"
	"os/exec"
)

type Nsenter struct {
	opts []string
	cmds []string
}

func New() (*Nsenter, error) {
	return &Nsenter{}, nil
}

func (n *Nsenter) GetExecCmd() *exec.Cmd {
	args := n.opts
	args = append(args, "--")
	args = append(args, n.cmds...)

	return exec.Command("nsenter", args...)
}

func (n *Nsenter) SetProgram(c []string) *Nsenter {
	n.cmds = c
	return n
}

func (n *Nsenter) SetOptAll() *Nsenter {
	n.opts = append(n.opts, "--all")
	return n
}

func (n *Nsenter) SetOptTarget(pid uint64) *Nsenter {
	n.opts = append(n.opts, "--target="+fmt.Sprint(pid))
	return n
}

func (n *Nsenter) SetOptMount(file *string) *Nsenter {
	if file == nil {
		n.opts = append(n.opts, "--mount")
	} else {
		n.opts = append(n.opts, "--mount="+*file)
	}
	return n
}

func (n *Nsenter) SetOptUTS(file *string) *Nsenter {
	if file == nil {
		n.opts = append(n.opts, "--uts")
	} else {
		n.opts = append(n.opts, "--uts="+*file)
	}
	return n
}

func (n *Nsenter) SetOptIPC(file *string) *Nsenter {
	if file == nil {
		n.opts = append(n.opts, "--ipc")
	} else {
		n.opts = append(n.opts, "--ipc="+*file)
	}
	return n
}

func (n *Nsenter) SetOptNetwork(file *string) *Nsenter {
	if file == nil {
		n.opts = append(n.opts, "--net")
	} else {
		n.opts = append(n.opts, "--net="+*file)
	}
	return n
}

func (n *Nsenter) SetOptPID(file *string) *Nsenter {
	if file == nil {
		n.opts = append(n.opts, "--pid")
	} else {
		n.opts = append(n.opts, "--pid="+*file)
	}
	return n
}

func (n *Nsenter) SetOptCgroup(file *string) *Nsenter {
	if file == nil {
		n.opts = append(n.opts, "--cgroup")
	} else {
		n.opts = append(n.opts, "--cgroup="+*file)
	}
	return n
}

func (n *Nsenter) SetOptUser(file *string) *Nsenter {
	if file == nil {
		n.opts = append(n.opts, "--user")
	} else {
		n.opts = append(n.opts, "--user="+*file)
	}
	return n
}

func (n *Nsenter) SetOptUid(uid int) *Nsenter {
	n.opts = append(n.opts, "--setuid="+fmt.Sprint(uid))
	return n
}

func (n *Nsenter) SetOptGid(gid int) *Nsenter {
	n.opts = append(n.opts, "--setgid="+fmt.Sprint(gid))
	return n
}

func (n *Nsenter) SetOptPreserveCredentials() *Nsenter {
	n.opts = append(n.opts, "--preserve-credentials")
	return n
}

func (n *Nsenter) SetOptRoot(path *string) *Nsenter {
	if path == nil {
		n.opts = append(n.opts, "--root")
	} else {
		n.opts = append(n.opts, "--root="+*path)
	}
	return n
}

func (n *Nsenter) SetOptWd(path *string) *Nsenter {
	if path == nil {
		n.opts = append(n.opts, "--wd")
	} else {
		n.opts = append(n.opts, "--wd="+*path)
	}
	return n
}

func (n *Nsenter) SetOptNoFork() *Nsenter {
	n.opts = append(n.opts, "--no-fork")
	return n
}

func (n *Nsenter) SetOptFollowContext() *Nsenter {
	n.opts = append(n.opts, "--follow-context")
	return n
}
