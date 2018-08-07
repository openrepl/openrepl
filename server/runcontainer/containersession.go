package main

import (
	"archive/tar"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

// ContainerSessionConfig is a configuration for a ContainerSession,
type ContainerSessionConfig struct {
	// OutputBufferSize is the size of the buffer to read output into.
	OutputBufferSize int

	// ShutdownTimeout is the timeout for shutting down a websocket.
	ShutdownTimeout time.Duration

	// DockerClient is the docker client to use to create containers.
	DockerClient *client.Client

	// PingRate is the amount of time to wait between sending pings.
	PingRate time.Duration

	// ContainerStopTimeout is the timeout for stopping a container.
	ContainerStopTimeout time.Duration

	// StartTimeout is the timeout for starting if using HandleContainerSession.
	StartTimeout time.Duration

	// SessionTimeout is the timeout for the session if using HandleContainerSession.
	SessionTimeout time.Duration

	// Upgrader is the websocket upgrader to use if using HandleContainerSession.
	Upgrader websocket.Upgrader
}

// ContainerSession is a terminal session with a container over a websocket.
type ContainerSession struct {
	// Container is the container being controlled.
	Container io.ReadWriteCloser

	// Client is the client websocket connection.
	Client *websocket.Conn

	// Config is the configuration of the ContainerSession.
	Config *ContainerSessionConfig

	// IsRun is whether the session is running pre-written code.
	IsRun bool

	// ContainerConfig is the ContainerConfig to be used to create the container.
	// Only necessary when using CreateContainer.
	ContainerConfig ContainerConfig
}

// Close closes the ContainerSession.
func (cs *ContainerSession) Close() {
	// shut down container
	if cs.Container != nil {
		cs.Container.Close()
	}

	// attempt to gracefully shutdown websocket
	cerr := cs.Client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if cerr == nil {
		donech := make(chan struct{})
		go func() {
			defer close(donech)
			// drain client messages and wait for disconnect
			var e error
			for e == nil {
				_, _, e = cs.Client.ReadMessage()
			}
		}()
		timer := time.NewTimer(cs.Config.ShutdownTimeout)
		defer timer.Stop()
		select {
		case <-donech:
		case <-timer.C:
		}
	}

	// close websocket
	cs.Client.Close()
}

// runOutput copies output from the container to the client.
func (cs *ContainerSession) runOutput(errch chan<- error) {
	var err error
	defer func() { errch <- err }()
	buf := make([]byte, cs.Config.OutputBufferSize)
	for err == nil {
		var n int

		// run read
		n, err = cs.Container.Read(buf)
		if err != nil {
			return
		}

		// send data to client
		err = cs.Client.WriteMessage(websocket.TextMessage, buf[:n])
		if err != nil {
			return
		}
	}
}

// runInput copies input from the client to the container.
func (cs *ContainerSession) runInput(errch chan<- error) {
	var err error
	defer func() { errch <- err }()
	for err == nil {
		var t int
		var r io.Reader

		// get next websocket message reader
		t, r, err = cs.Client.NextReader()
		if err != nil {
			return
		}

		// handle close sent by client
		if t == websocket.CloseMessage {
			io.Copy(ioutil.Discard, r)
			return
		}

		// copy to container
		_, err = io.Copy(cs.Container, r)
		if err != nil {
			return
		}
	}
}

func (cs *ContainerSession) runPing(errch chan<- error) {
	// record pong messages
	pongch := make(chan struct{}, 1)
	cs.Client.SetPongHandler(func(appData string) error {
		select {
		case pongch <- struct{}{}:
		}
		return nil
	})

	// start playing ping-pong
	go func() {
		var err error
		defer func() { errch <- err }()
		tick := time.NewTicker(cs.Config.PingRate)
		defer tick.Stop()
		for range tick.C {
			// send ping
			err = cs.Client.WriteControl(websocket.PingMessage, []byte{1}, time.Now().Add(10*time.Second))
			if err != nil {
				return
			}

			// wait for pong
			select {
			case <-pongch:
				// we are good - client sent pong on time
			case <-tick.C:
				// timeout while waiting for pong - stalled client
				err = errors.New("stalled client")
				return
			}
		}
	}()
}

// RunIO runs input and output for the session, closing afterwards.
func (cs *ContainerSession) RunIO(ctx context.Context) error {
	errch := make(chan error, 2)

	// start output
	go cs.runOutput(errch)

	// start input
	go cs.runInput(errch)

	// start ping-pong
	cs.runPing(errch)

	// wait for error
	err := <-errch

	// close session
	cs.Close()

	// ignore second/third error
	<-errch
	<-errch

	return err
}

// StatusUpdate is a status message which can be sent to the client.
type StatusUpdate struct {
	Status string `json:"status"`
	Error  string `json:"err,omitempty"`
}

// UpdateStatus sends a StatusUpdate to the client.
func (cs *ContainerSession) UpdateStatus(status StatusUpdate) error {
	return cs.Client.WriteJSON(status)
}

// packCodeTarball generates a tarball containing dat as a file called "code".
func packCodeTarball(dat []byte) io.ReadCloser {
	// create pipe
	r, w := io.Pipe()
	go func() {
		// handle closing, passing any error to the reader
		var err error
		defer func() {
			if err == nil {
				w.Close()
			} else {
				w.CloseWithError(err)
			}
		}()

		// prepare tar for writing
		tw := tar.NewWriter(w)
		defer func() {
			cerr := tw.Close()
			if cerr != nil && err == nil {
				err = cerr
			}
		}()

		// write tar header
		err = tw.WriteHeader(&tar.Header{
			Name: "code",
			Mode: 0444,
			Size: int64(len(dat)),
		})
		if err != nil {
			return
		}

		// add file to tarball
		_, err = tw.Write(dat)
		if err != nil {
			return
		}
	}()
	return r
}

// sendCode sends client code to the container.
func (cs *ContainerSession) sendCode(ctx context.Context, c *Container) error {
	// update status to ready
	err := cs.UpdateStatus(StatusUpdate{Status: "ready"})
	if err != nil {
		return err
	}

	// accept user code
	t, dat, err := cs.Client.ReadMessage()
	if err != nil {
		return err
	}
	if t != websocket.BinaryMessage && t != websocket.TextMessage {
		return err
	}

	// update status to uploading
	err = cs.UpdateStatus(StatusUpdate{Status: "uploading"})
	if err != nil {
		return err
	}

	// send code to Docker
	tr := packCodeTarball(dat)
	err = c.cli.CopyToContainer(ctx, c.ID, "/", tr, types.CopyToContainerOptions{})
	tr.Close()
	if err != nil {
		cs.UpdateStatus(StatusUpdate{Status: "error", Error: err.Error()})
		return err
	}

	// update status to starting
	err = cs.UpdateStatus(StatusUpdate{Status: "starting"})
	if err != nil {
		return err
	}

	return nil
}

// CreateContainer creates and starts a container.
func (cs *ContainerSession) CreateContainer(ctx context.Context) error {
	// select prestart hook
	var prestart func(context.Context, *Container) error
	if cs.IsRun {
		prestart = cs.sendCode
	}

	// deploy container
	c, err := cs.ContainerConfig.Deploy(ctx, cs.Config.DockerClient, cs.Config.ContainerStopTimeout, prestart)
	if err != nil {
		return err
	}

	// save container for I/O
	cs.Container = c

	return nil
}

// HandleContainerSession processes a container session.
func HandleContainerSession(w http.ResponseWriter, r *http.Request, isrun bool, cc ContainerConfig, sc *ContainerSessionConfig) {
	// upgrade websocket connection
	ws, err := sc.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade: %s", err.Error())
		return
	}

	// create ContainerSession
	cs := &ContainerSession{
		Client:          ws,
		Config:          sc,
		IsRun:           isrun,
		ContainerConfig: cc,
	}
	defer cs.Close()

	// set status to "starting"
	err = cs.UpdateStatus(StatusUpdate{Status: "starting"})
	if err != nil {
		return
	}

	// start container
	startctx, scancel := context.WithTimeout(context.Background(), sc.StartTimeout)
	defer scancel()
	err = cs.CreateContainer(startctx)
	if err != nil {
		cs.UpdateStatus(StatusUpdate{Status: "error", Error: err.Error()})
		log.Printf("failed to start: %s", err.Error())
		return
	}

	// set status to "running"
	err = cs.UpdateStatus(StatusUpdate{Status: "running"})
	if err != nil {
		return
	}

	// run session IO
	sessctx, cancel := context.WithTimeout(context.Background(), sc.SessionTimeout)
	defer cancel()
	err = cs.RunIO(sessctx)
	if err != nil {
		log.Printf("I/O stopped with error: %s", err.Error())
	}
}
