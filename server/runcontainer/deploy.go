package main

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// ContainerConfig is a container configuration.
type ContainerConfig struct {
	Image   string   `json:"image"`
	Command []string `json:"cmd"`
}

// Container is a running container.
type Container struct {
	clck         sync.Mutex
	closed       bool
	cli          *client.Client
	ID           string
	IO           io.ReadWriteCloser
	closetimeout time.Duration
}

func (c *Container) Write(dat []byte) (int, error) {
	return c.IO.Write(dat)
}

func (c *Container) Read(dat []byte) (int, error) {
	return c.IO.Read(dat)
}

// Close closes and removes the container.
func (c *Container) Close() error {
	// lock closed field
	c.clck.Lock()
	defer c.clck.Unlock()

	// dont reclose if already closed
	if c.closed {
		return nil
	}
	defer func() { c.closed = true }()

	// close websocket
	cerr := c.IO.Close()

	// remove container
	ctx, cancel := context.WithTimeout(context.Background(), c.closetimeout)
	defer cancel()
	rerr := c.cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
		Force: true,
	})

	// handle errors
	if rerr != nil {
		log.Printf("failed to remove container: %s", rerr.Error())
	}
	err := cerr
	if err != nil {
		err = rerr
	}
	return err
}

// Deploy deploys a container with this configuration.
func (cc ContainerConfig) Deploy(ctx context.Context, cli *client.Client, stoptimeout time.Duration, prestart func(context.Context, *Container) error) (cont *Container, err error) {
	// create container
	c, err := cli.ContainerCreate(ctx, &container.Config{
		Image:           cc.Image,
		Cmd:             cc.Command,
		Tty:             true,
		OpenStdin:       true,
		NetworkDisabled: true,
	}, &container.HostConfig{
		Resources: container.Resources{
			NanoCPUs: int64(time.Second/time.Nanosecond) / 2, // 1/2 CPU cap
			Memory:   1 << 27,                                // cap at 128MB
		},
	}, nil, "")
	if err != nil {
		return nil, err
	}

	// cleanup container on failed startup
	defer func() {
		if err != nil {
			delctx, cancel := context.WithTimeout(context.Background(), stoptimeout)
			defer cancel()
			rerr := cli.ContainerRemove(delctx, c.ID, types.ContainerRemoveOptions{
				Force: true,
			})
			if rerr != nil {
				log.Printf("failed to remove container: %s", rerr.Error())
			}
		}
	}()

	cont = &Container{
		cli:          cli,
		ID:           c.ID,
		closetimeout: stoptimeout,
	}

	// run prestart hook
	if prestart != nil {
		err = prestart(ctx, cont)
		if err != nil {
			return nil, err
		}
	}

	// attach to container
	resp, err := cli.ContainerAttach(ctx, c.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, err
	}

	// start container
	err = cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	// convert to websocket
	cont.IO = resp.Conn

	return cont, nil
}
