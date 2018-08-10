package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
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
	http.HandleFunc("/highlight", HandleHighlight)
	http.HandleFunc("/highlight.css", HandleCSS)

	panic(http.ListenAndServe(serve, nil))
}

// Code is a struct containing code with metadata.
type Code struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

var form = html.New(html.TabWidth(8), html.ClassPrefix("h-"), html.WithClasses())
var style *chroma.Style
var styleCSS []byte
var styleETag string

func init() {
	style = styles.Get("swapoff")
	if style == nil {
		style = styles.Fallback
	}

	buf := bytes.NewBuffer(nil)
	err := form.WriteCSS(buf, style)
	if err != nil {
		panic(err)
	}
	styleCSS = buf.Bytes()
	hash := sha256.Sum256(styleCSS)
	styleETag = hex.EncodeToString(hash[:])
}

// HandleHighlight handles /highlight requests.
func HandleHighlight(w http.ResponseWriter, r *http.Request) {
	// only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// parse input
	var c Code
	err := json.NewDecoder(io.LimitReader(r.Body, 1024*1024)).Decode(&c)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load body: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// fix python naming for chroma
	if c.Language == "python2" {
		c.Language = "python"
	}

	// prepare syntax highlighter
	lexer := lexers.Get(c.Language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// highlight
	iter, err := lexer.Tokenise(nil, c.Code)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to tokenize: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	err = form.Format(w, style, iter)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to format: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// HandleCSS handles a request for the example highlight stylesheet.
func HandleCSS(w http.ResponseWriter, r *http.Request) {
	// ETag caching
	w.Header().Set("Etag", styleETag)
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, styleETag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// send data
	w.Write(styleCSS)
}
