package nsenter

import (
	"fmt"
	"os/exec"
)

type Builder struct {
	opts []string
	cmds []string
}

func New() (*Builder, error) {
	return &Builder{}, nil
}

func (b *Builder) GetExecCmd() *exec.Cmd {
	args := b.opts
	args = append(args, "--")
	args = append(args, b.cmds...)

	return exec.Command("nsenter", args...)
}

func (b *Builder) SetProgram(c []string) *Builder {
	b.cmds = c
	return b
}

func (b *Builder) SetOptAll() *Builder {
	b.opts = append(b.opts, "--all")
	return b
}

func (b *Builder) SetOptTarget(pid uint32) *Builder {
	b.opts = append(b.opts, "--target="+fmt.Sprint(pid))
	return b
}

func (b *Builder) SetOptMount(file *string) *Builder {
	if file == nil {
		b.opts = append(b.opts, "--mount")
	} else {
		b.opts = append(b.opts, "--mount="+*file)
	}
	return b
}

func (b *Builder) SetOptUTS(file *string) *Builder {
	if file == nil {
		b.opts = append(b.opts, "--uts")
	} else {
		b.opts = append(b.opts, "--uts="+*file)
	}
	return b
}

func (b *Builder) SetOptIPC(file *string) *Builder {
	if file == nil {
		b.opts = append(b.opts, "--ipc")
	} else {
		b.opts = append(b.opts, "--ipc="+*file)
	}
	return b
}

func (b *Builder) SetOptNetwork(file *string) *Builder {
	if file == nil {
		b.opts = append(b.opts, "--net")
	} else {
		b.opts = append(b.opts, "--net="+*file)
	}
	return b
}

func (b *Builder) SetOptPID(file *string) *Builder {
	if file == nil {
		b.opts = append(b.opts, "--pid")
	} else {
		b.opts = append(b.opts, "--pid="+*file)
	}
	return b
}

func (b *Builder) SetOptCgroup(file *string) *Builder {
	if file == nil {
		b.opts = append(b.opts, "--cgroup")
	} else {
		b.opts = append(b.opts, "--cgroup="+*file)
	}
	return b
}

func (b *Builder) SetOptUser(file *string) *Builder {
	if file == nil {
		b.opts = append(b.opts, "--user")
	} else {
		b.opts = append(b.opts, "--user="+*file)
	}
	return b
}

func (b *Builder) SetOptUid(uid int) *Builder {
	b.opts = append(b.opts, "--setuid="+fmt.Sprint(uid))
	return b
}

func (b *Builder) SetOptGid(gid int) *Builder {
	b.opts = append(b.opts, "--setgid="+fmt.Sprint(gid))
	return b
}

func (b *Builder) SetOptPreserveCredentials() *Builder {
	b.opts = append(b.opts, "--preserve-credentials")
	return b
}

func (b *Builder) SetOptRoot(path *string) *Builder {
	if path == nil {
		b.opts = append(b.opts, "--root")
	} else {
		b.opts = append(b.opts, "--root="+*path)
	}
	return b
}

func (b *Builder) SetOptWd(path *string) *Builder {
	if path == nil {
		b.opts = append(b.opts, "--wd")
	} else {
		b.opts = append(b.opts, "--wd="+*path)
	}
	return b
}

func (b *Builder) SetOptNoFork() *Builder {
	b.opts = append(b.opts, "--no-fork")
	return b
}

func (b *Builder) SetOptFollowContext() *Builder {
	b.opts = append(b.opts, "--follow-context")
	return b
}
