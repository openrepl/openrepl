package main

import (
	"archive/tar"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/jadr2ddude/websocket"
)

func main() {
	dcli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	srv := &ContainerServer{
		DockerClient: dcli,
	}
	f, err := os.Open("langs.json")
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(f).Decode(&srv.Containers)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/term", srv.HandleTerminal)
	http.HandleFunc("/run", srv.HandleRun)
	panic(http.ListenAndServe(":80", nil))
}

// ContainerConfig is a container configuration.
type ContainerConfig struct {
	Image   string   `json:"image"`
	Command []string `json:"cmd"`
}

// pullImg pulls the docker image used by the ContainerConfig.
func (cc ContainerConfig) pullImg(ctx context.Context, cli *client.Client) (err error) {
	pr, err := cli.ImagePull(ctx, "docker.io/library/"+cc.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer func() {
		cerr := pr.Close()
		if cerr != nil && err == nil {
			err = cerr
		}
	}()
	_, err = io.Copy(os.Stdout, pr)
	if err != nil {
		return err
	}
	return nil
}

// Container is a running container.
type Container struct {
	cli *client.Client
	ID  string
	IO  io.ReadWriteCloser
}

// Close closes and removes the container.
func (c *Container) Close(ctx context.Context) error {
	// close websocket
	cerr := c.IO.Close()

	// remove container
	rerr := c.cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
		Force: true,
	})

	// handle errors
	if rerr != nil {
		log.Printf("Failed to remove container: %s", rerr.Error())
	}
	err := cerr
	if err != nil {
		err = rerr
	}
	return err
}

// Deploy deploys a container with this configuration.
func (cc ContainerConfig) Deploy(ctx context.Context, cli *client.Client, prestart func(*Container) error) (cont *Container, err error) {
	/*
		    // pull image
			err = cc.pullImg(ctx, cli)
			if err != nil {
				return nil, err
			}
	*/

	// create container
	c, err := cli.ContainerCreate(ctx, &container.Config{
		Image:           cc.Image,
		Cmd:             cc.Command,
		Tty:             true,
		OpenStdin:       true,
		NetworkDisabled: true,
	}, nil, nil, "")
	if err != nil {
		return nil, err
	}

	// cleanup container on failed startup
	defer func() {
		if err != nil {
			delctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			rerr := cli.ContainerRemove(delctx, c.ID, types.ContainerRemoveOptions{
				Force: true,
			})
			if rerr != nil {
				log.Printf("Failed to remove container: %s", rerr.Error())
			}
		}
	}()

	cont = &Container{
		cli: cli,
		ID:  c.ID,
	}

	// run prestart hook
	if prestart != nil {
		err = prestart(cont)
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

// Language is a configuration for a programming language.
type Language struct {
	RunContainer  ContainerConfig `json:"run"`
	TermContainer ContainerConfig `json:"term"`
}

// ContainerServer is a server that runs containers
type ContainerServer struct {
	// DockerClient is the client to Docker.
	DockerClient *client.Client

	// Containers is a map of language names to container names.
	Containers map[string]Language

	// Upgrader is a websocket Upgrader used for all websocket connections.
	Upgrader websocket.Upgrader
}

// StatusUpdate is a status message which can be sent to the client.
type StatusUpdate struct {
	Status string `json:"status"`
	Error  string `json:"err,omitempty"`
}

func copyWebSocket(dst *websocket.Conn, src *websocket.Conn, cancel context.CancelFunc, n int) {
	defer cancel()
	for {
		// read message
		t, dat, err := src.ReadMessage()
		if err != nil {
			log.Println("Stopping transfer", err.Error(), n)
			return
		}
		log.Println(t, dat)

		switch t {
		case websocket.CloseMessage:
			// shut down
			return
		case websocket.BinaryMessage:
			// forward message
			err = dst.WriteMessage(t, dat)
			if err != nil {
				return
			}
		}
	}
}

// HandleTerminal serves an interactive terminal websocket.
func (cs *ContainerServer) HandleTerminal(w http.ResponseWriter, r *http.Request) {
	log.Println("term")
	// get language
	lang, ok := cs.Containers[r.URL.Query().Get("lang")]
	if !ok {
		http.Error(w, "language not supported", http.StatusBadRequest)
		return
	}

	// upgrade websocket
	log.Println("upgrade")
	conn, err := cs.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// send status "starting"
	log.Println("starting")
	err = conn.WriteJSON(StatusUpdate{Status: "starting"})
	if err != nil {
		return
	}

	// deploy container with 1 min timeout
	startctx, startcancel := context.WithTimeout(context.Background(), time.Minute)
	defer startcancel()
	c, err := lang.TermContainer.Deploy(startctx, cs.DockerClient, nil)
	log.Println("deploy", c, err)
	if err != nil {
		conn.WriteJSON(StatusUpdate{Status: "error", Error: err.Error()})
		err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err == nil {
			donech := make(chan struct{})
			go func() {
				defer close(donech)
				// drain client messages and wait for disconnect
				var e error
				for e == nil {
					_, _, e = conn.ReadMessage()
				}
			}()
			timer := time.NewTimer(10 * time.Second)
			defer timer.Stop()
			select {
			case <-donech:
			case <-timer.C:
			}
		}
	}
	defer func() {
		stopctx, stopcancel := context.WithTimeout(context.Background(), time.Minute)
		defer stopcancel()
		cerr := c.Close(stopctx)
		if cerr != nil {
			log.Printf("Failed to stop container %q: %s", c.ID, cerr)
		}
	}()

	// update status to running
	log.Println("running")
	err = conn.WriteJSON(StatusUpdate{Status: "running"})
	if err != nil {
		return
	}

	// bridge connections
	runctx, cancel := context.WithCancel(context.Background())
	go func() { // output
		defer cancel()
		buf := make([]byte, 1024)
		for {
			n, err := c.IO.Read(buf)
			log.Println(n, err)
			if err != nil {
				return
			}

			log.Println(hex.Dump(buf[:n]))

			err = conn.WriteMessage(websocket.TextMessage, buf[:n])
			if err != nil {
				return
			}
		}
	}()
	go func() { // input
		defer cancel()
		for {
			t, r, err := conn.NextReader()
			if err != nil {
				return
			}
			switch t {
			case websocket.TextMessage:
				fallthrough
			case websocket.BinaryMessage:
				_, err = io.Copy(c.IO, r)
				if err != nil {
					return
				}
			case websocket.CloseMessage:
				return
			}
		}
	}()
	//go copyWebSocket(c.IO, conn, cancel, 2)
	log.Println("Copying")
	<-runctx.Done()
	log.Println("Done")
}

func packCodeTarball(dat []byte) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		var err error
		defer func() {
			if err == nil {
				w.Close()
			} else {
				w.CloseWithError(err)
			}
		}()
		tw := tar.NewWriter(w)
		defer tw.Close()

		err = tw.WriteHeader(&tar.Header{
			Name: "code",
			Mode: 0444,
			Size: int64(len(dat)),
		})
		if err != nil {
			return
		}

		_, err = tw.Write(dat)
		if err != nil {
			return
		}
	}()
	return r
}

// HandleRun serves an interactive terminal running user code over a websocket.
func (cs *ContainerServer) HandleRun(w http.ResponseWriter, r *http.Request) {
	// get language
	lang, ok := cs.Containers[r.URL.Query().Get("lang")]
	if !ok {
		http.Error(w, "language not supported", http.StatusBadRequest)
		return
	}

	// upgrade websocket
	conn, err := cs.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// send status "creating"
	err = conn.WriteJSON(StatusUpdate{Status: "creating"})
	if err != nil {
		return
	}

	// deploy container with 1 min timeout
	startctx, startcancel := context.WithTimeout(context.Background(), time.Minute)
	defer startcancel()
	c, err := lang.RunContainer.Deploy(startctx, cs.DockerClient, func(c *Container) error {
		// update status to ready
		err := conn.WriteJSON(StatusUpdate{Status: "ready"})
		if err != nil {
			return err
		}

		// accept user code
		t, dat, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if t != websocket.BinaryMessage && t != websocket.TextMessage {
			log.Println("Client sent invalid message type")
			return err
		}

		// update status to uploading
		err = conn.WriteJSON(StatusUpdate{Status: "uploading"})
		if err != nil {
			return err
		}

		// send code to Docker
		tr := packCodeTarball(dat)
		err = c.cli.CopyToContainer(startctx, c.ID, "/", tr, types.CopyToContainerOptions{})
		tr.Close()
		if err != nil {
			conn.WriteJSON(StatusUpdate{Status: "error", Error: err.Error()})
			return err
		}

		// update status to starting
		err = conn.WriteJSON(StatusUpdate{Status: "starting"})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		conn.WriteJSON(StatusUpdate{Status: "error", Error: err.Error()})
		err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err == nil {
			donech := make(chan struct{})
			go func() {
				defer close(donech)
				// drain client messages and wait for disconnect
				var e error
				for e == nil {
					_, _, e = conn.ReadMessage()
				}
			}()
			timer := time.NewTimer(10 * time.Second)
			defer timer.Stop()
			select {
			case <-donech:
			case <-timer.C:
			}
		}
	}
	defer func() {
		stopctx, stopcancel := context.WithTimeout(context.Background(), time.Minute)
		defer stopcancel()
		cerr := c.Close(stopctx)
		if cerr != nil {
			log.Printf("Failed to stop container %q: %s", c.ID, cerr)
		}
	}()

	// update status to running
	err = conn.WriteJSON(StatusUpdate{Status: "running"})
	if err != nil {
		return
	}

	// bridge connections
	runctx, cancel := context.WithCancel(context.Background())
	go func() { // output
		defer cancel()
		buf := make([]byte, 1024)
		for {
			n, err := c.IO.Read(buf)
			log.Println(n, err)
			if err != nil {
				return
			}

			err = conn.WriteMessage(websocket.TextMessage, buf[:n])
			if err != nil {
				return
			}
		}
	}()
	go func() { // input
		defer cancel()
		for {
			t, r, err := conn.NextReader()
			if err != nil {
				return
			}
			switch t {
			case websocket.TextMessage:
				fallthrough
			case websocket.BinaryMessage:
				_, err = io.Copy(c.IO, r)
				if err != nil {
					return
				}
			case websocket.CloseMessage:
				return
			}
		}
	}()
	log.Println("Copying")
	<-runctx.Done()
	log.Println("Done")
}
