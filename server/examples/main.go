package main

import (
	"flag"
	"net/http"
)

func main() {
	var esdir string
	var serve string
	flag.StringVar(&esdir, "examples", "/examples", "dir containing examples")
	flag.StringVar(&serve, "http", ":80", "http server address")
	flag.Parse()

	es, err := LoadExampleSet(esdir)
	if err != nil {
		panic(err)
	}

	http.Handle("/query", es)

	panic(http.ListenAndServe(serve, nil))
}
