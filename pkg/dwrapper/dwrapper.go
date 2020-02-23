package dwrapper

import (
	"context"

	"github.com/containerd/containerd"
	taskservice "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/namespaces"
)

type Dwrapper struct {
	client *containerd.Client
	ctx    context.Context
}

func New(dSocket string, ns string) (*Dwrapper, error) {
	client, err := containerd.New(dSocket)
	if err != nil {
		return nil, err
	}
	ctx := namespaces.WithNamespace(context.Background(), ns)

	return &Dwrapper{
		client: client,
		ctx:    ctx,
	}, nil
}

func (d *Dwrapper) Close() error {
	return d.client.Close()
}

func (d *Dwrapper) GetInitPid(cName string) (uint32, error) {
	taskClient := d.client.TaskService()
	c, err := taskClient.Get(d.ctx, &taskservice.GetRequest{
		ContainerID: cName,
	})
	if err != nil {
		return 0, err
	}
	return c.Process.Pid, nil
}

func (d *Dwrapper) GetRootDir(cName string) (string, error) {
	c, err := d.client.LoadContainer(d.ctx, cName)
	if err != nil {
		return "", err
	}
	s, err := c.Spec(d.ctx)
	if err != nil {
		return "", err
	}
	return s.Root.Path, nil
}

func (d *Dwrapper) GetWorkingDir(cName string) (string, error) {
	c, err := d.client.LoadContainer(d.ctx, cName)
	if err != nil {
		return "", err
	}
	s, err := c.Spec(d.ctx)
	if err != nil {
		return "", err
	}
	return s.Process.Cwd, nil
}

func (d *Dwrapper) GetEnv(cName string) ([]string, error) {
	c, err := d.client.LoadContainer(d.ctx, cName)
	if err != nil {
		return nil, err
	}
	s, err := c.Spec(d.ctx)
	if err != nil {
		return nil, err
	}
	return s.Process.Env, nil
}
