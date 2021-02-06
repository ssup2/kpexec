package dwrapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/containerd"
	taskservice "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/namespaces"
)

type Dwrapper struct {
	ctx context.Context

	client    *containerd.Client
	namespace string
}

func New(dSocket string, ns string) (*Dwrapper, error) {
	client, err := containerd.New(dSocket)
	if err != nil {
		return nil, err
	}
	ctx := namespaces.WithNamespace(context.Background(), ns)

	return &Dwrapper{
		ctx: ctx,

		client:    client,
		namespace: ns,
	}, nil
}

func (d *Dwrapper) Close() error {
	return d.client.Close()
}

func (d *Dwrapper) GetInitPid(cID string) (uint32, error) {
	taskClient := d.client.TaskService()
	c, err := taskClient.Get(d.ctx, &taskservice.GetRequest{
		ContainerID: cID,
	})
	if err != nil {
		return 0, err
	}
	return c.Process.Pid, nil
}

func (d *Dwrapper) GetRootDir(cID string) (string, error) {
	c, err := d.client.LoadContainer(d.ctx, cID)
	if err != nil {
		return "", err
	}
	s, err := c.Spec(d.ctx)
	if err != nil {
		return "", err
	}
	rootPath := s.Root.Path
	if !strings.HasPrefix(rootPath, "/") {
		return fmt.Sprintf("/run/containerd/io.containerd.runtime.v2.task/%s/%s/%s",
			d.namespace, cID, rootPath), nil
	}
	return rootPath, nil
}

func (d *Dwrapper) GetWorkingDir(cID string) (string, error) {
	c, err := d.client.LoadContainer(d.ctx, cID)
	if err != nil {
		return "", err
	}
	s, err := c.Spec(d.ctx)
	if err != nil {
		return "", err
	}
	return s.Process.Cwd, nil
}

func (d *Dwrapper) GetEnv(cID string) ([]string, error) {
	c, err := d.client.LoadContainer(d.ctx, cID)
	if err != nil {
		return nil, err
	}
	s, err := c.Spec(d.ctx)
	if err != nil {
		return nil, err
	}
	return s.Process.Env, nil
}
