package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"

	"github.com/rs/cors"
)

func main() {
	var n uint64
	var lck sync.Mutex
	aln := func() uint64 {
		lck.Lock()
		defer lck.Unlock()
		defer func() { n++ }()
		return n
	}
	http.Handle("/save", cors.New(cors.Options{
		AllowedMethods:   []string{"POST"},
		AllowCredentials: true,
	}).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Form parse error", http.StatusBadRequest)
			return
		}
		srcid := r.FormValue("srcid")
		if srcid == "" {
			http.Error(w, "Source file ID missing", http.StatusBadRequest)
			return
		}
		g, err := http.Get((&url.URL{Scheme: "http", Host: "60s", Path: "/get", RawQuery: fmt.Sprintf("id=%s", url.QueryEscape(srcid))}).String())
		if err != nil {
			http.Error(w, "backend request error", http.StatusBadGateway)
			return
		}
		defer g.Body.Close()
		if g.StatusCode != http.StatusOK {
			http.Error(w, "backend request error "+g.Status, http.StatusBadGateway)
			return
		}
		id := aln()
		exist := func(p string) bool {
			_, err := os.Stat(p)
			return !os.IsNotExist(err)
		}
		for exist(fmt.Sprintf("/dat/%d", id)) {
			id = aln()
		}
		f, err := os.OpenFile(fmt.Sprintf("/dat/%d", id), os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			http.Error(w, "file open error", http.StatusBadGateway)
			log.Printf("File open error: %s", err.Error())
			return
		}
		defer f.Close()
		io.Copy(f, g.Body)
		fmt.Fprint(w, id)
	})))
	http.Handle("/get", cors.New(cors.Options{
		AllowedMethods:   []string{"GET"},
		AllowCredentials: true,
	}).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Form parse error", http.StatusBadRequest)
			return
		}
		ids := r.FormValue("id")
		if ids == "" {
			http.Error(w, "ID missing", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseUint(ids, 10, 64)
		if err != nil {
			http.Error(w, "ID is not a valid base-10 integer", http.StatusBadRequest)
			return
		}
		fname := fmt.Sprintf("/dat/%d", id)
		_, err = os.Stat(fname)
		if os.IsNotExist(err) {
			http.Error(w, "Save with specified ID does not exist", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "Error statting save file", http.StatusBadRequest)
			log.Printf("Stat error: %s\n", err.Error())
			return
		}
		f, err := os.Open(fname)
		if err != nil {
			http.Error(w, "Error opening save file", http.StatusBadRequest)
			log.Printf("Open error: %s\n", err.Error())
			return
		}
		defer f.Close()
		io.Copy(w, f)
	})))
	http.ListenAndServe(":80", nil)
}
