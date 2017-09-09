package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rs/cors"
)

func main() {
	store := make(map[uint64][]byte)
	var n uint64
	var lck sync.Mutex
	http.Handle("/add", cors.New(cors.Options{
		AllowedMethods:   []string{"PUT"},
		AllowCredentials: true,
	}).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Not a put", http.StatusBadRequest)
			log.Printf("Incorrect method %s\n", r.Method)
			return
		}
		buf := bytes.NewBuffer(nil)
		_, err := io.Copy(buf, r.Body)
		if err != nil {
			return
		}
		lck.Lock()
		defer lck.Unlock()
		store[n] = buf.Bytes()
		go func() {
			time.Sleep(time.Minute)
			lck.Lock()
			defer lck.Unlock()
			delete(store, n)
		}()
		w.Write([]byte(fmt.Sprint(n)))
		n++
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
			http.Error(w, "No id in request", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseUint(ids, 10, 64)
		if err != nil {
			http.Error(w, "Id is not an int", http.StatusBadRequest)
			return
		}
		dat := func() []byte {
			lck.Lock()
			defer lck.Unlock()
			return store[id]
		}()
		if dat == nil {
			http.Error(w, "Data gone", http.StatusGone)
			return
		}
		io.Copy(w, bytes.NewBuffer(dat))
	})))
	http.ListenAndServe(":80", nil)
}
