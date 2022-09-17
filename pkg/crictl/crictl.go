package crictl

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/containerd/containerd"
	taskservice "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/namespaces"
	"github.com/tidwall/gjson"
)

const (
	cliCrictl           = "crictl"
	cliCrictlOptInspect = "inspect"

	runtimeContainerd = "containerd"
	runtimeCrio       = "cri-o"
	runtimeDocker     = "docker"

	contdNsDocker   = "moby"
	contdSocketPath = "/run/containerd/containerd.sock"
)

type Crictl struct {
	// For crictl CLI
	runtime string
	opts    []string

	// For containerd client
	// Docker CRI with "unix:///var/run/dockershim.sock" doesn't return PID, CWD and Env info.
	// To avoid this issue, we use containerd client directly instead of crictl CLI.
	dCtx    context.Context
	dClient *containerd.Client
}

func New(rt string) (*Crictl, error) {
	// Docker
	// Init containerd client
	if rt == runtimeDocker {
		client, err := containerd.New(contdSocketPath)
		if err != nil {
			return nil, err
		}
		ctx := namespaces.WithNamespace(context.Background(), contdNsDocker)

		return &Crictl{
			runtime: runtimeDocker,
			dCtx:    ctx,
			dClient: client,
		}, nil
	}

	// Else
	// Only set runtime
	return &Crictl{
		runtime: rt,
	}, nil
}

func (c *Crictl) SetSocketPath(socketPath string) error {
	// Docker
	// Replace new client for new socket path
	if c.runtime == runtimeDocker {
		client, err := containerd.New(socketPath)
		if err != nil {
			return err
		}
		c.dClient = client
	}

	// Else
	// Set runtime endpoint option
	c.opts = append(c.opts, "--runtime-endpoint", "unix://"+socketPath)
	return nil
}

func (c *Crictl) GetInitPid(contID string) (uint64, error) {
	// Docker
	// Get PID from containerd
	if c.runtime == runtimeDocker {
		taskClient := c.dClient.TaskService()
		cont, err := taskClient.Get(c.dCtx, &taskservice.GetRequest{
			ContainerID: contID,
		})
		if err != nil {
			return 0, err
		}
		return uint64(cont.Process.Pid), nil
	}

	// Else
	// Get container info through crictl
	args := append(c.opts, cliCrictlOptInspect, contID)
	cmd := exec.Command(cliCrictl, args...)
	info, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parsing init PID
	pid := gjson.Get(string(info), "info.pid")
	return pid.Uint(), nil
}

func (c *Crictl) GetRootPath(contID string) (string, error) {
	// Docker
	// Get rootfs from containerd
	if c.runtime == runtimeDocker {
		cont, err := c.dClient.LoadContainer(c.dCtx, contID)
		if err != nil {
			return "", err
		}
		spec, err := cont.Spec(c.dCtx)
		if err != nil {
			return "", err
		}
		return spec.Root.Path, nil
	}

	// Else
	// Get container info through crictl
	args := append(c.opts, cliCrictlOptInspect, contID)
	cmd := exec.Command(cliCrictl, args...)
	info, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parsing rootfs and get absolute container rootfs path according to container runtime
	rootDirPath := gjson.Get(string(info), "info.runtimeSpec.root.path")
	if c.runtime == runtimeContainerd {
		return fmt.Sprintf("/run/containerd/io.containerd.runtime.v2.task/k8s.io/%s/%s", contID, rootDirPath.String()), nil
	}
	return rootDirPath.String(), nil // CRI-O
}

func (c *Crictl) GetCWDPath(contID string) (string, error) {
	// Docker
	// Get CWD from containerd
	if c.runtime == runtimeDocker {
		cont, err := c.dClient.LoadContainer(c.dCtx, contID)
		if err != nil {
			return "", err
		}
		spec, err := cont.Spec(c.dCtx)
		if err != nil {
			return "", err
		}
		return spec.Process.Cwd, nil
	}

	// Else
	// Get container info through crictl
	args := append(c.opts, cliCrictlOptInspect, contID)
	cmd := exec.Command(cliCrictl, args...)
	info, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parsing working dir
	cwd := gjson.Get(string(info), "info.runtimeSpec.process.cwd")
	return cwd.String(), nil
}

func (c *Crictl) GetEnvs(contID string) ([]string, error) {
	// Docker
	// Get envs from containerd
	if c.runtime == runtimeDocker {
		cont, err := c.dClient.LoadContainer(c.dCtx, contID)
		if err != nil {
			return nil, err
		}
		spec, err := cont.Spec(c.dCtx)
		if err != nil {
			return nil, err
		}
		return spec.Process.Env, nil
	}

	// Else
	// Get container info through crictl
	args := append(c.opts, cliCrictlOptInspect, contID)
	cmd := exec.Command(cliCrictl, args...)
	info, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parsing envs
	var result []string
	envs := gjson.Get(string(info), "info.runtimeSpec.process.env")
	for _, env := range envs.Array() {
		result = append(result, env.String())
	}
	return result, nil
}
