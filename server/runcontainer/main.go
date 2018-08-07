package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/client"
)

func main() {
	dcli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	srv := &ContainerServer{
		SessionConfig: ContainerSessionConfig{
			OutputBufferSize:     1024,
			ShutdownTimeout:      10 * time.Second,
			DockerClient:         dcli,
			ContainerStopTimeout: time.Minute,
			StartTimeout:         time.Minute,
			SessionTimeout:       time.Hour,
			PingRate:             30 * time.Second,
		},
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
