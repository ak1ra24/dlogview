package api

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Container Operation struct
type Container struct {
	Ctx context.Context
	Cli *client.Client
}

// NewContainer func is Create Container Client
func NewContainer(ctx context.Context, cli *client.Client) *Container {
	container := &Container{
		Ctx: ctx,
		Cli: cli,
	}

	return container
}

// ListContainer func is docker ps -a
func (c *Container) ListContainer() ([]types.Container, error) {
	containers, err := c.Cli.ContainerList(c.Ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
		return nil, err
	}

	return containers, nil
}

// DisplayLog func is docker log
func (c *Container) DisplayLog(containerID string) error {
	r, err := c.Cli.ContainerLogs(c.Ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		log.Fatal(err)
		return err
	}

	io.Copy(os.Stdout, r)

	return nil
}
