package nsenter

import (
	"fmt"
	"os/exec"
)

type NsenterBuilder struct {
	opts []string
	prog []string
}

func New() (*NsenterBuilder, error) {
	return &NsenterBuilder{}, nil
}

func (n *NsenterBuilder) GetCmd() *exec.Cmd {
	args := n.opts
	args = append(args, "--")
	args = append(args, n.prog...)

	return exec.Command("nsenter", args...)
}

func (n *NsenterBuilder) SetProgram(p []string) *NsenterBuilder {
	n.prog = p
	return n
}

func (n *NsenterBuilder) SetOptAll() *NsenterBuilder {
	n.opts = append(n.opts, "--all")
	return n
}

func (n *NsenterBuilder) SetOptTarget(pid uint32) *NsenterBuilder {
	n.opts = append(n.opts, "--target="+fmt.Sprint(pid))
	return n
}

func (n *NsenterBuilder) SetOptMount(file *string) *NsenterBuilder {
	if file == nil {
		n.opts = append(n.opts, "--mount")
	} else {
		n.opts = append(n.opts, "--mount="+*file)
	}
	return n
}

func (n *NsenterBuilder) SetOptUTS(file *string) *NsenterBuilder {
	if file == nil {
		n.opts = append(n.opts, "--uts")
	} else {
		n.opts = append(n.opts, "--uts="+*file)
	}
	return n
}

func (n *NsenterBuilder) SetOptIPC(file *string) *NsenterBuilder {
	if file == nil {
		n.opts = append(n.opts, "--ipc")
	} else {
		n.opts = append(n.opts, "--ipc="+*file)
	}
	return n
}

func (n *NsenterBuilder) SetOptNetwork(file *string) *NsenterBuilder {
	if file == nil {
		n.opts = append(n.opts, "--net")
	} else {
		n.opts = append(n.opts, "--net="+*file)
	}
	return n
}

func (n *NsenterBuilder) SetOptPID(file *string) *NsenterBuilder {
	if file == nil {
		n.opts = append(n.opts, "--pid")
	} else {
		n.opts = append(n.opts, "--pid="+*file)
	}
	return n
}

func (n *NsenterBuilder) SetOptCgroup(file *string) *NsenterBuilder {
	if file == nil {
		n.opts = append(n.opts, "--cgroup")
	} else {
		n.opts = append(n.opts, "--cgroup="+*file)
	}
	return n
}

func (n *NsenterBuilder) SetOptUser(file *string) *NsenterBuilder {
	if file == nil {
		n.opts = append(n.opts, "--user")
	} else {
		n.opts = append(n.opts, "--user="+*file)
	}
	return n
}

func (n *NsenterBuilder) SetOptUid(uid int) *NsenterBuilder {
	n.opts = append(n.opts, "--setuid="+fmt.Sprint(uid))
	return n
}

func (n *NsenterBuilder) SetOptGid(gid int) *NsenterBuilder {
	n.opts = append(n.opts, "--setgid="+fmt.Sprint(gid))
	return n
}

func (n *NsenterBuilder) SetOptPreserveCredentials() *NsenterBuilder {
	n.opts = append(n.opts, "--preserve-credentials")
	return n
}

func (n *NsenterBuilder) SetOptRoot(dir string) *NsenterBuilder {
	n.opts = append(n.opts, "--root="+dir)
	return n
}

func (n *NsenterBuilder) SetOptWd(dir string) *NsenterBuilder {
	n.opts = append(n.opts, "--wd="+dir)
	return n
}

func (n *NsenterBuilder) SetOptNoFork() *NsenterBuilder {
	n.opts = append(n.opts, "--no-fork")
	return n
}

func (n *NsenterBuilder) SetOptFollowContext() *NsenterBuilder {
	n.opts = append(n.opts, "--follow-context")
	return n
}
