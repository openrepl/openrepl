package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	store := make(map[uint64][]byte)
	var n uint64
	var lck sync.Mutex
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Not a put", http.StatusBadRequest)
			return
		}
		buf := bytes.NewBuffer(nil)
		gz := gzip.NewWriter(buf)
		_, err := io.Copy(gz, r.Body)
		if err != nil {
			return
		}
		err = gz.Flush()
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
	})
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
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
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			buf := bytes.NewBuffer(dat)
			gzr, err := gzip.NewReader(buf)
			if err != nil {
				return
			}
			io.Copy(w, gzr)
		} else {
			io.Copy(w, bytes.NewBuffer(dat))
		}
	})
	http.ListenAndServe(":80", nil)
}
