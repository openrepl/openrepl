package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Language is a configuration for a programming language.
type Language struct {
	RunContainer  ContainerConfig `json:"run"`
	TermContainer ContainerConfig `json:"term"`
}

// ContainerServer is a server that runs containers
type ContainerServer struct {
	// SessionConfig is the ContainerSessionConfig to use in all ContainerSessions.
	SessionConfig ContainerSessionConfig

	// Containers is a map of language names to container names.
	Containers map[string]Language

	// Upgrader is a websocket Upgrader used for all websocket connections.
	Upgrader websocket.Upgrader
}

// HandleTerminal serves an interactive terminal websocket.
func (cs *ContainerServer) HandleTerminal(w http.ResponseWriter, r *http.Request) {
	// get language
	lang, ok := cs.Containers[r.URL.Query().Get("lang")]
	if !ok {
		http.Error(w, "language not supported", http.StatusBadRequest)
		return
	}

	// run ContainerSession
	HandleContainerSession(w, r, false, lang.TermContainer, &cs.SessionConfig)
}

// HandleRun serves an interactive terminal websocket running user code.
func (cs *ContainerServer) HandleRun(w http.ResponseWriter, r *http.Request) {
	// get language
	lang, ok := cs.Containers[r.URL.Query().Get("lang")]
	if !ok {
		http.Error(w, "language not supported", http.StatusBadRequest)
		return
	}

	// run ContainerSession
	HandleContainerSession(w, r, true, lang.RunContainer, &cs.SessionConfig)
}
