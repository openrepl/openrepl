package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/yhat/wsutil"
)

func main() {
	langs := strings.Split("lua python", " ")
	rp := func(str string) {
		http.HandleFunc("/"+str, func(w http.ResponseWriter, r *http.Request) {
			g, err := http.Get((&url.URL{Scheme: "http", Host: "localhost:65", Path: str}).String())
			if err != nil {
				log.Println(err.Error())
				return
			}
			rh := w.Header()
			for n, h := range g.Header {
				for _, v := range h {
					rh.Add(n, v)
				}
			}
			io.Copy(w, g.Body)
		})
	}
	for _, la := range langs {
		l := la
		http.HandleFunc("/"+l, func(w http.ResponseWriter, r *http.Request) {
			args := r.URL.Query()["arg"]
			if len(args) < 1 || args[0] != l {
				r.URL.RawQuery = "arg=" + l
				http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
				return
			}
			g, err := http.Get((&url.URL{Scheme: "http", Host: "localhost:65", Path: "/", RawQuery: "arg=" + l}).String())
			if err != nil {
				return
			}
			rh := w.Header()
			for n, h := range g.Header {
				for _, v := range h {
					rh.Add(n, v)
				}
			}
			io.Copy(w, g.Body)
		})
		http.HandleFunc(fmt.Sprintf("/%sws", l),
			func(w http.ResponseWriter, r *http.Request) {
				r.URL.Path = "/ws"
				wsutil.NewSingleHostReverseProxy(&url.URL{
					Scheme: "ws",
					Host:   "localhost:65",
					Path:   "",
				}).ServeHTTP(w, r)
			},
		)
	}
	rp("css/xterm.css")
	rp("css/xterm_customize.css")
	rp("css/index.css")
	rp("js/gotty-bundle.js")
	rp("js/auth_token.js")
	rp("js/config.js")
	rp("auth_token.js")
	rp("config.js")
	gtty := exec.Command("gotty", "-p", "65", "--permit-arguments", "--close-signal", "1", "-w", "bash", "--", "sess.sh")
	err := gtty.Start()
	if err != nil {
		panic(err)
	}
	defer gtty.Process.Kill()
	http.ListenAndServe(":52", nil)
}
